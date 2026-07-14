package media

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/internal/taskctx"
)

const (
	fc2IngestPlanVersion  = 5
	fc2IngestPlanFileName = "fc2_media_ingest_precheck_plan.json"
	fc2SourceAlbumName    = "下载中"
	fc2MediaAlbumName     = "FcMedia"
	fc2AlbumSourceType    = "sh_magnet"
)

type FC2IngestResult struct {
	GeneratedAt int64                `json:"generated_at"`
	Roots       int                  `json:"roots"`
	Total       int                  `json:"total"`
	Success     int                  `json:"success"`
	Failed      int                  `json:"failed"`
	Precheck    IngestPrecheckStat   `json:"precheck"`
	Items       []*FC2IngestFileItem `json:"items"`
}

type FC2IngestFileItem struct {
	Status            string `json:"status"`
	SourceDir         string `json:"source_dir"`
	RootDir           string `json:"root_dir"`
	SourcePath        string `json:"source_path"`
	FC2Code           string `json:"fc2_code"`
	FC2Key            string `json:"fc2_key"`
	FC2MovieID        int64  `json:"fc2_movie_id"`
	FC2MediaID        int64  `json:"fc2_media_id"`
	SourceThreadID    int64  `json:"source_thread_id"`
	SourceTorrentHash string `json:"source_torrent_hash"`
	TargetFileName    string `json:"target_file_name"`
	TargetDir         string `json:"target_dir"`
	Alias             string `json:"alias"`
	Size              int64  `json:"size"`
	BirthTime         int64  `json:"birth_time"`
	TargetPath        string `json:"target_path"`
	FailedPath        string `json:"failed_path"`
	Error             string `json:"error"`
}

type fc2RawMediaMeta struct {
	code     string
	key      string
	ext      string
	hasSub   int64
	selfMake int64
}

type fc2IngestPreparedItem struct {
	root              fc2MediaRoot
	sourcePath        string
	meta              fc2RawMediaMeta
	fc2MovieID        int64
	sourceThreadID    int64
	sourceTorrentHash string
	favoriteSource    *favoriteAlbumSourceInfo
	hasMask           int64
}

type fc2MediaRoot struct {
	sourceDir string
	rootDir   string
}

type fc2IngestRootResult struct {
	GeneratedAt int64
	Precheck    IngestPrecheckStat
	Items       []*FC2IngestFileItem
}

func (s *Service) fc2MediaRoots() []fc2MediaRoot {
	raw := s.deps.Config.FC2Media.RootDirs
	out := make([]fc2MediaRoot, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, configuredRoot := range raw {
		rootDir := filepath.Clean(strings.TrimSpace(configuredRoot))
		if rootDir == "." {
			continue
		}
		if _, ok := seen[rootDir]; ok {
			continue
		}
		seen[rootDir] = struct{}{}
		out = append(out, fc2MediaRoot{sourceDir: buildRootLayout(rootDir).ingestNew, rootDir: rootDir})
	}
	return out
}

func (s *Service) FC2IngestPrecheck(ctx context.Context) (*FC2IngestResult, error) {
	roots := s.fc2MediaRoots()
	result := &FC2IngestResult{Roots: len(roots), Items: make([]*FC2IngestFileItem, 0, 64)}
	if len(roots) == 0 {
		return result, nil
	}
	for _, root := range roots {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return result, err
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "fc2_media_precheck_root_begin",
			Message: "开始预处理 root: " + root.rootDir,
		})
		rootResult, err := s.fc2IngestPrecheckRoot(ctx, root, time.Now())
		mergeFC2IngestResult(result, rootResult)
		if err != nil {
			return result, err
		}
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "fc2_media_precheck_done",
		Message:      buildPrecheckDoneMessage(result.Precheck.Total, result.Precheck.Passed, result.Precheck.Failed),
		HandledCount: result.Precheck.Total,
		SuccessCount: result.Precheck.Passed,
		FailedCount:  result.Precheck.Failed,
	})
	return result, nil
}

func (s *Service) FC2IngestCommit(ctx context.Context) (*FC2IngestResult, error) {
	roots := s.fc2MediaRoots()
	result := &FC2IngestResult{Roots: len(roots), Items: make([]*FC2IngestFileItem, 0, 64)}
	if len(roots) == 0 {
		return result, nil
	}
	for _, root := range roots {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return result, err
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "fc2_media_commit_root_begin",
			Message: "开始执行第二段插入 root: " + root.rootDir,
		})
		rootResult, err := s.fc2IngestCommitRoot(ctx, root, time.Now())
		mergeFC2IngestResult(result, rootResult)
		if err != nil {
			return result, err
		}
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "fc2_media_commit_done",
		Message:      "第二段插入处理完成",
		HandledCount: result.Total,
		SuccessCount: result.Success,
		FailedCount:  result.Failed,
	})
	return result, nil
}

func mergeFC2IngestResult(result *FC2IngestResult, pairResult *fc2IngestRootResult) {
	if result == nil || pairResult == nil {
		return
	}
	if pairResult.GeneratedAt > result.GeneratedAt {
		result.GeneratedAt = pairResult.GeneratedAt
	}
	result.Precheck.Total += pairResult.Precheck.Total
	result.Precheck.Passed += pairResult.Precheck.Passed
	result.Precheck.Failed += pairResult.Precheck.Failed
	result.Total += len(pairResult.Items)
	for _, item := range pairResult.Items {
		if item == nil {
			continue
		}
		if strings.TrimSpace(item.Error) == "" {
			result.Success++
		} else {
			result.Failed++
		}
	}
	result.Items = append(result.Items, pairResult.Items...)
}
