package media

import (
	"context"
	"errors"
	"time"

	"rudy_gc/internal/taskctx"
)

const (
	processDirName       = "001_process"
	ingestNewDirName     = "001_ingest_new"
	tmpDirName           = "002_tmp"
	failedDirName        = "003_failed"
	rollbackDirName      = "004_rollback"
	mediaDirName         = "media"
	watchedDirName       = "watched"
	maxFilesPerLeafDir   = 18
	maxLeafDirsPerYear   = 20
	defaultFilePerm      = 0o755
	defaultSourceHashLen = 40
	minMediaFileSize     = 100 * 1024 * 1024 // 100MB，对齐旧 film process
)

const (
	ingestItemStatusPass = "pass"
	ingestItemStatusFail = "fail"
)

type IngestNewResult struct {
	Roots    int                `json:"roots"`
	Total    int                `json:"total"`
	Success  int                `json:"success"`
	Failed   int                `json:"failed"`
	Precheck IngestPrecheckStat `json:"precheck"`
	Items    []*IngestFileItem  `json:"items"`
}

type IngestPrecheckStat struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

type IngestFileItem struct {
	Status            string `json:"status"`
	RootDir           string `json:"root_dir"`
	SourcePath        string `json:"source_path"`
	MovieName         string `json:"movie_name"`
	TargetFileName    string `json:"target_file_name"`
	TargetDir         string `json:"target_dir"`
	Alias             string `json:"alias"`
	SourceTorrentHash string `json:"source_torrent_hash"`
	Size              int64  `json:"size"`
	BirthTime         int64  `json:"birth_time"`
	TargetPath        string `json:"target_path"`
	FailedPath        string `json:"failed_path"`
	Error             string `json:"error"`
}

func (s *Service) IngestNew(ctx context.Context) (*IngestNewResult, error) {
	// 兼容旧入口：media_ingest_new 默认执行第一段预处理。
	return s.IngestPrecheck(ctx)
}

func (s *Service) IngestPrecheck(ctx context.Context) (*IngestNewResult, error) {
	roots := s.mediaRoots()
	result := &IngestNewResult{
		Roots: len(roots),
		Items: make([]*IngestFileItem, 0, 64),
	}
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
			Stage:   "media_precheck_root_begin",
			Message: "开始预处理 root: " + root,
		})

		rootResult, err := s.ingestPrecheckOneRoot(ctx, root)
		result.Precheck.Total += rootResult.Precheck.Total
		result.Precheck.Passed += rootResult.Precheck.Passed
		result.Precheck.Failed += rootResult.Precheck.Failed
		result.Total += len(rootResult.Items)
		for _, item := range rootResult.Items {
			if item.Error == "" {
				result.Success++
			} else {
				result.Failed++
			}
		}
		result.Items = append(result.Items, rootResult.Items...)
		if err != nil {
			return result, err
		}
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "media_precheck_done",
		Message:      buildPrecheckDoneMessage(result.Precheck.Total, result.Precheck.Passed, result.Precheck.Failed),
		HandledCount: result.Precheck.Total,
		SuccessCount: result.Precheck.Passed,
		FailedCount:  result.Precheck.Failed,
	})
	return result, nil
}

func (s *Service) IngestCommit(ctx context.Context) (result *IngestNewResult, err error) {
	defer func() {
		if rebuildErr := s.rebuildMediaAggsAfterFlow(ctx, "media_ingest"); rebuildErr != nil {
			if err == nil {
				err = rebuildErr
			} else {
				err = errors.Join(err, rebuildErr)
			}
		}
	}()

	roots := s.mediaRoots()
	result = &IngestNewResult{
		Roots: len(roots),
		Items: make([]*IngestFileItem, 0, 64),
	}
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
			Stage:   "media_commit_root_begin",
			Message: "开始执行第二段插入 root: " + root,
		})

		rootResult, err := s.ingestCommitOneRoot(ctx, root, time.Now())
		result.Precheck.Total += rootResult.Precheck.Total
		result.Precheck.Passed += rootResult.Precheck.Passed
		result.Precheck.Failed += rootResult.Precheck.Failed
		result.Total += len(rootResult.Items)
		for _, item := range rootResult.Items {
			if item.Error == "" {
				result.Success++
			} else {
				result.Failed++
			}
		}
		result.Items = append(result.Items, rootResult.Items...)
		if err != nil {
			return result, err
		}
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "media_commit_done",
		Message:      "第二段插入处理完成",
		HandledCount: result.Total,
		SuccessCount: result.Success,
		FailedCount:  result.Failed,
	})
	return result, nil
}

func buildPrecheckDoneMessage(total, passed, failed int) string {
	switch {
	case total == 0:
		return "预处理完成：没有可处理文件"
	case failed == 0:
		return "预处理完成：全部通过，可以执行第二段插入操作"
	default:
		return "预处理完成：部分失败，可返回修正后重跑预处理，或执行第二段仅插入通过项"
	}
}
