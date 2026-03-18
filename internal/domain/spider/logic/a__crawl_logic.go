package logic

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
	"sync"
	"time"

	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"

	"github.com/zeromicro/go-zero/core/threading"
)

/* ========= 常量定义 ========= */

/* ========= 结构体 ========= */

type CrawlLogic struct {
	deps     *svc.Deps
	movieSvc *movie.MovieService
}

func NewCrawlLogic(deps *svc.Deps) *CrawlLogic {
	return &CrawlLogic{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}

/* ========= 具体流程 ========= */

func (l *CrawlLogic) CrawlDailyBestProcession(ctx context.Context, isSync bool) error {
	return l.runPipeline(
		ctx,
		consts.RecordTypeDailyBest,
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseDailyBestinv(ctx, isSync); err != nil {
				return 0, err
			}
			return l.FetchAndParseDetails(ctx)
		},
		isSync, // ✅ 同步模式下不记录
		l.ProcessBestinvRank,
		l.DownLoadAllPicture,
		l.TranslateTitle,
	)
}

// 活跃 Seeds 流程
func (l *CrawlLogic) CrawlBySeedsActiveProcession(ctx context.Context) error {
	return l.runPipeline(
		ctx,
		consts.RecordTypeSeedsActive,
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseInventoryBySeedActive(ctx); err != nil {
				return 0, err
			}
			return l.FetchAndParseDetails(ctx)
		},
		false, // ✅ 保持记录
		l.DownLoadAllPicture,
		l.TranslateTitle,
	)
}

// 指定 Seed 名称流程
func (l *CrawlLogic) CrawlBySeedName(ctx context.Context, name string) error {
	return l.runPipeline(
		ctx,
		consts.RecordTypeSeedName,
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseInventoryBySeedName(ctx, name); err != nil {
				return 0, err
			}
			return l.FetchAndParseDetails(ctx)
		},
		false, // ✅ 保持记录
		l.DownLoadAllPicture,
		l.TranslateTitle,
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
