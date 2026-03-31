package spider

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/fetchqueue"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/movie"
	"sync"
	"time"

	"rudy_gc/internal/svc"

	"github.com/zeromicro/go-zero/core/threading"
)

/* ========= 常量定义 ========= */

/* ========= 结构体 ========= */

type CrawlLogic struct {
	deps         *runtimeDeps
	movieSvc     *movie.Service
	fetchSiteSvc *fetchsite.Service
	fetchQueue   *fetchqueue.Service
}

func NewCrawlLogic(deps *svc.Deps) *CrawlLogic {
	return &CrawlLogic{
		deps:         newRuntimeDeps(deps),
		movieSvc:     movie.NewService(deps),
		fetchSiteSvc: fetchsite.NewService(deps),
		fetchQueue:   fetchqueue.NewService(deps),
	}
}

/* ========= 具体流程 ========= */

func (l *CrawlLogic) CrawlDailyBestProcession(ctx context.Context, isSync bool, autoFetchSite bool) error {
	start := time.Now()
	var affected *affectedMovieNumbers

	detailNum, err := func(ctx context.Context) (int64, error) {
		if err := l.FetchAndParseDailyBestinv(ctx, isSync); err != nil {
			return 0, err
		}
		detailNum, parsedAffected, err := l.FetchAndParseDetails(ctx)
		if err != nil {
			return 0, err
		}
		affected = parsedAffected
		return detailNum, nil
	}(ctx)
	if err != nil {
		return err
	}

	end := time.Now()
	if !isSync {
		l.saveRecord(ctx, consts.RecordTypeDailyBest, start, end, detailNum)
	}

	if err := l.ProcessBestinvRank(ctx); err != nil {
		return err
	}
	if err := l.movieSvc.RebuildCurrentRankPeriods(ctx); err != nil {
		return err
	}

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

// 活跃 Seeds 流程
func (l *CrawlLogic) CrawlBySeedsActiveProcession(ctx context.Context, autoFetchSite bool) error {
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

	return l.runPipeline(
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
			return detailNum, nil
		},
		false, // ✅ 保持记录
		postSteps...,
	)
}

// 指定 Seed 名称流程
func (l *CrawlLogic) CrawlBySeedName(ctx context.Context, name string, autoFetchSite bool) error {
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

	return l.runPipeline(
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
			return detailNum, nil
		},
		false, // ✅ 保持记录
		postSteps...,
	)
}

/* ========= 通用小工具 ========= */

// 并行执行多个步骤，收集错误并返回 errors.Join(...)
func (l *CrawlLogic) runParallel(ctx context.Context, fns ...func(context.Context) error) error {
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
func (l *CrawlLogic) runPipeline(
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
