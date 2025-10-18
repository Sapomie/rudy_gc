package logic

import (
	"context"
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
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
	var err error

	err = l.FetchAndParseDailyBestinv(ctx)
	if err != nil {
		return err
	}

	_, err = l.FetchAndParseDetails(ctx)
	if err != nil {
		return err
	}

	err = l.ProcessBestinvRank(ctx)
	if err != nil {
		return err
	}

	err = l.DownLoadAllPicture(ctx)
	if err != nil {
		return err
	}

	err = l.TranslateTitle(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *CrawlLogic) CrawlBySeedsProcession(ctx context.Context) error {
	var err error

	err = l.FetchAndParseInventoryBySeed(ctx)
	if err != nil {
		return err
	}

	_, err = l.FetchAndParseDetails(ctx)
	if err != nil {
		return err
	}

	err = l.DownLoadAllPicture(ctx)
	if err != nil {
		return err
	}

	err = l.TranslateTitle(ctx)
	if err != nil {
		return err
	}
	return nil
}
