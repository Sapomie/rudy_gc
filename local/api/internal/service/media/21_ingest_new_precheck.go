package media

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/taskctx"
)

type ingestPreparedItem struct {
	sourcePath     string
	meta           rawMovieMeta
	movieInfo      movieInfo
	favoriteSource *favoriteAlbumSourceInfo
	action         string
	notice         string
	existing       *ingestExistingMediaSnapshot
}

type ingestExistingMediaSnapshot struct {
	ID                int64  `json:"id"`
	IsRemoved         int64  `json:"is_removed"`
	RootDir           string `json:"root_dir"`
	FullDir           string `json:"full_dir"`
	FileName          string `json:"file_name"`
	SourceTorrentHash string `json:"source_torrent_hash"`
	UpdatedOn         int64  `json:"updated_on"`
}

type ingestPrecheckDecision struct {
	prepared   *ingestPreparedItem
	failedItem *IngestFileItem
}

func (s *Service) precheckIngestFiles(ctx context.Context, layout rootLayout, files []string) ([]*ingestPrecheckDecision, error) {
	decisions := make([]*ingestPrecheckDecision, 0, len(files))
	precheckPassed := 0
	precheckFailed := 0

	for idx, sourcePath := range files {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return decisions, err
		}
		select {
		case <-ctx.Done():
			return decisions, ctx.Err()
		default:
		}

		decision := s.precheckOneIngestFile(ctx, layout, sourcePath)
		decisions = append(decisions, decision)

		fileName := filepath.Base(sourcePath)
		if decision != nil && decision.prepared != nil {
			precheckPassed++
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:        "media_precheck_item_pass",
				Message:      fmt.Sprintf("预处理通过：%s", fileName),
				HandledCount: idx + 1,
				SuccessCount: precheckPassed,
				FailedCount:  precheckFailed,
				QueuedCount:  len(files) - (idx + 1),
			})
			continue
		}

		precheckFailed++
		errMessage := "unknown precheck error"
		if decision != nil && decision.failedItem != nil && decision.failedItem.Error != "" {
			errMessage = decision.failedItem.Error
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:        "media_precheck_item_failed",
			Message:      fmt.Sprintf("预处理失败：%s，err=%s", fileName, errMessage),
			HandledCount: idx + 1,
			SuccessCount: precheckPassed,
			FailedCount:  precheckFailed,
			QueuedCount:  len(files) - (idx + 1),
		})
	}
	return decisions, nil
}

func (s *Service) precheckOneIngestFile(ctx context.Context, layout rootLayout, sourcePath string) *ingestPrecheckDecision {
	item := &IngestFileItem{
		RootDir:    layout.rootDir,
		SourcePath: sourcePath,
	}

	rawName := filepath.Base(sourcePath)
	meta, err := parseRawMovieMeta(rawName)
	if err != nil {
		item.Error = fmt.Sprintf("precheck invalid movie name: %v", err)
		return &ingestPrecheckDecision{failedItem: item}
	}
	item.MovieName = meta.movieName

	movieInfo, err := s.findMovieInfoByName(ctx, meta.movieName)
	if err != nil {
		item.Error = fmt.Sprintf("precheck movie lookup failed: %v", err)
		return &ingestPrecheckDecision{failedItem: item}
	}
	action, existing, notice, err := s.inspectMediaIngestAction(ctx, movieInfo.javID)
	if err != nil {
		item.Error = fmt.Sprintf("precheck media action failed: %v", err)
		return &ingestPrecheckDecision{failedItem: item}
	}

	favoriteSource, found, err := s.findFavoriteAlbumSourceInfo(ctx, movieInfo.javID)
	if err != nil {
		item.Error = fmt.Sprintf("precheck favorite lookup failed: %v", err)
		return &ingestPrecheckDecision{failedItem: item}
	}
	if !found || favoriteSource == nil || favoriteSource.item == nil {
		item.Error = fmt.Sprintf("precheck favorite item missing or info_hash empty: %s", movieInfo.javID)
		return &ingestPrecheckDecision{failedItem: item}
	}

	return &ingestPrecheckDecision{
		prepared: &ingestPreparedItem{
			sourcePath:     sourcePath,
			meta:           meta,
			movieInfo:      movieInfo,
			favoriteSource: favoriteSource,
			action:         action,
			notice:         notice,
			existing:       existing,
		},
	}
}

func (s *Service) inspectMediaIngestAction(ctx context.Context, movieJavID string) (string, *ingestExistingMediaSnapshot, string, error) {
	existing, err := s.findExistingMediaForIngest(ctx, movieJavID)
	if err != nil {
		return "", nil, "", err
	}
	if existing == nil {
		return ingestActionInsert, nil, "", nil
	}

	snapshot := snapshotExistingMedia(existing)
	if existing.IsRemoved == consts.FilmIsRemoved {
		return ingestActionRestore, snapshot, "影片记录已移除，将使用新版本完整数据恢复为活动状态", nil
	}

	existingPath := mediaSnapshotPath(snapshot)
	if existingPath == "" {
		return "", nil, "", fmt.Errorf("活动影片旧文件路径为空: %s", movieJavID)
	}
	info, err := os.Stat(existingPath)
	if err != nil {
		return "", nil, "", fmt.Errorf("活动影片旧文件不可用: %s: %w", existingPath, err)
	}
	if !info.Mode().IsRegular() {
		return "", nil, "", fmt.Errorf("活动影片旧路径不是文件: %s", existingPath)
	}

	removedDir := buildRemovedDir(snapshot.RootDir)
	if removedDir == "" {
		return "", nil, "", fmt.Errorf("活动影片 root_dir 为空: %s", movieJavID)
	}
	notice := fmt.Sprintf("活动影片已存在，将替换旧版本；旧文件将移动到 %s", removedDir)
	return ingestActionReplace, snapshot, notice, nil
}

func snapshotExistingMedia(row *moviex.WMedia) *ingestExistingMediaSnapshot {
	if row == nil {
		return nil
	}
	return &ingestExistingMediaSnapshot{
		ID:                row.Id,
		IsRemoved:         row.IsRemoved,
		RootDir:           strings.TrimSpace(row.RootDir),
		FullDir:           strings.TrimSpace(row.FullDir),
		FileName:          strings.TrimSpace(row.FileName),
		SourceTorrentHash: strings.TrimSpace(row.SourceTorrentHash),
		UpdatedOn:         row.UpdatedOn,
	}
}

func mediaSnapshotPath(snapshot *ingestExistingMediaSnapshot) string {
	if snapshot == nil {
		return ""
	}
	return joinMediaPath(snapshot.FullDir, snapshot.FileName)
}

func mediaSnapshotMatches(snapshot *ingestExistingMediaSnapshot, row *moviex.WMedia) bool {
	if snapshot == nil || row == nil {
		return snapshot == nil && row == nil
	}
	return snapshot.ID == row.Id &&
		snapshot.IsRemoved == row.IsRemoved &&
		filepath.Clean(snapshot.RootDir) == filepath.Clean(strings.TrimSpace(row.RootDir)) &&
		filepath.Clean(snapshot.FullDir) == filepath.Clean(strings.TrimSpace(row.FullDir)) &&
		snapshot.FileName == strings.TrimSpace(row.FileName) &&
		snapshot.SourceTorrentHash == strings.TrimSpace(row.SourceTorrentHash) &&
		snapshot.UpdatedOn == row.UpdatedOn
}
