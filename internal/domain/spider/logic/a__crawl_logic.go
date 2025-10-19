package logic

import (
	"context"
	"errors"
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
	"sync"
	"time"
)

type CrawlLogic struct {
	deps      *svc.Deps
	castIdMap map[int64]struct{}
	movieSvc  *movie.Service
}

func NewCrawlLogic(deps *svc.Deps) *CrawlLogic {
	return &CrawlLogic{
		deps:      deps,
		castIdMap: map[int64]struct{}{},
		movieSvc:  movie.NewMovieService(deps),
	}
}

func (l *CrawlLogic) CrawlDailyBestProcession(ctx context.Context) error {
	start := time.Now()
	if err := l.FetchAndParseDailyBestinv(ctx); err != nil {
		return err
	}

	detailNum, err := l.FetchAndParseDetails(ctx)
	if err != nil {
		return err
	}
	end := time.Now()

	// ✅ 抽出独立函数
	l.saveRecord(ctx, "Best", start, end, detailNum)

	var wg sync.WaitGroup
	wg.Add(3)

	errs := make([]error, 0, 3)
	var mu sync.Mutex

	run := func(f func(context.Context) error) {
		defer wg.Done()
		if err := f(ctx); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	}

	go run(l.ProcessBestinvRank)
	go run(l.DownLoadAllPicture)
	go run(l.TranslateTitle)

	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...) // Go 1.20+
	}
	return nil
}

func (l *CrawlLogic) CrawlBySeedsProcession(ctx context.Context) error {
	start := time.Now()

	// ===== 串行部分 =====
	if err := l.FetchAndParseInventoryBySeed(ctx); err != nil {
		return err
	}

	var (
		detailNum int64
		err       error
	)

	if detailNum, err = l.FetchAndParseDetails(ctx); err != nil {
		return err
	}

	end := time.Now()
	// ✅ 记录本次 Seeds 流程
	l.saveRecord(ctx, "Seeds", start, end, detailNum)

	// ===== 并行部分（DownLoadAllPicture + TranslateTitle）=====
	var wg sync.WaitGroup
	wg.Add(2)

	var mu sync.Mutex
	errs := make([]error, 0, 2)

	run := func(fn func(context.Context) error) {
		defer wg.Done()
		if err := fn(ctx); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	}

	go run(l.DownLoadAllPicture)
	go run(l.TranslateTitle)

	wg.Wait()

	// ===== 汇总错误 =====
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
