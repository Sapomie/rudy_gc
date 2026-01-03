package sc

import (
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
	"sync"
)

type ScService struct {
	deps     *svc.Deps
	movieSvc *movie.MovieService
	copyMu   sync.Mutex
	copyTask *copyTask
}

func NewScService(deps *svc.Deps) *ScService {
	return &ScService{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}
