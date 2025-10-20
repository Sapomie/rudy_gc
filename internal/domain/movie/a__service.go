package movie

import (
	"rudy_gc/internal/svc"
)

type MovieService struct {
	deps *svc.Deps
}

func NewMovieService(deps *svc.Deps) *MovieService {
	return &MovieService{deps}
}
