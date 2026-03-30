package spider

import (
	"context"
	"fmt"
	"time"

	"rudy_gc/internal/taskctx"
	"rudy_gc/internal/types"
)

func (l *CrawlLogic) runFetchSiteAfterDetail(ctx context.Context, affected *affectedMovieNumbers, onlyReleased bool) error {
	movies, skippedUnreleased := filterFetchSiteMoviesByReleaseDate(affected.listMovieTypes(), onlyReleased)
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:           "fetch_site_after_detail_prepare",
		Message:         fmt.Sprintf("详情完成后准备外站抓取：%d 部影片，未上映跳过=%d", len(movies), skippedUnreleased),
		CurrentPhaseKey: "detail",
	})
	if len(movies) == 0 {
		return nil
	}

	javbusResult, sukebeiResult, err := l.fetchQueue.RunAfterDetailFetchTasks(ctx, movies)
	if err != nil {
		return err
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:           "fetch_site_after_detail_done",
		Message:         fmt.Sprintf("详情后外站抓取完成：JavBus %d/%d，Sukebei %d/%d", javbusResult.Success, javbusResult.Handled, sukebeiResult.Success, sukebeiResult.Handled),
		CurrentPhaseKey: "detail",
	})
	return nil
}

func filterFetchSiteMoviesByReleaseDate(movies []*types.MovieType, onlyReleased bool) ([]*types.MovieType, int) {
	if !onlyReleased {
		return movies, 0
	}
	now := time.Now().Unix()
	out := make([]*types.MovieType, 0, len(movies))
	skipped := 0
	for _, mv := range movies {
		if mv == nil || mv.AMovie == nil {
			continue
		}
		releaseDate := mv.AMovie.ReleasingDate
		if releaseDate > now {
			skipped++
			continue
		}
		out = append(out, mv)
	}
	return out, skipped
}
