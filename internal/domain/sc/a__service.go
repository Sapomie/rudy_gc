package sc

import (
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type ScService struct {
	deps     *svc.Deps
	movieSvc *movie.Service
}

func NewScService(deps *svc.Deps) *ScService {
	return &ScService{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}
