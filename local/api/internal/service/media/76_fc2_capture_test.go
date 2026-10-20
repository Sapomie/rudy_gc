package media

import (
	"context"
	"errors"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

const testFC2CaptureKey = "FC2PPV-1234567-1"

type fc2MediaImageModelStub struct {
	inserted  *moviex.FC2MediaImage
	row       *moviex.FC2MediaImage
	rows      []*moviex.FC2MediaImage
	deleteErr error
	deleted   bool
}

func (m *fc2MediaImageModelStub) Insert(_ context.Context, data *moviex.FC2MediaImage) (*moviex.FC2MediaImage, error) {
	row := *data
	row.ID = 1
	m.inserted = &row
	m.row = &row
	return &row, nil
}

func (m *fc2MediaImageModelStub) FindOneByMediaIDAndID(_ context.Context, mediaID, id int64) (*moviex.FC2MediaImage, error) {
	if m.row == nil || m.row.FC2MediaID != mediaID || m.row.ID != id {
		return nil, moviex.ErrNotFound
	}
	row := *m.row
	return &row, nil
}

func (m *fc2MediaImageModelStub) ListByMediaID(context.Context, int64) ([]*moviex.FC2MediaImage, error) {
	return m.rows, nil
}

type fc2MovieModelStub struct {
	movie        *moviex.FC2Movie
	coverUpdates []int64
	updateErr    error
}

func (*fc2MovieModelStub) Upsert(context.Context, *moviex.FC2Movie) (*moviex.FC2Movie, error) {
	return nil, nil
}

func (m *fc2MovieModelStub) FindOne(_ context.Context, id int64) (*moviex.FC2Movie, error) {
	if m.movie == nil || m.movie.ID != id {
		return nil, moviex.ErrNotFound
	}
	row := *m.movie
	return &row, nil
}

func (m *fc2MovieModelStub) FindOneByKey(context.Context, string) (*moviex.FC2Movie, error) {
	return m.movie, nil
}

func (*fc2MovieModelStub) ListByIDs(context.Context, []int64) ([]*moviex.FC2Movie, error) {
	return []*moviex.FC2Movie{}, nil
}

func (*fc2MovieModelStub) ListByCode(context.Context, string) ([]*moviex.FC2Movie, error) {
	return []*moviex.FC2Movie{}, nil
}

func (*fc2MovieModelStub) ListPage(context.Context, int64, int64, string, string, string, string) ([]*moviex.FC2Movie, error) {
	return []*moviex.FC2Movie{}, nil
}

func (*fc2MovieModelStub) CountAll(context.Context, string, string) (int64, error) {
	return 0, nil
}

func (*fc2MovieModelStub) ListCards(context.Context, int64, int64, bool, bool, string, string) ([]*moviex.FC2Movie, error) {
	return []*moviex.FC2Movie{}, nil
}

func (*fc2MovieModelStub) CountCards(context.Context, bool) (int64, error) {
	return 0, nil
}

func (*fc2MovieModelStub) UpdateMediaSnapshot(context.Context, int64, int64, int64, int64) error {
	return nil
}

func (m *fc2MovieModelStub) UpdateCoverImage(_ context.Context, id, coverImageID, _ int64) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.movie == nil || m.movie.ID != id {
		return moviex.ErrNotFound
	}
	m.coverUpdates = append(m.coverUpdates, coverImageID)
	m.movie.CoverImageID = coverImageID
	return nil
}

func (*fc2MediaImageModelStub) ListByMediaIDs(context.Context, []int64) ([]*moviex.FC2MediaImage, error) {
	return []*moviex.FC2MediaImage{}, nil
}

func (m *fc2MediaImageModelStub) DeleteByMediaIDAndID(_ context.Context, mediaID, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if m.row == nil || m.row.FC2MediaID != mediaID || m.row.ID != id {
		return moviex.ErrNotFound
	}
	m.deleted = true
	return nil
}

func TestBuildFC2ContactSheetTimestamps(t *testing.T) {
	got := buildFC2ContactSheetTimestamps(100, fc2DefaultCoverColumns*fc2DefaultCoverRows)
	want := []float64{5, 15, 25, 35, 45, 55, 65, 75, 85, 95}
	if len(got) != len(want) {
		t.Fatalf("时间点数量 = %d，期望 %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("时间点[%d] = %f，期望 %f", index, got[index], want[index])
		}
	}
}

func TestValidateFC2DefaultCoverMedia(t *testing.T) {
	mediaRow := &moviex.FC2Media{ID: 1, Duration: 120}
	if err := validateFC2DefaultCoverMedia(mediaRow); err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	err := validateFC2DefaultCoverMedia(&moviex.FC2Media{ID: 1})
	if !errors.Is(err, ErrInvalidFC2CaptureRequest) {
		t.Fatalf("无效媒体错误 = %v", err)
	}
}

func TestFC2CaptureDirNameUsesNormalizedFC2Key(t *testing.T) {
	got, err := fc2CaptureDirName(&moviex.FC2Media{FC2Key: "fc2ppv-1234567-1"})
	if err != nil {
		t.Fatalf("生成图片目录名失败: %v", err)
	}
	if got != testFC2CaptureKey {
		t.Fatalf("图片目录名 = %q，期望 %q", got, testFC2CaptureKey)
	}
}

func TestFC2CaptureDirNameRejectsInvalidFC2Key(t *testing.T) {
	for _, mediaRow := range []*moviex.FC2Media{nil, {}, {FC2Key: "../7"}} {
		if _, err := fc2CaptureDirName(mediaRow); !errors.Is(err, ErrInvalidFC2CaptureRequest) {
			t.Fatalf("无效 fc2_key 错误 = %v", err)
		}
	}
}

func TestPrepareFC2CaptureUsesConfiguredLocalImageDir(t *testing.T) {
	mediaRoot := t.TempDir()
	localImageDir := t.TempDir()
	sourceDir := filepath.Join(mediaRoot, "media")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("创建媒体目录失败: %v", err)
	}
	const fileName = "test.mp4"
	sourcePath := filepath.Join(sourceDir, fileName)
	if err := os.WriteFile(sourcePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("创建媒体文件失败: %v", err)
	}

	service := &Service{deps: &dep.Dep{Config: config.Config{
		FC2Media: config.FC2MediaConfig{
			RootDirs:      []string{mediaRoot},
			LocalImageDir: localImageDir,
		},
	}}}
	gotSourcePath, gotCaptureDir, err := service.prepareFC2Capture(&moviex.FC2Media{
		ID:       7,
		FC2Key:   testFC2CaptureKey,
		RootDir:  mediaRoot,
		FullDir:  sourceDir,
		FileName: fileName,
	})
	if err != nil {
		t.Fatalf("准备图片目录失败: %v", err)
	}
	if gotSourcePath != sourcePath {
		t.Fatalf("媒体路径 = %q，期望 %q", gotSourcePath, sourcePath)
	}
	wantCaptureDir := filepath.Join(localImageDir, testFC2CaptureKey)
	if gotCaptureDir != wantCaptureDir {
		t.Fatalf("图片目录 = %q，期望 %q", gotCaptureDir, wantCaptureDir)
	}
	if stat, statErr := os.Stat(gotCaptureDir); statErr != nil || !stat.IsDir() {
		t.Fatalf("图片目录未创建: stat=%v err=%v", stat, statErr)
	}
}

func TestGenerateFC2DefaultCoverUsesFixedLayoutAndSetsCover(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg 未安装")
	}
	mediaRoot := t.TempDir()
	localImageDir := t.TempDir()
	sourceDir := filepath.Join(mediaRoot, "media")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("创建媒体目录失败: %v", err)
	}
	const fileName = "test.mp4"
	sourcePath := filepath.Join(sourceDir, fileName)
	command := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-nostdin", "-y", "-f", "lavfi", "-i", "color=c=black:s=320x180:d=2", "-c:v", "mpeg4", sourcePath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("创建测试视频失败: %v: %s", err, output)
	}

	imageModel := &fc2MediaImageModelStub{}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		Config: config.Config{FC2Media: config.FC2MediaConfig{
			RootDirs:      []string{mediaRoot},
			LocalImageDir: localImageDir,
		}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}
	imageRow, err := service.GenerateFC2DefaultCover(context.Background(), &moviex.FC2Media{
		ID:         7,
		FC2MovieID: 28,
		FC2Key:     testFC2CaptureKey,
		RootDir:    mediaRoot,
		FullDir:    sourceDir,
		FileName:   fileName,
		Duration:   2,
	})
	if err != nil {
		t.Fatalf("生成默认联系表封面失败: %v", err)
	}
	wantDir := filepath.Join(localImageDir, testFC2CaptureKey)
	if filepath.Dir(imageRow.FilePath) != wantDir {
		t.Fatalf("封面目录 = %q，期望 %q", filepath.Dir(imageRow.FilePath), wantDir)
	}
	if imageModel.inserted == nil || imageModel.inserted.FilePath != imageRow.FilePath || imageModel.inserted.ImageType != FC2MediaImageTypeContactSheet {
		t.Fatalf("图片记录未写入正确路径: %#v", imageModel.inserted)
	}
	file, err := os.Open(imageRow.FilePath)
	if err != nil {
		t.Fatalf("打开封面失败: %v", err)
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		t.Fatalf("读取封面尺寸失败: %v", err)
	}
	if config.Width != fc2DefaultCoverColumns*fc2DefaultCoverCellWidth {
		t.Fatalf("封面宽度 = %d，期望 %d", config.Width, fc2DefaultCoverColumns*fc2DefaultCoverCellWidth)
	}
	if len(movieModel.coverUpdates) != 1 || movieModel.coverUpdates[0] != imageRow.ID {
		t.Fatalf("默认联系表未自动设为封面: %#v", movieModel.coverUpdates)
	}
}

func TestGenerateFC2DetailCoverUsesFiveByFiveLayoutWithoutChangingCardCover(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg 未安装")
	}
	mediaRoot := t.TempDir()
	localImageDir := t.TempDir()
	sourceDir := filepath.Join(mediaRoot, "media")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("创建媒体目录失败: %v", err)
	}
	const fileName = "test.mp4"
	sourcePath := filepath.Join(sourceDir, fileName)
	command := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-nostdin", "-y", "-f", "lavfi", "-i", "color=c=black:s=320x180:d=2", "-c:v", "mpeg4", sourcePath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("创建测试视频失败: %v: %s", err, output)
	}

	imageModel := &fc2MediaImageModelStub{}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28, CoverImageID: 99}}
	service := &Service{deps: &dep.Dep{
		Config: config.Config{FC2Media: config.FC2MediaConfig{
			RootDirs:      []string{mediaRoot},
			LocalImageDir: localImageDir,
		}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}
	imageRow, err := service.GenerateFC2DetailCover(context.Background(), &moviex.FC2Media{
		ID:         7,
		FC2MovieID: 28,
		FC2Key:     testFC2CaptureKey,
		RootDir:    mediaRoot,
		FullDir:    sourceDir,
		FileName:   fileName,
		Duration:   2,
	})
	if err != nil {
		t.Fatalf("生成详情联系表封面失败: %v", err)
	}
	if imageModel.inserted == nil || imageModel.inserted.ImageType != FC2MediaImageTypeDetailContactSheet {
		t.Fatalf("详情图片记录类型错误: %#v", imageModel.inserted)
	}
	file, err := os.Open(imageRow.FilePath)
	if err != nil {
		t.Fatalf("打开详情封面失败: %v", err)
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		t.Fatalf("读取详情封面尺寸失败: %v", err)
	}
	if config.Width != fc2DetailCoverColumns*fc2DetailCoverCellWidth {
		t.Fatalf("详情封面宽度 = %d，期望 %d", config.Width, fc2DetailCoverColumns*fc2DetailCoverCellWidth)
	}
	if len(movieModel.coverUpdates) != 0 || movieModel.movie.CoverImageID != 99 {
		t.Fatalf("详情联系表修改了卡片封面: updates=%#v movie=%#v", movieModel.coverUpdates, movieModel.movie)
	}
}

func TestSetFC2CoverImageAcceptsContactSheet(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "contact-sheet.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建测试联系表失败: %v", err)
	}
	imageModel := &fc2MediaImageModelStub{row: &moviex.FC2MediaImage{
		ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath,
	}}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	image, err := service.SetFC2CoverImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28}, 9)
	if err != nil {
		t.Fatalf("设置联系表封面失败: %v", err)
	}
	if image.ID != 9 {
		t.Fatalf("返回图片 ID = %d，期望 9", image.ID)
	}
	if len(movieModel.coverUpdates) != 1 || movieModel.coverUpdates[0] != 9 {
		t.Fatalf("联系表封面更新记录 = %#v，期望 [9]", movieModel.coverUpdates)
	}
}

func TestEnsureFC2DefaultCoversReusesExistingCardAndDetailSheets(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "contact-sheet.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建测试联系表失败: %v", err)
	}
	detailFilePath := filepath.Join(t.TempDir(), "detail-contact-sheet.png")
	if err := os.WriteFile(detailFilePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建测试详情联系表失败: %v", err)
	}
	mediaRow := &moviex.FC2Media{ID: 7, FC2MovieID: 28}
	mediaModel := &fakeFC2MediaModel{byID: mediaRow}
	imageModel := &fc2MediaImageModelStub{rows: []*moviex.FC2MediaImage{
		{ID: 9, FC2MediaID: mediaRow.ID, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath},
		{ID: 10, FC2MediaID: mediaRow.ID, ImageType: FC2MediaImageTypeDetailContactSheet, FilePath: detailFilePath},
	}}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: mediaRow.FC2MovieID}}
	service := &Service{deps: &dep.Dep{
		FC2MediaModel:      mediaModel,
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	if err := service.ensureFC2DefaultCovers(context.Background(), mediaRow.ID); err != nil {
		t.Fatalf("复用已有卡片与详情联系表失败: %v", err)
	}
	if len(movieModel.coverUpdates) != 1 || movieModel.coverUpdates[0] != 9 {
		t.Fatalf("已有联系表封面更新记录 = %#v，期望 [9]", movieModel.coverUpdates)
	}
}

func TestSetFC2CoverImageRejectsMissingContactSheetFile(t *testing.T) {
	imageModel := &fc2MediaImageModelStub{row: &moviex.FC2MediaImage{
		ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet,
		FilePath: filepath.Join(t.TempDir(), "missing.png"),
	}}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	_, err := service.SetFC2CoverImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28}, 9)
	if !errors.Is(err, ErrInvalidFC2CaptureRequest) {
		t.Fatalf("缺失文件的联系表设置封面错误 = %v", err)
	}
	if len(movieModel.coverUpdates) != 0 {
		t.Fatalf("缺失文件的联系表不应更新封面: %#v", movieModel.coverUpdates)
	}
}

func TestDeleteFC2MediaImageMovesFileToDeletedDirectory(t *testing.T) {
	localImageDir := t.TempDir()
	activeDir := filepath.Join(localImageDir, testFC2CaptureKey)
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("创建联系表目录失败: %v", err)
	}
	filePath := filepath.Join(activeDir, "contact-sheet-test.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建联系表文件失败: %v", err)
	}
	imageModel := &fc2MediaImageModelStub{
		row:  &moviex.FC2MediaImage{ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath},
		rows: []*moviex.FC2MediaImage{{ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath}},
	}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		Config:             config.Config{FC2Media: config.FC2MediaConfig{LocalImageDir: localImageDir}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}
	if _, err := service.DeleteFC2MediaImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28, FC2Key: testFC2CaptureKey}, 9); err != nil {
		t.Fatalf("删除联系表失败: %v", err)
	}
	if !imageModel.deleted {
		t.Fatal("图片记录未删除")
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("原联系表仍然存在: %v", err)
	}
	archivedPath := filepath.Join(localImageDir, "deleted", testFC2CaptureKey, filepath.Base(filePath))
	if content, err := os.ReadFile(archivedPath); err != nil || string(content) != "image" {
		t.Fatalf("删除归档文件无效: content=%q err=%v", content, err)
	}
}

func TestDeleteFC2MediaImageRestoresFileWhenRecordDeleteFails(t *testing.T) {
	localImageDir := t.TempDir()
	activeDir := filepath.Join(localImageDir, testFC2CaptureKey)
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("创建联系表目录失败: %v", err)
	}
	filePath := filepath.Join(activeDir, "contact-sheet-test.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建联系表文件失败: %v", err)
	}
	imageModel := &fc2MediaImageModelStub{
		row:       &moviex.FC2MediaImage{ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath},
		rows:      []*moviex.FC2MediaImage{{ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath}},
		deleteErr: errors.New("delete failed"),
	}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		Config:             config.Config{FC2Media: config.FC2MediaConfig{LocalImageDir: localImageDir}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}
	if _, err := service.DeleteFC2MediaImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28, FC2Key: testFC2CaptureKey}, 9); err == nil {
		t.Fatal("数据库删除失败时应返回错误")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("原联系表未恢复: %v", err)
	}
	archivedPath := filepath.Join(localImageDir, "deleted", testFC2CaptureKey, filepath.Base(filePath))
	if _, err := os.Stat(archivedPath); !os.IsNotExist(err) {
		t.Fatalf("失败后仍存在删除归档: %v", err)
	}
}

func TestDeleteFC2MediaImageSelectsNewestRemainingContactSheetAsCover(t *testing.T) {
	localImageDir := t.TempDir()
	activeDir := filepath.Join(localImageDir, testFC2CaptureKey)
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("创建联系表目录失败: %v", err)
	}
	filePath := filepath.Join(activeDir, "contact-sheet-current.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建联系表文件失败: %v", err)
	}
	current := &moviex.FC2MediaImage{ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath, CreatedOn: 300}
	newestRemainingPath := filepath.Join(activeDir, "contact-sheet-newest-remaining.png")
	olderRemainingPath := filepath.Join(activeDir, "contact-sheet-older-remaining.png")
	for _, path := range []string{newestRemainingPath, olderRemainingPath} {
		if err := os.WriteFile(path, []byte("image"), 0o644); err != nil {
			t.Fatalf("创建剩余联系表失败: %v", err)
		}
	}
	newestRemaining := &moviex.FC2MediaImage{ID: 8, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: newestRemainingPath, CreatedOn: 200}
	olderRemaining := &moviex.FC2MediaImage{ID: 6, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: olderRemainingPath, CreatedOn: 100}
	imageModel := &fc2MediaImageModelStub{
		row:  current,
		rows: []*moviex.FC2MediaImage{current, newestRemaining, olderRemaining},
	}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28, CoverImageID: 0}}
	service := &Service{deps: &dep.Dep{
		Config:             config.Config{FC2Media: config.FC2MediaConfig{LocalImageDir: localImageDir}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	coverImageID, err := service.DeleteFC2MediaImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28, FC2Key: testFC2CaptureKey}, 9)
	if err != nil {
		t.Fatalf("删除当前封面失败: %v", err)
	}
	if coverImageID != 8 {
		t.Fatalf("替补封面 ID = %d，期望 8", coverImageID)
	}
	if len(movieModel.coverUpdates) != 1 || movieModel.coverUpdates[0] != 8 {
		t.Fatalf("封面更新记录 = %#v，期望 [8]", movieModel.coverUpdates)
	}
}

func TestDeleteFC2MediaImageFallsBackToRemainingContactSheet(t *testing.T) {
	localImageDir := t.TempDir()
	activeDir := filepath.Join(localImageDir, testFC2CaptureKey)
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("创建联系表目录失败: %v", err)
	}
	filePath := filepath.Join(activeDir, "contact-sheet-current.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建联系表文件失败: %v", err)
	}
	current := &moviex.FC2MediaImage{ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filePath}
	contactSheetPath := filepath.Join(activeDir, "contact-sheet.png")
	if err := os.WriteFile(contactSheetPath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建联系表文件失败: %v", err)
	}
	imageModel := &fc2MediaImageModelStub{
		row: current,
		rows: []*moviex.FC2MediaImage{
			current,
			{ID: 8, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: contactSheetPath},
		},
	}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28, CoverImageID: 9}}
	service := &Service{deps: &dep.Dep{
		Config:             config.Config{FC2Media: config.FC2MediaConfig{LocalImageDir: localImageDir}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	coverImageID, err := service.DeleteFC2MediaImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28, FC2Key: testFC2CaptureKey}, 9)
	if err != nil {
		t.Fatalf("删除当前联系表失败: %v", err)
	}
	if coverImageID != 8 {
		t.Fatalf("删除联系表后的封面 ID = %d，期望联系表 8", coverImageID)
	}
	if len(movieModel.coverUpdates) != 1 || movieModel.coverUpdates[0] != 8 {
		t.Fatalf("封面更新记录 = %#v，期望 [8]", movieModel.coverUpdates)
	}
}

func TestFormatFFmpegSeconds(t *testing.T) {
	if got := formatFFmpegSeconds(12.5); got != "12.500000" {
		t.Fatalf("格式化结果 = %q", got)
	}
}
