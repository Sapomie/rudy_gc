package logic

import (
	"context"
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type CrawlLogic struct {
	ctx       context.Context
	deps      *svc.Deps
	castIdMap map[int64]struct{}
	movieSvc  *movie.Service
}

func NewCrawlLogic(ctx context.Context, deps *svc.Deps) *CrawlLogic {
	return &CrawlLogic{
		ctx:       ctx,
		deps:      deps,
		castIdMap: map[int64]struct{}{},
		movieSvc:  movie.NewMovieService(deps),
	}
}

func (l *CrawlLogic) DailyBestProcession() error {
	var err error
	err = l.FetchAndParseDailyBestinv()
	if err != nil {
		return err
	}
	_, err = l.FetchAndParseDetails()
	if err != nil {
		return err
	}
	err = l.ProcessBestinvRank()
	if err != nil {
		return err
	}
	err = l.DownLoadAllPicture()
	if err != nil {
		return err
	}
	err = l.TranslateTitle()
	if err != nil {
		return err
	}
	return nil
}
