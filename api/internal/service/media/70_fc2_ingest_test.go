package media

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/repository/moviex"
)

type fakeFC2MediaModel struct {
	insertErr   error
	insertCount int
	updateCount int
	existing    *moviex.FC2Media
	byID        *moviex.FC2Media
	findErr     error
}

func (m *fakeFC2MediaModel) Insert(_ context.Context, data *moviex.FC2Media) (*moviex.FC2Media, error) {
	m.insertCount++
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	return data, nil
}

func (m *fakeFC2MediaModel) FindOne(_ context.Context, id int64) (*moviex.FC2Media, error) {
	if m.byID != nil && m.byID.ID == id {
		return m.byID, nil
	}
	return nil, moviex.ErrNotFound
}

func (m *fakeFC2MediaModel) FindOneByMovieID(context.Context, int64) (*moviex.FC2Media, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.existing != nil {
		return m.existing, nil
	}
	return nil, moviex.ErrNotFound
}

func (m *fakeFC2MediaModel) FindOneByKey(context.Context, string) (*moviex.FC2Media, error) {
	return nil, moviex.ErrNotFound
}

func (m *fakeFC2MediaModel) FindOneByFileName(context.Context, string) (*moviex.FC2Media, error) {
	return nil, moviex.ErrNotFound
}

func (m *fakeFC2MediaModel) ListAll(context.Context) ([]*moviex.FC2Media, error) {
	return []*moviex.FC2Media{}, nil
}

func (m *fakeFC2MediaModel) ListByFileNames(context.Context, []string) ([]*moviex.FC2Media, error) {
	return []*moviex.FC2Media{}, nil
}

func (m *fakeFC2MediaModel) ListByKeys(context.Context, []string) ([]*moviex.FC2Media, error) {
	return []*moviex.FC2Media{}, nil
}

func (m *fakeFC2MediaModel) ListByMovieIDs(context.Context, []int64) ([]*moviex.FC2Media, error) {
	return []*moviex.FC2Media{}, nil
}

func (m *fakeFC2MediaModel) ListByFullDirPrefixes(context.Context, []string) ([]*moviex.FC2Media, error) {
	return []*moviex.FC2Media{}, nil
}

func (m *fakeFC2MediaModel) ListSmartPickCandidates(context.Context, moviex.FC2SmartPickMediaFilter) ([]*moviex.FC2Media, error) {
	return []*moviex.FC2Media{}, nil
}

func (m *fakeFC2MediaModel) Update(context.Context, *moviex.FC2Media) error {
	m.updateCount++
	return nil
}

func (m *fakeFC2MediaModel) Delete(context.Context, int64) error { return nil }

func TestMergeFC2IngestResultNilRootResult(t *testing.T) {
	result := &FC2IngestResult{}

	mergeFC2IngestResult(result, nil)

	if result.Precheck.Total != 0 || result.Total != 0 || result.Success != 0 || result.Failed != 0 {
		t.Fatalf("nil root result changed aggregate: %+v", result)
	}
}

func TestFC2AlbumNames(t *testing.T) {
	if fc2SourceAlbumName != "下载中" {
		t.Fatalf("FC2来源相册 = %q", fc2SourceAlbumName)
	}
	if fc2MediaAlbumName != "FcMedia" {
		t.Fatalf("FC2目标相册 = %q", fc2MediaAlbumName)
	}
	if fc2MediaAlbumName == mediaAlbumName {
		t.Fatal("FC2目标相册与普通媒体目标相册重复")
	}
}

func TestReserveFC2BatchMovieRejectsDuplicateMovie(t *testing.T) {
	seen := make(map[int64]string)
	first := &fc2IngestPreparedItem{sourcePath: "/tmp/FC2PPV-3377076.mp4", meta: fc2RawMediaMeta{key: "FC2PPV-3377076-1"}, fc2MovieID: 72}
	second := &fc2IngestPreparedItem{sourcePath: "/tmp/FC2PPV-3377076-C.mp4", meta: fc2RawMediaMeta{key: "FC2PPV-3377076-1"}, fc2MovieID: 72}
	if !reserveFC2BatchMovie(first, &FC2IngestFileItem{}, seen) {
		t.Fatal("first movie was rejected")
	}
	item := &FC2IngestFileItem{}
	if reserveFC2BatchMovie(second, item, seen) {
		t.Fatal("duplicate movie was accepted")
	}
	if item.Error == "" {
		t.Fatal("duplicate movie error is empty")
	}
}

func TestInsertFC2MediaNeverUpdatesExistingRow(t *testing.T) {
	duplicateErr := errors.New("Error 1062: Duplicate entry '72' for key 'uk_fc2_media_movie_id'")
	model := &fakeFC2MediaModel{insertErr: duplicateErr}
	service := NewService(&dep.Dep{FC2MediaModel: model})
	_, err := service.insertFC2Media(context.Background(), &moviex.FC2Media{FC2MovieID: 72})
	if err == nil {
		t.Fatal("duplicate insert returned nil error")
	}
	if model.insertCount != 1 || model.updateCount != 0 {
		t.Fatalf("insert=%d update=%d", model.insertCount, model.updateCount)
	}
}

func TestEnsureFC2MediaInsertableRejectsExistingMedia(t *testing.T) {
	model := &fakeFC2MediaModel{existing: &moviex.FC2Media{FC2MovieID: 72, FileName: "existing.mp4"}}
	service := NewService(&dep.Dep{FC2MediaModel: model})
	err := service.ensureFC2MediaInsertable(context.Background(), 72, "FC2PPV-3377076-1")
	if err == nil {
		t.Fatal("existing media was accepted")
	}
}

func TestMergeFC2IngestResultCounts(t *testing.T) {
	result := &FC2IngestResult{}
	rootResult := &fc2IngestRootResult{
		Precheck: IngestPrecheckStat{Total: 2, Passed: 1, Failed: 1},
		Items: []*FC2IngestFileItem{
			{Status: ingestItemStatusPass},
			{Status: ingestItemStatusFail, Error: "failed"},
		},
	}

	mergeFC2IngestResult(result, rootResult)

	if result.Precheck.Total != 2 || result.Precheck.Passed != 1 || result.Precheck.Failed != 1 {
		t.Fatalf("unexpected precheck aggregate: %+v", result.Precheck)
	}
	if result.Total != 2 || result.Success != 1 || result.Failed != 1 || len(result.Items) != 2 {
		t.Fatalf("unexpected item aggregate: total=%d success=%d failed=%d items=%d", result.Total, result.Success, result.Failed, len(result.Items))
	}
}

func TestParseFC2RawMediaMeta(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		code        string
		resolvedKey string
		hasSub      int64
		selfMake    int64
		target      string
	}{
		{
			name:        "普通番号由来源解析第一条影片",
			fileName:    "FC2PPV-2711719-C~1.mp4",
			code:        "FC2PPV-2711719",
			resolvedKey: "FC2PPV-2711719-1",
			hasSub:      int64(consts.FilmHasSub),
			selfMake:    int64(consts.FilmSelfMake),
			target:      "15122252531-2711719-1_sub_self_nomsk.mp4",
		},
		{
			name:        "普通番号可由来源解析第二条影片",
			fileName:    "FC2PPV-2724256-C.mp4",
			code:        "FC2PPV-2724256",
			resolvedKey: "FC2PPV-2724256-2",
			hasSub:      int64(consts.FilmHasSub),
			selfMake:    int64(consts.FilmNoSelfMake),
			target:      "15122252531-2724256-2_sub_nomsk.mp4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := parseFC2RawMediaMeta(tt.fileName)
			if err != nil {
				t.Fatalf("parseFC2RawMediaMeta() error = %v", err)
			}
			if meta.code != tt.code || meta.key != "" || meta.hasSub != tt.hasSub || meta.selfMake != tt.selfMake {
				t.Fatalf("parseFC2RawMediaMeta() = %#v", meta)
			}
			meta.key = tt.resolvedKey
			if got := buildFC2TargetFileName(meta, int64(consts.FilmNoMosaic)); got != tt.target {
				t.Fatalf("buildFC2TargetFileName() = %q, want %q", got, tt.target)
			}
		})
	}
}

func TestParseFC2RawMediaMetaRejectsEmbeddedSequenceAndInvalidPrefix(t *testing.T) {
	for _, fileName := range []string{"FC2PPV-3377076-2.mp4", "FC2-PPV-4694056-C~1.mp4"} {
		if _, err := parseFC2RawMediaMeta(fileName); err == nil {
			t.Fatalf("文件名应解析失败: %s", fileName)
		}
	}
}

func TestRequireUniqueFC2DownloadSource(t *testing.T) {
	makeCandidate := func(key string, movieID int64) *fc2ResolvedDownloadSource {
		return &fc2ResolvedDownloadSource{
			movie:  &moviex.FC2Movie{ID: movieID, FC2Key: key},
			thread: &moviex.ShSehuatangThread{Id: movieID + 100, FC2MovieID: movieID},
			favoriteSource: &favoriteAlbumSourceInfo{
				favoriteAlbumID: 1,
				item:            &moviex.TmAlbumItem{Id: movieID + 200, SourceRowId: movieID + 300},
			},
		}
	}

	first := makeCandidate("FC2PPV-3377076-1", 72)
	got, err := requireUniqueFC2DownloadSource("FC2PPV-3377076", []*fc2ResolvedDownloadSource{first})
	if err != nil || got != first {
		t.Fatalf("唯一来源解析失败: got=%#v err=%v", got, err)
	}
	if _, err := requireUniqueFC2DownloadSource("FC2PPV-3377076", nil); err == nil {
		t.Fatal("缺少下载中来源时应失败")
	}
	if _, err := requireUniqueFC2DownloadSource("FC2PPV-3377076", []*fc2ResolvedDownloadSource{first, makeCandidate("FC2PPV-3377076-2", 73)}); err == nil {
		t.Fatal("多个下载中来源时应判定为歧义")
	}
}

func TestBuildFC2MediaAliasMatchesWMedia(t *testing.T) {
	birthTime := time.Date(2026, 7, 12, 12, 30, 0, 0, time.Local).Unix()
	const fileSize int64 = 1427794666

	got := buildFC2MediaAlias("FC2PPV-4862980-1", birthTime, fileSize)
	want := buildMediaAlias(rawMovieMeta{movieName: "15122252531-4862980-1"}, birthTime, fileSize)
	if got != want {
		t.Fatalf("buildFC2MediaAlias() = %q, w_media alias = %q", got, want)
	}
	if want != "15122252531-4862980-1_2026-07-12_1427794666" {
		t.Fatalf("buildMediaAlias() = %q", want)
	}
}

func TestFC2IngestPlanKeepsTorrentAlbumSource(t *testing.T) {
	prepared := &fc2IngestPreparedItem{
		sourcePath:        "/tmp/FC2PPV-4862980-C~1.mp4",
		meta:              fc2RawMediaMeta{code: "FC2PPV-4862980", key: "FC2PPV-4862980-1", ext: ".mp4", hasSub: int64(consts.FilmHasSub), selfMake: int64(consts.FilmSelfMake)},
		fc2MovieID:        9,
		sourceThreadID:    99,
		sourceTorrentHash: "95DD9145D03EFF79B5A91DAC47A09461F9B270F0",
		favoriteSource: &favoriteAlbumSourceInfo{
			favoriteAlbumID: 3,
			infoHash:        "95DD9145D03EFF79B5A91DAC47A09461F9B270F0",
			item: &moviex.TmAlbumItem{
				Id:          3606,
				AlbumId:     3,
				SourceType:  "sh_magnet",
				SourceRowId: 147,
			},
		},
	}
	plan := buildFC2IngestPlan(fc2MediaRoot{sourceDir: "/tmp", rootDir: "/root"}, []*fc2IngestPreparedItem{prepared}, []*FC2IngestFileItem{{Status: ingestItemStatusPass}})
	if len(plan.Entries) != 1 {
		t.Fatalf("plan entries = %d", len(plan.Entries))
	}
	entry := plan.Entries[0]
	if entry.FavoriteAlbumID != 3 || entry.FavoriteItemID != 3606 || entry.FavoriteSourceType != "sh_magnet" || entry.FavoriteSourceRowID != 147 || entry.FavoriteInfoHash != prepared.sourceTorrentHash {
		t.Fatalf("torrent album source missing from plan: %+v", entry)
	}
}

func TestBuildFC2PreparedFromPlanRestoresTorrentAlbumSource(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "FC2PPV-4862980-C~1.mp4")
	if err := os.WriteFile(sourcePath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := &fc2IngestPlanEntry{
		SourcePath:          sourcePath,
		FC2Code:             "FC2PPV-4862980",
		FC2Key:              "FC2PPV-4862980-1",
		Ext:                 ".mp4",
		HasSub:              int64(consts.FilmHasSub),
		SelfMake:            int64(consts.FilmSelfMake),
		HasMask:             int64(consts.FilmNoMosaic),
		FC2MovieID:          9,
		SourceThreadID:      99,
		SourceTorrentHash:   "95DD9145D03EFF79B5A91DAC47A09461F9B270F0",
		FavoriteAlbumID:     3,
		FavoriteItemID:      3606,
		FavoriteSourceType:  "sh_magnet",
		FavoriteSourceRowID: 147,
		FavoriteInfoHash:    "95DD9145D03EFF79B5A91DAC47A09461F9B270F0",
	}
	prepared, err := buildFC2PreparedFromPlan(fc2MediaRoot{sourceDir: filepath.Dir(sourcePath), rootDir: t.TempDir()}, entry)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.favoriteSource == nil || prepared.favoriteSource.item == nil {
		t.Fatal("torrent album source was not restored")
	}
	if prepared.favoriteSource.favoriteAlbumID != 3 || prepared.favoriteSource.item.Id != 3606 || prepared.favoriteSource.item.SourceType != "sh_magnet" || prepared.favoriteSource.item.SourceRowId != 147 {
		t.Fatalf("restored torrent album source mismatch: %+v", prepared.favoriteSource)
	}
}

func TestFC2RootLayoutMatchesMediaLayout(t *testing.T) {
	layout := buildRootLayout(filepath.Join(t.TempDir(), "f1"))
	if err := ensureRootLayout(layout); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		layout.ingestNew,
		layout.tmp,
		layout.failed,
		layout.rollback,
		layout.removed,
		layout.waitMovie,
		layout.media,
		layout.watched,
	} {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("目录未创建: %s, err=%v", path, err)
		}
	}
}

func TestFC2TargetDirectoryUsesMediaAllocator(t *testing.T) {
	layout := buildRootLayout(filepath.Join(t.TempDir(), "f1"))
	if err := ensureRootLayout(layout); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.Local)
	want, err := previewTargetDirectory(layout, now)
	if err != nil {
		t.Fatal(err)
	}
	got, err := allocateTargetDirectoryPathUnder(layout.media, now)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("FC2目标目录=%s, Media目标目录=%s", got, want)
	}
}

func TestListFC2SourceFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"FC2PPV-2-C.mp4", "FC2PPV-1-C.mp4", "FC2PPV-3-C.mp4", "._FC2PPV-1-C.mp4", ".hidden.mkv", "note.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"FC2PPV-2-C.mp4", "FC2PPV-1-C.mp4"} {
		if err := os.Truncate(filepath.Join(dir, name), minMediaFileSize); err != nil {
			t.Fatal(err)
		}
	}
	files, err := listFC2SourceFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(dir, "FC2PPV-1-C.mp4"), filepath.Join(dir, "FC2PPV-2-C.mp4")}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("listFC2SourceFiles() = %#v, want %#v", files, want)
	}
}
