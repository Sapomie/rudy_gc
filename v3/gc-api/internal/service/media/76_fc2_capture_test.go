package media

import (
	"context"
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"rudy_gc/internal/config"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
)

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

func (*fc2MovieModelStub) ListCards(context.Context, int64, int64, bool, bool) ([]*moviex.FC2Movie, error) {
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

func (*fc2MediaImageModelStub) ListScreenshotsByMediaIDs(context.Context, []int64) ([]*moviex.FC2MediaImage, error) {
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
	got := buildFC2ContactSheetTimestamps(120, 4)
	want := []float64{15, 45, 75, 105}
	if len(got) != len(want) {
		t.Fatalf("时间点数量 = %d，期望 %d", len(got), len(want))
	}
	for index := range want {
		if math.Abs(got[index]-want[index]) > 0.000001 {
			t.Fatalf("时间点[%d] = %f，期望 %f", index, got[index], want[index])
		}
	}
}

func TestValidateFC2ContactSheetOptions(t *testing.T) {
	mediaRow := &moviex.FC2Media{ID: 1, Duration: 120}
	count, err := validateFC2ContactSheetOptions(mediaRow, FC2ContactSheetOptions{Columns: 4, Rows: 3, CellWidth: 320})
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	if count != 12 {
		t.Fatalf("帧数 = %d，期望 12", count)
	}

	_, err = validateFC2ContactSheetOptions(mediaRow, FC2ContactSheetOptions{Columns: 0, Rows: 3, CellWidth: 320})
	if !errors.Is(err, ErrInvalidFC2CaptureRequest) {
		t.Fatalf("无效参数错误 = %v", err)
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
		RootDir:  mediaRoot,
		FullDir:  sourceDir,
		FileName: fileName,
	})
	if err != nil {
		t.Fatalf("准备截图目录失败: %v", err)
	}
	if gotSourcePath != sourcePath {
		t.Fatalf("媒体路径 = %q，期望 %q", gotSourcePath, sourcePath)
	}
	wantCaptureDir := filepath.Join(localImageDir, "7")
	if gotCaptureDir != wantCaptureDir {
		t.Fatalf("截图目录 = %q，期望 %q", gotCaptureDir, wantCaptureDir)
	}
	if stat, statErr := os.Stat(gotCaptureDir); statErr != nil || !stat.IsDir() {
		t.Fatalf("截图目录未创建: stat=%v err=%v", stat, statErr)
	}
}

func TestGenerateFC2ScreenshotWritesToConfiguredLocalImageDir(t *testing.T) {
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
	image, err := service.GenerateFC2Screenshot(context.Background(), &moviex.FC2Media{
		ID:         7,
		FC2MovieID: 28,
		RootDir:    mediaRoot,
		FullDir:    sourceDir,
		FileName:   fileName,
		Duration:   2,
	}, FC2ScreenshotOptions{TimeSeconds: 0.5})
	if err != nil {
		t.Fatalf("生成截图失败: %v", err)
	}
	wantDir := filepath.Join(localImageDir, "7")
	if filepath.Dir(image.FilePath) != wantDir {
		t.Fatalf("截图目录 = %q，期望 %q", filepath.Dir(image.FilePath), wantDir)
	}
	if imageModel.inserted == nil || imageModel.inserted.FilePath != image.FilePath {
		t.Fatalf("图片记录未写入正确路径: %#v", imageModel.inserted)
	}
	if stat, statErr := os.Stat(image.FilePath); statErr != nil || !stat.Mode().IsRegular() {
		t.Fatalf("截图文件未生成: stat=%v err=%v", stat, statErr)
	}
	if len(movieModel.coverUpdates) != 1 || movieModel.coverUpdates[0] != image.ID {
		t.Fatalf("新截图未自动设为封面: %#v", movieModel.coverUpdates)
	}
}

func TestSetFC2CoverImageAcceptsScreenshot(t *testing.T) {
	imageModel := &fc2MediaImageModelStub{row: &moviex.FC2MediaImage{
		ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeScreenshot,
	}}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	image, err := service.SetFC2CoverImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28}, 9)
	if err != nil {
		t.Fatalf("设置截图封面失败: %v", err)
	}
	if image.ID != 9 {
		t.Fatalf("返回图片 ID = %d，期望 9", image.ID)
	}
	if len(movieModel.coverUpdates) != 1 || movieModel.coverUpdates[0] != 9 {
		t.Fatalf("封面更新记录 = %#v，期望 [9]", movieModel.coverUpdates)
	}
}

func TestSetFC2CoverImageRejectsContactSheet(t *testing.T) {
	imageModel := &fc2MediaImageModelStub{row: &moviex.FC2MediaImage{
		ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet,
	}}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	_, err := service.SetFC2CoverImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28}, 9)
	if !errors.Is(err, ErrInvalidFC2CaptureRequest) {
		t.Fatalf("联系表设置封面错误 = %v", err)
	}
	if len(movieModel.coverUpdates) != 0 {
		t.Fatalf("联系表不应更新封面: %#v", movieModel.coverUpdates)
	}
}

func TestDeleteFC2MediaImageMovesFileToDeletedDirectory(t *testing.T) {
	localImageDir := t.TempDir()
	activeDir := filepath.Join(localImageDir, "7")
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("创建截图目录失败: %v", err)
	}
	filePath := filepath.Join(activeDir, "screenshot-test.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建截图文件失败: %v", err)
	}
	imageModel := &fc2MediaImageModelStub{row: &moviex.FC2MediaImage{ID: 9, FC2MediaID: 7, FilePath: filePath}}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		Config:             config.Config{FC2Media: config.FC2MediaConfig{LocalImageDir: localImageDir}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}
	if _, err := service.DeleteFC2MediaImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28}, 9); err != nil {
		t.Fatalf("删除截图失败: %v", err)
	}
	if !imageModel.deleted {
		t.Fatal("图片记录未删除")
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("原截图仍然存在: %v", err)
	}
	archivedPath := filepath.Join(localImageDir, "deleted", "7", filepath.Base(filePath))
	if content, err := os.ReadFile(archivedPath); err != nil || string(content) != "image" {
		t.Fatalf("删除归档文件无效: content=%q err=%v", content, err)
	}
}

func TestDeleteFC2MediaImageRestoresFileWhenRecordDeleteFails(t *testing.T) {
	localImageDir := t.TempDir()
	activeDir := filepath.Join(localImageDir, "7")
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("创建截图目录失败: %v", err)
	}
	filePath := filepath.Join(activeDir, "screenshot-test.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建截图文件失败: %v", err)
	}
	imageModel := &fc2MediaImageModelStub{
		row:       &moviex.FC2MediaImage{ID: 9, FC2MediaID: 7, FilePath: filePath},
		deleteErr: errors.New("delete failed"),
	}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28}}
	service := &Service{deps: &dep.Dep{
		Config:             config.Config{FC2Media: config.FC2MediaConfig{LocalImageDir: localImageDir}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}
	if _, err := service.DeleteFC2MediaImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28}, 9); err == nil {
		t.Fatal("数据库删除失败时应返回错误")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("原截图未恢复: %v", err)
	}
	archivedPath := filepath.Join(localImageDir, "deleted", "7", filepath.Base(filePath))
	if _, err := os.Stat(archivedPath); !os.IsNotExist(err) {
		t.Fatalf("失败后仍存在删除归档: %v", err)
	}
}

func TestDeleteFC2MediaImageSelectsNewestRemainingScreenshotAsCover(t *testing.T) {
	localImageDir := t.TempDir()
	activeDir := filepath.Join(localImageDir, "7")
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("创建截图目录失败: %v", err)
	}
	filePath := filepath.Join(activeDir, "screenshot-current.png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建截图文件失败: %v", err)
	}
	current := &moviex.FC2MediaImage{ID: 9, FC2MediaID: 7, ImageType: FC2MediaImageTypeScreenshot, FilePath: filePath, CreatedOn: 300}
	newestRemaining := &moviex.FC2MediaImage{ID: 8, FC2MediaID: 7, ImageType: FC2MediaImageTypeScreenshot, CreatedOn: 200}
	olderRemaining := &moviex.FC2MediaImage{ID: 6, FC2MediaID: 7, ImageType: FC2MediaImageTypeScreenshot, CreatedOn: 100}
	imageModel := &fc2MediaImageModelStub{
		row:  current,
		rows: []*moviex.FC2MediaImage{current, newestRemaining, olderRemaining},
	}
	movieModel := &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: 28, CoverImageID: 9}}
	service := &Service{deps: &dep.Dep{
		Config:             config.Config{FC2Media: config.FC2MediaConfig{LocalImageDir: localImageDir}},
		FC2MediaImageModel: imageModel,
		FC2MovieModel:      movieModel,
	}}

	coverImageID, err := service.DeleteFC2MediaImage(context.Background(), &moviex.FC2Media{ID: 7, FC2MovieID: 28}, 9)
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

func TestFormatFFmpegSeconds(t *testing.T) {
	if got := formatFFmpegSeconds(12.5); got != "12.500000" {
		t.Fatalf("格式化结果 = %q", got)
	}
}
