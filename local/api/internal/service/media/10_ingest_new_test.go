package media

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type ingestWMediaModel struct {
	moviex.WMediaModel
	existing    *moviex.WMedia
	findErr     error
	insertErr   error
	updateErr   error
	insertCount int
	updateCount int
	updated     *moviex.WMedia
}

func (m *ingestWMediaModel) FindOneByMovieJavIdSourceType(_ context.Context, movieJavID string, sourceType int64) (*moviex.WMedia, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.existing != nil && m.existing.MovieJavId == movieJavID && m.existing.SourceType == sourceType {
		copyRow := *m.existing
		return &copyRow, nil
	}
	return nil, moviex.ErrNotFound
}

func (m *ingestWMediaModel) Insert(_ context.Context, _ *moviex.WMedia) (sql.Result, error) {
	m.insertCount++
	return nil, m.insertErr
}

func (m *ingestWMediaModel) FindOneByMovieJavId(_ context.Context, movieJavID string) (*moviex.WMedia, error) {
	if m.existing != nil && m.existing.MovieJavId == movieJavID {
		copyRow := *m.existing
		return &copyRow, nil
	}
	return nil, moviex.ErrNotFound
}

func (m *ingestWMediaModel) Update(_ context.Context, row *moviex.WMedia) error {
	m.updateCount++
	if m.updateErr != nil {
		return m.updateErr
	}
	copyRow := *row
	m.updated = &copyRow
	m.existing = &copyRow
	return nil
}

type ingestGScStatModel struct {
	moviex.GScStatModel
}

func (*ingestGScStatModel) FindOneByMovieJavId(context.Context, string) (*moviex.GScStat, error) {
	return nil, moviex.ErrNotFound
}

func TestMergeIngestRootResultNilRootResult(t *testing.T) {
	result := &IngestNewResult{}

	mergeIngestRootResult(result, nil)

	if result.Precheck.Total != 0 || result.Total != 0 || result.Success != 0 || result.Failed != 0 {
		t.Fatalf("nil root result changed aggregate: %+v", result)
	}
}

func TestMergeIngestRootResultCounts(t *testing.T) {
	result := &IngestNewResult{}
	rootResult := &ingestRootResult{
		Precheck: IngestPrecheckStat{
			Total:  2,
			Passed: 1,
			Failed: 1,
		},
		Items: []*IngestFileItem{
			{Status: ingestItemStatusPass},
			{Status: ingestItemStatusFail, Error: "failed"},
		},
	}

	mergeIngestRootResult(result, rootResult)

	if result.Precheck.Total != 2 || result.Precheck.Passed != 1 || result.Precheck.Failed != 1 {
		t.Fatalf("unexpected precheck aggregate: %+v", result.Precheck)
	}
	if result.Total != 2 || result.Success != 1 || result.Failed != 1 || len(result.Items) != 2 {
		t.Fatalf("unexpected item aggregate: total=%d success=%d failed=%d items=%d", result.Total, result.Success, result.Failed, len(result.Items))
	}
}

func TestParseRawMovieMetaSupportsSerialLetterSuffix(t *testing.T) {
	meta, err := parseRawMovieMeta("OLM-044N~1.mp4")
	if err != nil {
		t.Fatalf("parseRawMovieMeta failed: %v", err)
	}
	if meta.movieName != "OLM-044N" {
		t.Fatalf("unexpected movie name: %s", meta.movieName)
	}
	if meta.ext != ".mp4" {
		t.Fatalf("unexpected extension: %s", meta.ext)
	}
	if meta.selfMake != consts.FilmSelfMake {
		t.Fatalf("unexpected selfMake: %d", meta.selfMake)
	}
}

func TestInspectMediaIngestActionWarnsAndAllowsActiveReplacement(t *testing.T) {
	const movieJavID = "ABP-123"
	rootDir := t.TempDir()
	oldDir := filepath.Join(rootDir, mediaDirName, "old")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(oldDir, "existing.mp4")
	if err := os.WriteFile(oldPath, []byte("old-video"), 0o644); err != nil {
		t.Fatal(err)
	}
	mediaModel := &ingestWMediaModel{existing: &moviex.WMedia{
		Id:         12,
		MovieJavId: movieJavID,
		FileName:   "existing.mp4",
		SourceType: consts.WMediaSourceNative,
		IsRemoved:  consts.FilmIsNotRemoved,
		RootDir:    rootDir,
		FullDir:    oldDir,
		UpdatedOn:  100,
	}}
	service := NewService(&dep.Dep{WMediaModel: mediaModel})

	action, snapshot, notice, err := service.inspectMediaIngestAction(context.Background(), movieJavID)
	if err != nil {
		t.Fatalf("活动影片替换检查失败: %v", err)
	}
	if action != ingestActionReplace || snapshot == nil || snapshot.ID != mediaModel.existing.Id {
		t.Fatalf("活动影片处理方式异常: action=%q snapshot=%+v", action, snapshot)
	}
	if mediaSnapshotPath(snapshot) != oldPath || !strings.Contains(notice, "将替换旧版本") || !strings.Contains(notice, removedDirName) {
		t.Fatalf("活动影片替换提醒异常: path=%q notice=%q", mediaSnapshotPath(snapshot), notice)
	}
}

func TestValidatePreparedMediaIngestStateRejectsPlanStateChangeBeforeMovingSource(t *testing.T) {
	const movieJavID = "ABP-123"
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	if err := ensureRootLayout(layout); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(layout.ingestNew, movieJavID+".mp4")
	if err := os.WriteFile(sourcePath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	current := &moviex.WMedia{
		Id:         8,
		MovieJavId: movieJavID,
		FileName:   "existing.mp4",
		SourceType: consts.WMediaSourceNative,
		IsRemoved:  consts.FilmIsNotRemoved,
		UpdatedOn:  200,
	}
	mediaModel := &ingestWMediaModel{existing: current}
	service := NewService(&dep.Dep{WMediaModel: mediaModel})

	_, err := service.validatePreparedMediaIngestState(context.Background(), &ingestPreparedItem{
		sourcePath: sourcePath,
		meta:       rawMovieMeta{movieName: movieJavID, ext: ".mp4"},
		movieInfo:  movieInfo{javID: movieJavID},
		action:     ingestActionInsert,
	})
	if err == nil || !strings.Contains(err.Error(), "状态已变化") {
		t.Fatalf("计划状态变化校验结果 = %v", err)
	}
	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("待入库源文件未保留: %v", err)
	}
	if mediaModel.insertCount != 0 || mediaModel.updateCount != 0 {
		t.Fatalf("计划状态变化仍修改数据: insert=%d update=%d", mediaModel.insertCount, mediaModel.updateCount)
	}
}

func TestInsertMediaNeverUpdatesOnDuplicate(t *testing.T) {
	duplicateErr := errors.New("Error 1062: Duplicate entry 'ABP-123' for key 'idx_w_media_movie_jav_id_source_type'")
	mediaModel := &ingestWMediaModel{insertErr: duplicateErr}
	service := NewService(&dep.Dep{WMediaModel: mediaModel})

	err := service.insertMedia(context.Background(), &moviex.WMedia{MovieJavId: "ABP-123"})
	if err == nil || !strings.Contains(err.Error(), "拒绝覆盖") {
		t.Fatalf("唯一键冲突错误 = %v", err)
	}
	if mediaModel.insertCount != 1 || mediaModel.updateCount != 0 {
		t.Fatalf("insert=%d update=%d", mediaModel.insertCount, mediaModel.updateCount)
	}
}

func TestInspectMediaIngestActionAllowsRemovedMediaRestore(t *testing.T) {
	mediaModel := &ingestWMediaModel{existing: &moviex.WMedia{
		Id:         9,
		MovieJavId: "ABP-123",
		FileName:   "removed.mp4",
		SourceType: consts.WMediaSourceNative,
		IsRemoved:  consts.FilmIsRemoved,
	}}
	service := NewService(&dep.Dep{WMediaModel: mediaModel})

	action, snapshot, notice, err := service.inspectMediaIngestAction(context.Background(), "ABP-123")
	if err != nil {
		t.Fatalf("已移除影片恢复检查失败: %v", err)
	}
	if action != ingestActionRestore || snapshot == nil || !strings.Contains(notice, "恢复为活动状态") {
		t.Fatalf("已移除影片恢复结果异常: action=%q snapshot=%+v notice=%q", action, snapshot, notice)
	}
}

func TestPersistMediaIngestReplacesRemovedMediaWithNewMetadata(t *testing.T) {
	existing := &moviex.WMedia{
		Id:                72,
		MovieJavId:        "ABP-123",
		MovieName:         "ABP-123",
		FileName:          "old.mp4",
		SourceType:        consts.WMediaSourceNative,
		SourceTorrentHash: "OLD_HASH",
		DirectoryId:       3,
		RootDir:           "/old-root",
		FullDir:           "/old-root/media/old",
		Alias:             "old-alias",
		Size:              100,
		Width:             640,
		Height:            360,
		BitRate:           800,
		Duration:          1000,
		FrameAverage:      24,
		HasSub:            consts.FilmNoSub,
		SelfMake:          consts.FilmNoSelfMake,
		HasMask:           consts.FilmNotErased,
		NeedScanMeta:      consts.FilmMetaDataNeedScan,
		IsRemoved:         consts.FilmIsRemoved,
		RemoveTime:        500,
		BirthTime:         600,
		ReleasingDate:     700,
		CreatedOn:         800,
		UpdatedOn:         900,
	}
	mediaModel := &ingestWMediaModel{existing: existing}
	service := NewService(&dep.Dep{
		WMediaModel:  mediaModel,
		GScStatModel: &ingestGScStatModel{},
	})
	newRow := &moviex.WMedia{
		MovieJavId:        "ABP-123",
		MovieName:         "ABP-123",
		FileName:          "new_sub_self.mp4",
		SourceType:        consts.WMediaSourceNative,
		SourceTorrentHash: "NEW_HASH",
		DirectoryId:       13,
		RootDir:           "/new-root",
		FullDir:           "/new-root/media/new",
		Alias:             "new-alias",
		Size:              200,
		Width:             1920,
		Height:            1080,
		BitRate:           4800,
		Duration:          2000,
		FrameAverage:      60,
		HasSub:            consts.FilmHasSub,
		SelfMake:          consts.FilmSelfMake,
		HasMask:           consts.FilmNoMosaic,
		NeedScanMeta:      consts.FilmMetaDataNoNeedScan,
		IsRemoved:         consts.FilmIsNotRemoved,
		RemoveTime:        0,
		BirthTime:         1600,
		ReleasingDate:     1700,
		CreatedOn:         1800,
		UpdatedOn:         1900,
	}
	want := *newRow
	want.Id = existing.Id
	want.CreatedOn = existing.CreatedOn

	previous, err := service.persistMediaIngest(context.Background(), newRow, snapshotExistingMedia(existing))
	if err != nil {
		t.Fatalf("已移除影片新版本入库失败: %v", err)
	}
	if previous == nil || previous.Id != existing.Id {
		t.Fatalf("旧媒体快照 = %+v", previous)
	}
	if mediaModel.insertCount != 0 || mediaModel.updateCount != 1 {
		t.Fatalf("insert=%d update=%d", mediaModel.insertCount, mediaModel.updateCount)
	}
	if mediaModel.updated == nil || !reflect.DeepEqual(*mediaModel.updated, want) {
		t.Fatalf("已移除影片更新结果 = %+v，期望 %+v", mediaModel.updated, want)
	}
}

func TestPersistPreparedMediaIngestMovesActiveOldFileAndUpdatesMetadata(t *testing.T) {
	rootDir := t.TempDir()
	oldDir := filepath.Join(rootDir, mediaDirName, "old")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(oldDir, "ABP-123.mp4")
	if err := os.WriteFile(oldPath, []byte("old-video"), 0o644); err != nil {
		t.Fatal(err)
	}
	existing := &moviex.WMedia{
		Id:                31,
		MovieJavId:        "ABP-123",
		MovieName:         "ABP-123",
		FileName:          filepath.Base(oldPath),
		SourceType:        consts.WMediaSourceNative,
		SourceTorrentHash: "OLD_HASH",
		RootDir:           rootDir,
		FullDir:           oldDir,
		IsRemoved:         consts.FilmIsNotRemoved,
		CreatedOn:         100,
		UpdatedOn:         200,
	}
	mediaModel := &ingestWMediaModel{existing: existing}
	service := NewService(&dep.Dep{WMediaModel: mediaModel})
	prepared := &ingestPreparedItem{
		movieInfo: movieInfo{javID: existing.MovieJavId},
		action:    ingestActionReplace,
		existing:  snapshotExistingMedia(existing),
	}
	newRow := &moviex.WMedia{
		MovieJavId:        existing.MovieJavId,
		MovieName:         existing.MovieName,
		FileName:          "ABP-123_sub.mp4",
		SourceType:        consts.WMediaSourceNative,
		SourceTorrentHash: "NEW_HASH",
		RootDir:           rootDir,
		FullDir:           filepath.Join(rootDir, mediaDirName, "new"),
		Size:              999,
		Width:             1920,
		Height:            1080,
		BitRate:           6000,
		Duration:          3600,
		FrameAverage:      60,
		IsRemoved:         consts.FilmIsNotRemoved,
		RemoveTime:        0,
		CreatedOn:         300,
		UpdatedOn:         400,
	}

	previous, removedPath, err := service.persistPreparedMediaIngest(context.Background(), prepared, existing, newRow)
	if err != nil {
		t.Fatalf("活动影片替换失败: %v", err)
	}
	if previous == nil || removedPath == "" || filepath.Dir(removedPath) != buildRemovedDir(rootDir) {
		t.Fatalf("活动影片替换结果异常: previous=%+v removed=%q", previous, removedPath)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("旧文件仍位于原路径: %v", err)
	}
	if payload, err := os.ReadFile(removedPath); err != nil || string(payload) != "old-video" {
		t.Fatalf("removed 旧文件异常: payload=%q err=%v", payload, err)
	}
	want := *newRow
	want.Id = existing.Id
	want.CreatedOn = existing.CreatedOn
	if mediaModel.updated == nil || !reflect.DeepEqual(*mediaModel.updated, want) {
		t.Fatalf("活动影片更新结果 = %+v，期望 %+v", mediaModel.updated, want)
	}
}

func TestPersistPreparedMediaIngestRollsBackOldFileWhenUpdateFails(t *testing.T) {
	rootDir := t.TempDir()
	oldDir := filepath.Join(rootDir, mediaDirName, "old")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(oldDir, "ABP-123.mp4")
	if err := os.WriteFile(oldPath, []byte("old-video"), 0o644); err != nil {
		t.Fatal(err)
	}
	existing := &moviex.WMedia{
		Id:         41,
		MovieJavId: "ABP-123",
		MovieName:  "ABP-123",
		FileName:   filepath.Base(oldPath),
		SourceType: consts.WMediaSourceNative,
		RootDir:    rootDir,
		FullDir:    oldDir,
		IsRemoved:  consts.FilmIsNotRemoved,
		UpdatedOn:  200,
	}
	mediaModel := &ingestWMediaModel{existing: existing, updateErr: errors.New("update failed")}
	service := NewService(&dep.Dep{WMediaModel: mediaModel})
	prepared := &ingestPreparedItem{
		movieInfo: movieInfo{javID: existing.MovieJavId},
		action:    ingestActionReplace,
		existing:  snapshotExistingMedia(existing),
	}

	_, removedPath, err := service.persistPreparedMediaIngest(context.Background(), prepared, existing, &moviex.WMedia{
		MovieJavId: existing.MovieJavId,
		MovieName:  existing.MovieName,
		FileName:   "new.mp4",
		SourceType: consts.WMediaSourceNative,
	})
	if err == nil || !strings.Contains(err.Error(), "update failed") {
		t.Fatalf("数据库更新失败结果 = %v", err)
	}
	if removedPath != "" {
		t.Fatalf("回滚成功后仍返回 removed 路径: %q", removedPath)
	}
	if payload, statErr := os.ReadFile(oldPath); statErr != nil || string(payload) != "old-video" {
		t.Fatalf("旧文件未恢复: payload=%q err=%v", payload, statErr)
	}
	entries, readErr := os.ReadDir(buildRemovedDir(rootDir))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("回滚后 removed 目录仍有文件: %v", entries)
	}
}

func TestIngestPrecheckPlanPreservesActiveReplacementWarning(t *testing.T) {
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	if err := ensureRootLayout(layout); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(layout.ingestNew, "ABP-123.mp4")
	if err := os.WriteFile(sourcePath, []byte("new-video"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot := &ingestExistingMediaSnapshot{
		ID:                51,
		IsRemoved:         consts.FilmIsNotRemoved,
		RootDir:           rootDir,
		FullDir:           filepath.Join(rootDir, mediaDirName, "old"),
		FileName:          "ABP-123.mp4",
		SourceTorrentHash: "OLD_HASH",
		UpdatedOn:         300,
	}
	prepared := &ingestPreparedItem{
		sourcePath: sourcePath,
		meta: rawMovieMeta{
			movieName: "ABP-123",
			ext:       ".mp4",
		},
		movieInfo: movieInfo{javID: "ABP-123", releasingDay: 123},
		favoriteSource: &favoriteAlbumSourceInfo{
			favoriteAlbumID: 61,
			infoHash:        "NEW_HASH",
			item: &moviex.TmAlbumItem{
				Id:          71,
				SourceType:  "sukebei",
				SourceRowId: 81,
			},
		},
		action:   ingestActionReplace,
		notice:   "活动影片已存在，将替换旧版本",
		existing: snapshot,
	}
	plan := buildIngestPrecheckPlan(layout, []*ingestPreparedItem{prepared}, []*IngestFileItem{{
		Status:       ingestItemStatusPass,
		Action:       ingestActionReplace,
		Notice:       prepared.notice,
		RootDir:      rootDir,
		SourcePath:   sourcePath,
		MovieName:    "ABP-123",
		ExistingPath: mediaSnapshotPath(snapshot),
	}})
	if err := (&Service{}).saveIngestPrecheckPlan(layout, plan); err != nil {
		t.Fatal(err)
	}
	loaded, err := (&Service{}).loadIngestPrecheckPlan(layout)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != ingestPrecheckPlanVersion || len(loaded.Entries) != 1 || len(loaded.Checks) != 1 {
		t.Fatalf("替换计划内容异常: %+v", loaded)
	}
	restored, err := buildPreparedFromPlanEntry(loaded.Entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if restored.action != ingestActionReplace || restored.notice != prepared.notice || !reflect.DeepEqual(restored.existing, snapshot) {
		t.Fatalf("替换计划恢复结果异常: %+v", restored)
	}
	if loaded.Checks[0].ExistingPath != mediaSnapshotPath(snapshot) || loaded.Checks[0].Notice != prepared.notice {
		t.Fatalf("替换提醒未写入计划检查项: %+v", loaded.Checks[0])
	}
}
