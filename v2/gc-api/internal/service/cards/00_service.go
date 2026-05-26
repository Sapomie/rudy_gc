package cards

import (
	"rudy-gc-api/internal/dep"
	"rudy-gc-api/internal/model/modelx"
)

type Service struct {
	repo *modelx.MovieReadRepo
}

func New(d *dep.Dep) *Service {
	return &Service{
		repo: modelx.NewMovieReadRepo(d.Conn, d.Config),
	}
}
