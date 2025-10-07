package film

import (
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type Service struct {
	deps     *svc.Deps
	movieSvc *movie.Service
}

func New(deps *svc.Deps) *Service {
	return &Service{
		deps:     deps,
		movieSvc: movie.New(deps),
	}
}
