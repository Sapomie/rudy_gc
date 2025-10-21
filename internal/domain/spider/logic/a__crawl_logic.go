package logic

import (
	"context"
	"errors"
	"sync"
	"time"

	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type CrawlLogic struct {
	deps      *svc.Deps
	castIdMap map[int64]struct{}
	movieSvc  *movie.MovieService
}

func NewCrawlLogic(deps *svc.Deps) *CrawlLogic {
	return &CrawlLogic{
		deps:      deps,
		castIdMap: map[int64]struct{}{},
		movieSvc:  movie.NewMovieService(deps),
	}
}

/* ========= 具体流程 ========= */

// 每日榜流程：Best → FetchDetails → 并行（ProcessBestinvRank / DownLoadAllPicture / TranslateTitle）
func (l *CrawlLogic) CrawlDailyBestProcession(ctx context.Context) error {
	return l.runPipeline(
		ctx,
		"DailyBest",
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseDailyBestinv(ctx); err != nil {
				return 0, err
			}
			return l.FetchAndParseDetails(ctx)
		},
		l.ProcessBestinvRank,
		l.DownLoadAllPicture,
		l.TranslateTitle,
	)
}

// 活跃 Seeds 流程：SeedsActive → FetchDetails → 并行（DownLoadAllPicture / TranslateTitle）
func (l *CrawlLogic) CrawlBySeedsActiveProcession(ctx context.Context) error {
	return l.runPipeline(
		ctx,
		"SeedsActive",
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseInventoryBySeedActive(ctx); err != nil {
				return 0, err
			}
			return l.FetchAndParseDetails(ctx)
		},
		l.DownLoadAllPicture,
		l.TranslateTitle,
	)
}

// 指定 Seed 名称流程：Seed(name) → FetchDetails → 并行（DownLoadAllPicture / TranslateTitle）
func (l *CrawlLogic) CrawlBySeedName(ctx context.Context, name string) error {
	return l.runPipeline(
		ctx,
		"SeedName",
		func(ctx context.Context) (int64, error) {
			if err := l.FetchAndParseInventoryBySeedName(ctx, name); err != nil {
				return 0, err
			}
			return l.FetchAndParseDetails(ctx)
		},
		l.DownLoadAllPicture,
		l.TranslateTitle,
	)
}

/* ========= 通用小工具 ========= */

// 并行执行多个步骤，收集错误并返回 errors.Join(...)。
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
		go func() {
			defer wg.Done()
			if err := fn(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// 通用管线：记录开始/结束时间，执行“前置抓取”得到 detailNum → 保存 record → 并行执行后续动作。
func (l *CrawlLogic) runPipeline(
	ctx context.Context,
	recordType string, // 记录类型：e.g. "Best"/"Seeds"
	pre func(context.Context) (detailNum int64, err error), // 前置抓取逻辑（各流程自定义）
	post ...func(context.Context) error, // 并行后置动作
) error {
	start := time.Now()

	detailNum, err := pre(ctx)
	if err != nil {
		return err
	}

	end := time.Now()
	l.saveRecord(ctx, recordType, start, end, detailNum)

	// 并行后置动作（可为空）
	return l.runParallel(ctx, post...)
}
