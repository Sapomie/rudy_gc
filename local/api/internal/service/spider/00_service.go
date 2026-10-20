package spider

import (
	"rudy_gc/local/api/internal/dep"
	dailybestsync "rudy_gc/local/api/internal/service/dailybest"
	"rudy_gc/local/api/internal/service/fetchqueue"
	"rudy_gc/local/api/internal/service/fetchsite"
	"rudy_gc/local/api/internal/service/movie"
	"rudy_gc/local/api/internal/service/personstat"
)

type Service struct {
	deps          *runtimeDeps
	movieSvc      *movie.Service
	fetchSiteSvc  *fetchsite.Service
	fetchQueue    *fetchqueue.Service
	dailyBestSync *dailybestsync.Service
	personStatSvc *personstat.Service
}

func NewService(deps *dep.Dep) *Service {
	personStatSvc := personstat.NewService(deps)
	return &Service{
		deps:          newRuntimeDeps(deps, personStatSvc),
		movieSvc:      movie.NewService(deps),
		fetchSiteSvc:  fetchsite.NewService(deps),
		fetchQueue:    fetchqueue.NewService(deps),
		dailyBestSync: deps.DailyBestSync,
		personStatSvc: personStatSvc,
	}
}
