package spider

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/threading"

	"rudy_gc/local/api/internal/consts"
	dailybestsync "rudy_gc/local/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
)

func (l *Service) CrawlDailyBestProcession(ctx context.Context, isSync bool, autoFetchSite bool) (err error) {
	return l.crawlDailyBestProcession(ctx, isSync, autoFetchSite, nil)
}

func (l *Service) CrawlDailyBestSyncProcession(
	ctx context.Context,
	targets []dailybesttypes.SyncTarget,
	autoFetchSite bool,
) error {
	return l.crawlDailyBestProcession(ctx, true, autoFetchSite, targets)
}

func (l *Service) crawlDailyBestProcession(
	ctx context.Context,
	isSync bool,
	autoFetchSite bool,
	targets []dailybesttypes.SyncTarget,
) (err error) {
	start := time.Now()
	var affected *affectedMovieNumbers
	var syncResult *dailybestsync.Result
	processCompleted := false
	defer func() {
		if isSync && !processCompleted && err != nil && l.dailyBestSync != nil {
			l.dailyBestSync.MarkFailed(ctx, syncResult, err)
		}
	}()
	flowKey := "spider_daily_best"
	if isSync {
		flowKey = "spider_daily_best_sync"
	}

	detailNum, err := func(ctx context.Context) (int64, error) {
		result, err := l.FetchAndParseDailyBestinv(ctx, isSync, targets)
		if err != nil {
			return 0, err
		}
		syncResult = result
		if isSync && (syncResult == nil || len(syncResult.Batches) == 0) {
			affected = newAffectedMovieNumbers()
			return 0, nil
		}
		detailNum, parsedAffected, err := l.FetchAndParseDetails(ctx)
		if err != nil {
			return 0, err
		}
		affected = parsedAffected
		l.rebuildAffectedMovieAggs(flowKey, affected)
		return detailNum, nil
	}(ctx)
	if err != nil {
		return err
	}
	if isSync && (syncResult == nil || len(syncResult.Batches) == 0) {
		processCompleted = true
		return nil
	}

	end := time.Now()
	if !isSync {
		l.saveRecord(ctx, consts.RecordTypeDailyBest, start, end, detailNum)
	}

	var affectedRankTimings []dailybesttypes.RankDayTiming
	if isSync {
		if affectedRankTimings, err = l.ProcessDailyBestSyncRanks(ctx, syncResult.Batches); err != nil {
			return err
		}
	} else if affectedRankTimings, err = l.ProcessBestinvRank(ctx); err != nil {
		return err
	}
	affectedRankDays := rankDayNumbers(affectedRankTimings)
	if len(affectedRankDays) > 0 {
		if err = l.movieSvc.CreateRankDayStats(ctx, affectedRankTimings); err != nil {
			return err
		}
		if err = l.movieSvc.RefreshRankPeriodStats(ctx, affectedRankDays); err != nil {
			return err
		}
		if err = l.movieSvc.RefreshRankPeriods(ctx, affectedRankDays); err != nil {
			return err
		}
	}
	if isSync {
		if err = l.dailyBestSync.MarkProcessed(ctx, syncResult); err != nil {
			return err
		}
	}
	processCompleted = true

	postSteps := []func(context.Context) error{
		l.DownLoadAllPicture,
		l.TranslateTitle,
	}
	if autoFetchSite {
		postSteps = append(postSteps, func(ctx context.Context) error {
			return l.runFetchSiteAfterDetail(ctx, affected, true)
		})
	}

	return l.runParallel(ctx, postSteps...)
}

func rankDayNumbers(timings []dailybesttypes.RankDayTiming) []int64 {
	dayNumbers := make([]int64, 0, len(timings))
	for _, timing := range timings {
		if timing.DayNumber > 0 {
			dayNumbers = append(dayNumbers, timing.DayNumber)
		}
	}
	return dayNumbers
}

// 活跃 Seeds 流程
func (l *Service) CrawlBySeedsActiveProcession(ctx context.Context, autoFetchSite bool) (err error) {
	var affected *affectedMovieNumbers
	postSteps := []func(context.Context) error{
		l.DownLoadAllPicture,
		l.TranslateTitle,
	}
	if autoFetchSite {
		postSteps = append(postSteps, func(ctx context.Context) error {
			return l.runFetchSiteAfterDetail(ctx, affected, false)
		})
	}

	err = l.runPipeline(
		ctx,
		consts.RecordTypeSeedsActive,
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseInventoryBySeedActive(ctx); err != nil {
				return 0, err
			}
			detailNum, parsedAffected, err := l.FetchAndParseDetails(ctx)
			if err != nil {
				return 0, err
			}
			affected = parsedAffected
			l.rebuildAffectedMovieAggs("spider_seeds", affected)
			return detailNum, nil
		},
		false, // ✅ 保持记录
		postSteps...,
	)
	return err
}

// 指定 Seed 名称流程
func (l *Service) CrawlBySeedName(ctx context.Context, name string, autoFetchSite bool) (err error) {
	var affected *affectedMovieNumbers
	postSteps := []func(context.Context) error{
		l.DownLoadAllPicture,
		l.TranslateTitle,
	}
	if autoFetchSite {
		postSteps = append(postSteps, func(ctx context.Context) error {
			return l.runFetchSiteAfterDetail(ctx, affected, false)
		})
	}

	err = l.runPipeline(
		ctx,
		consts.RecordTypeSeedName,
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseInventoryBySeedName(ctx, name); err != nil {
				return 0, err
			}
			detailNum, parsedAffected, err := l.FetchAndParseDetails(ctx)
			if err != nil {
				return 0, err
			}
			affected = parsedAffected
			l.rebuildAffectedMovieAggs("spider_seed_by_name", affected)
			return detailNum, nil
		},
		false, // ✅ 保持记录
		postSteps...,
	)
	return err
}

func (l *Service) RunPostProcess(ctx context.Context, withCover bool, withTranslate bool, translateStatuses []int64) error {
	postSteps := make([]func(context.Context) error, 0, 2)
	if withCover {
		postSteps = append(postSteps, l.DownLoadAllPicture)
	}
	if withTranslate {
		postSteps = append(postSteps, func(ctx context.Context) error {
			return l.TranslateTitleWithStatuses(ctx, translateStatuses)
		})
	}
	return l.runParallel(ctx, postSteps...)
}

// 并行执行多个步骤，收集错误并返回 errors.Join(...)
func (l *Service) runParallel(ctx context.Context, fns ...func(context.Context) error) error {
	if len(fns) == 0 {
		return nil
	}
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	wg.Add(len(fns))
	for _, fn := range fns {
		fn := fn
		threading.GoSafe(func() {
			defer wg.Done()
			if err := fn(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// 通用管线：记录开始/结束时间 → 执行前置抓取 →（可选保存 record）→ 并行执行后续动作。
func (l *Service) runPipeline(
	ctx context.Context,
	recordType string, // 记录类型
	pre func(context.Context) (detailNum int64, err error), // 前置抓取逻辑
	isSync bool, // ✅ 是否同步
	post ...func(context.Context) error, // 后置动作
) error {
	start := time.Now()

	detailNum, err := pre(ctx)
	if err != nil {
		return err
	}

	end := time.Now()

	// 仅当非同步模式才保存 record
	if !isSync {
		l.saveRecord(ctx, recordType, start, end, detailNum)
	}

	return l.runParallel(ctx, post...)
}

func (l *Service) rebuildAffectedMovieAggs(flowKey string, affected *affectedMovieNumbers) {
	if affected == nil || len(affected.movies) == 0 {
		return
	}

	javIDs := make([]string, 0, len(affected.movies))
	for javID := range affected.movies {
		if javID == "" {
			continue
		}
		javIDs = append(javIDs, javID)
	}
	l.movieSvc.EnqueueAggRebuildByMovieJavIDs(flowKey, javIDs...)
}
