package movie

import (
	"rudy_gc/internal/svc"
)

type Service struct {
	deps *svc.Deps
}

func NewMovieService(deps *svc.Deps) *Service {
	return &Service{deps}
}
