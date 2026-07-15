package media

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/repository/moviex"
)

type moveRemovedFC2MediaModel struct {
	*fakeFC2MediaModel
	updated *moviex.FC2Media
}

func (m *moveRemovedFC2MediaModel) FindOneByKey(_ context.Context, key string) (*moviex.FC2Media, error) {
	row := m.updated
	if row == nil {
		row = m.existing
	}
	if row == nil || row.FC2Key != key {
		return nil, moviex.ErrNotFound
	}
	cp := *row
	return &cp, nil
}

func (m *moveRemovedFC2MediaModel) FindOneByMovieID(_ context.Context, movieID int64) (*moviex.FC2Media, error) {
	row := m.updated
	if row == nil {
		row = m.existing
	}
	if row == nil || row.FC2MovieID != movieID {
		return nil, moviex.ErrNotFound
	}
	cp := *row
	return &cp, nil
}

func (m *moveRemovedFC2MediaModel) Update(_ context.Context, row *moviex.FC2Media) error {
	cp := *row
	m.updated = &cp
	return nil
}

type moveRemovedFC2ScStatModel struct{}

type moveRemovedFC2MovieModel struct {
	*fc2MovieModelStub
	updatedMovieID int64
	mediaPresent   int64
	mediaBirth     int64
}

func (m *moveRemovedFC2MovieModel) UpdateMediaSnapshot(_ context.Context, movieID, present, birthTime, _ int64) error {
	m.updatedMovieID = movieID
	m.mediaPresent = present
	m.mediaBirth = birthTime
	return nil
}

func (*moveRemovedFC2ScStatModel) Insert(context.Context, *moviex.FC2ScStat) (sql.Result, error) {
	return nil, nil
}

func (*moveRemovedFC2ScStatModel) FindOne(context.Context, int64) (*moviex.FC2ScStat, error) {
	return nil, moviex.ErrNotFound
}

func (*moveRemovedFC2ScStatModel) FindOneByMovieID(context.Context, int64) (*moviex.FC2ScStat, error) {
	return nil, moviex.ErrNotFound
}

func (*moveRemovedFC2ScStatModel) FindOneByKey(context.Context, string) (*moviex.FC2ScStat, error) {
	return nil, moviex.ErrNotFound
}

func (*moveRemovedFC2ScStatModel) Update(context.Context, *moviex.FC2ScStat) error { return nil }
func (*moveRemovedFC2ScStatModel) Delete(context.Context, int64) error             { return nil }

func (*moveRemovedFC2ScStatModel) ListByMovieIDs(context.Context, []int64) ([]*moviex.FC2ScStat, error) {
	return []*moviex.FC2ScStat{}, nil
}

func (*moveRemovedFC2ScStatModel) ListAllMovieIDs(context.Context) ([]int64, error) {
	return []int64{}, nil
}

func TestMoveFC2MediaToRemoved(t *testing.T) {
	rootDir := t.TempDir()
	fullDir := filepath.Join(rootDir, "FC2Media")
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(fullDir, "FC2PPV-1234567.mp4")
	if err := os.WriteFile(sourcePath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}

	mediaRow := &moviex.FC2Media{
		ID:                8,
		FC2MovieID:        29,
		FC2Key:            "FC2PPV-1234567-1",
		SourceTorrentHash: "ABC123",
		RootDir:           rootDir,
		FullDir:           fullDir,
		FileName:          filepath.Base(sourcePath),
		IsRemoved:         int64(consts.FilmIsNotRemoved),
		BirthTime:         100,
	}
	mediaModel := &moveRemovedFC2MediaModel{fakeFC2MediaModel: &fakeFC2MediaModel{existing: mediaRow}}
	movieModel := &moveRemovedFC2MovieModel{fc2MovieModelStub: &fc2MovieModelStub{movie: &moviex.FC2Movie{ID: mediaRow.FC2MovieID, FC2Key: mediaRow.FC2Key}}}
	service := NewService(&dep.Dep{
		FC2MediaModel:  mediaModel,
		FC2MovieModel:  movieModel,
		FC2ScStatModel: &moveRemovedFC2ScStatModel{},
	})

	result, err := service.MoveFC2MediaToRemoved(context.Background(), mediaRow.FC2Key)
	if err != nil {
		t.Fatalf("移动FC2媒体失败: %v", err)
	}
	if result == nil || !result.Ok {
		t.Fatalf("移动结果异常: %+v", result)
	}
	if _, err := os.Stat(sourcePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("源文件仍然存在或检查失败: %v", err)
	}
	if _, err := os.Stat(result.TargetPath); err != nil {
		t.Fatalf("删除目录中的目标文件不存在: %v", err)
	}
	wantDir := filepath.Join(rootDir, processDirName, removedDirName)
	if filepath.Dir(result.TargetPath) != wantDir {
		t.Fatalf("目标目录 = %q，期望 %q", filepath.Dir(result.TargetPath), wantDir)
	}
	if mediaModel.updated == nil {
		t.Fatal("fc2_media 未更新")
	}
	if mediaModel.updated.IsRemoved != int64(consts.FilmIsRemoved) || mediaModel.updated.RemoveTime <= 0 {
		t.Fatalf("fc2_media 删除状态异常: %+v", mediaModel.updated)
	}
	if movieModel.updatedMovieID != mediaRow.FC2MovieID || movieModel.mediaPresent != 0 || movieModel.mediaBirth != 0 {
		t.Fatalf("fc2_movie 媒体快照异常: movie_id=%d present=%d birth=%d", movieModel.updatedMovieID, movieModel.mediaPresent, movieModel.mediaBirth)
	}
}
