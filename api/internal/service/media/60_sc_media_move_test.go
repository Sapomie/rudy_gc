package media

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"rudy_gc/internal/dep"
	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/platform/config"
	"rudy_gc/internal/repository/moviex"
)

type scMediaMoveFC2MovieModel struct {
	*fc2MovieModelStub
	rows []*moviex.FC2Movie
}

func (m *scMediaMoveFC2MovieModel) ListByIDs(context.Context, []int64) ([]*moviex.FC2Movie, error) {
	return m.rows, nil
}

type scMediaMoveFC2MediaModel struct {
	*fakeFC2MediaModel
	rows []*moviex.FC2Media
}

type scMediaMoveWFolderModel struct {
	moviex.WFolderModel
	nextID int64
	rows   map[string]*moviex.WFolder
}

type scMediaMoveSQLResult int64

func (r scMediaMoveSQLResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r scMediaMoveSQLResult) RowsAffected() (int64, error) { return 1, nil }

func (m *scMediaMoveWFolderModel) FindOneByPathSourceType(_ context.Context, path string, sourceType int64) (*moviex.WFolder, error) {
	row := m.rows[path]
	if row == nil || row.SourceType != sourceType {
		return nil, moviex.ErrNotFound
	}
	copyRow := *row
	return &copyRow, nil
}

func (m *scMediaMoveWFolderModel) Insert(_ context.Context, row *moviex.WFolder) (sql.Result, error) {
	m.nextID++
	copyRow := *row
	copyRow.Id = m.nextID
	m.rows[copyRow.Path] = &copyRow
	return scMediaMoveSQLResult(copyRow.Id), nil
}

func (m *scMediaMoveWFolderModel) Update(_ context.Context, row *moviex.WFolder) error {
	copyRow := *row
	m.rows[copyRow.Path] = &copyRow
	return nil
}

func (m *scMediaMoveFC2MediaModel) ListByKeys(context.Context, []string) ([]*moviex.FC2Media, error) {
	return m.rows, nil
}

func TestBuildFC2ScMediaMovePrecheckItemsUsesFC2Key(t *testing.T) {
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	fullDir := filepath.Join(layout.media, "2026-001", "2026-07-17-001")
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fileName := "FC2PPV-1234567.mp4"
	if err := os.WriteFile(filepath.Join(fullDir, fileName), []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}

	const (
		movieID = int64(42)
		fc2Key  = "FC2PPV-1234567-1"
	)
	movieRow := &moviex.FC2Movie{ID: movieID, FC2Key: fc2Key, Title: "FC2影片标题"}
	mediaRow := &moviex.FC2Media{
		FC2MovieID: movieID,
		FC2Key:     fc2Key,
		FileName:   fileName,
		RootDir:    rootDir,
		FullDir:    fullDir,
	}
	service := NewService(&dep.Dep{
		Config: config.Config{FC2Media: config.FC2MediaConfig{RootDirs: []string{rootDir}}},
		FC2MovieModel: &scMediaMoveFC2MovieModel{
			fc2MovieModelStub: &fc2MovieModelStub{},
			rows:              []*moviex.FC2Movie{movieRow},
		},
		FC2MediaModel: &scMediaMoveFC2MediaModel{
			fakeFC2MediaModel: &fakeFC2MediaModel{},
			rows:              []*moviex.FC2Media{mediaRow},
		},
	})

	items, err := service.buildFC2ScMediaMovePrecheckItems(context.Background(), scMediaMoveModeSingle, "SC-001", []*moviex.GList{{
		Name:       "FC2PPV-1234567__FC2影片标题",
		ScName:     "SC-001",
		MovieType:  moviedomain.SehuatangThreadMovieTypeFC2PPV,
		FC2MovieId: movieID,
	}}, time.Date(2026, 7, 17, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("FC2预检查失败: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("预检查结果数量 = %d，期望 1", len(items))
	}
	item := items[0]
	if item.Status != scMediaMoveStatusPass {
		t.Fatalf("预检查状态 = %q，错误 = %q", item.Status, item.Error)
	}
	if item.MovieType != moviedomain.SehuatangThreadMovieTypeFC2PPV || item.FC2Key != fc2Key {
		t.Fatalf("FC2身份异常: type=%q key=%q", item.MovieType, item.FC2Key)
	}
	if item.MovieJavId != "" {
		t.Fatalf("FC2结果包含 movie_jav_id: %q", item.MovieJavId)
	}
	if item.MovieName != "FC2影片标题" {
		t.Fatalf("影片名称 = %q", item.MovieName)
	}
	if !isPathWithin(item.TargetDir, layout.watched) {
		t.Fatalf("目标目录未落入 watched: %q", item.TargetDir)
	}
}

func TestBuildScMediaMovePlanKeepsMixedMovieIdentities(t *testing.T) {
	plan := buildScMediaMovePlan(scMediaMoveModeAll, "", 100, []*ScMediaMoveItem{
		{
			Status:     scMediaMoveStatusPass,
			MovieType:  moviedomain.SehuatangThreadMovieTypeJAV,
			MovieJavId: "ABP-123",
			MovieName:  "JAV影片",
		},
		{
			Status:    scMediaMoveStatusPass,
			MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
			FC2Key:    "FC2PPV-1234567-1",
			MovieName: "FC2影片",
		},
	})
	if plan.Version != scMediaMovePlanVersion || plan.Total != 2 || plan.Movable != 2 || len(plan.Entries) != 2 {
		t.Fatalf("计划汇总异常: %+v", plan)
	}
	if plan.Entries[0].MovieType != moviedomain.SehuatangThreadMovieTypeJAV || plan.Entries[0].MovieJavId != "ABP-123" {
		t.Fatalf("JAV计划身份异常: %+v", plan.Entries[0])
	}
	if plan.Entries[1].MovieType != moviedomain.SehuatangThreadMovieTypeFC2PPV || plan.Entries[1].FC2Key != "FC2PPV-1234567-1" {
		t.Fatalf("FC2计划身份异常: %+v", plan.Entries[1])
	}
}

func TestLoadLegacyScMediaMovePlanDefaultsToJAV(t *testing.T) {
	rootDir := t.TempDir()
	service := NewService(&dep.Dep{Config: config.Config{Media: config.Media{RootDirs: []string{rootDir}}}})
	layout := buildRootLayout(rootDir)
	if err := ensureRootLayout(layout); err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{
  "version": 1,
  "mode": "single",
  "sc_name": "SC-LEGACY",
  "entries": [{"movie_jav_id":"ABP-123","movie_name":"旧计划"}],
  "checks": [{"status":"pass","movie_jav_id":"ABP-123","movie_name":"旧计划"}]
}`)
	if err := os.WriteFile(scMediaMovePlanPath(layout, scMediaMoveModeSingle, "SC-LEGACY"), payload, 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := service.loadScMediaMovePlan(scMediaMoveModeSingle, "SC-LEGACY")
	if err != nil {
		t.Fatalf("读取旧计划失败: %v", err)
	}
	if len(plan.Entries) != 1 || plan.Entries[0].MovieType != moviedomain.SehuatangThreadMovieTypeJAV {
		t.Fatalf("旧计划影片类型未归一为 JAV: %+v", plan.Entries)
	}
	if len(plan.Checks) != 1 || plan.Checks[0].MovieType != moviedomain.SehuatangThreadMovieTypeJAV {
		t.Fatalf("旧计划检查项类型未归一为 JAV: %+v", plan.Checks)
	}
}

func TestScMediaPlanLayoutFallsBackToFC2Root(t *testing.T) {
	rootDir := t.TempDir()
	service := NewService(&dep.Dep{Config: config.Config{FC2Media: config.FC2MediaConfig{RootDirs: []string{rootDir}}}})
	layout, err := service.scMediaPlanLayout()
	if err != nil {
		t.Fatalf("FC2根目录计划布局失败: %v", err)
	}
	if layout.rootDir != filepath.Clean(rootDir) {
		t.Fatalf("计划根目录 = %q，期望 %q", layout.rootDir, filepath.Clean(rootDir))
	}
}

func TestCommitOneScMediaMoveDispatchesFC2ByMovieType(t *testing.T) {
	service := NewService(&dep.Dep{})
	item := service.commitOneScMediaMove(context.Background(), &scMediaMovePlanEntry{
		MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
		MovieName: "FC2影片",
	}, time.Now())
	if item.MovieType != moviedomain.SehuatangThreadMovieTypeFC2PPV {
		t.Fatalf("提交分派类型 = %q", item.MovieType)
	}
	if item.Error != "计划中的 fc2_key 为空" {
		t.Fatalf("FC2提交校验错误 = %q", item.Error)
	}
}

func TestCommitOneFC2ScMediaMoveMovesFileAndUpdatesSnapshots(t *testing.T) {
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	sourceDir := filepath.Join(layout.media, "2026-001", "2026-07-17-001")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fileName := "FC2PPV-7654321.mp4"
	sourcePath := filepath.Join(sourceDir, fileName)
	if err := os.WriteFile(sourcePath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}

	mediaRow := &moviex.FC2Media{
		ID:         8,
		FC2MovieID: 29,
		FC2Key:     "FC2PPV-7654321-1",
		FileName:   fileName,
		RootDir:    rootDir,
		FullDir:    sourceDir,
		BirthTime:  123,
	}
	mediaModel := &moveRemovedFC2MediaModel{fakeFC2MediaModel: &fakeFC2MediaModel{existing: mediaRow}}
	movieModel := &moveRemovedFC2MovieModel{fc2MovieModelStub: &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: mediaRow.FC2MovieID, FC2Key: mediaRow.FC2Key}}}
	service := NewService(&dep.Dep{
		Config:         config.Config{FC2Media: config.FC2MediaConfig{RootDirs: []string{rootDir}}},
		WFolderModel:   &scMediaMoveWFolderModel{rows: make(map[string]*moviex.WFolder)},
		FC2MediaModel:  mediaModel,
		FC2MovieModel:  movieModel,
		FC2ScStatModel: &moveRemovedFC2ScStatModel{},
	})
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.Local)
	item := service.commitOneFC2ScMediaMove(context.Background(), &scMediaMovePlanEntry{
		MovieType:  moviedomain.SehuatangThreadMovieTypeFC2PPV,
		FC2Key:     mediaRow.FC2Key,
		MovieName:  "FC2影片",
		SourcePath: sourcePath,
	}, now)
	if item.Error != "" {
		t.Fatalf("FC2移动失败: %s", item.Error)
	}
	if _, err := os.Stat(sourcePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("源文件仍存在或检查失败: %v", err)
	}
	if _, err := os.Stat(item.TargetPath); err != nil {
		t.Fatalf("目标文件不存在: %v", err)
	}
	if !isPathWithin(item.TargetDir, layout.watched) {
		t.Fatalf("目标目录未落入 watched: %q", item.TargetDir)
	}
	if mediaModel.updated == nil || mediaModel.updated.FullDir != item.TargetDir || mediaModel.updated.DirectoryID <= 0 {
		t.Fatalf("fc2_media 路径更新异常: %+v", mediaModel.updated)
	}
	if movieModel.updatedMovieID != mediaRow.FC2MovieID || movieModel.mediaPresent != 1 || movieModel.mediaBirth != mediaRow.BirthTime {
		t.Fatalf("FC2媒体快照异常: movie_id=%d present=%d birth=%d", movieModel.updatedMovieID, movieModel.mediaPresent, movieModel.mediaBirth)
	}
}
