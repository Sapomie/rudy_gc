package media

import (
	"context"
	"fmt"
	"path/filepath"

	"rudy_gc/internal/taskctx"
)

type ingestPreparedItem struct {
	sourcePath     string
	meta           rawMovieMeta
	movieInfo      movieInfo
	favoriteSource *favoriteAlbumSourceInfo
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
		},
	}
}
