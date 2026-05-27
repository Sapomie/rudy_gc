package cards

import (
	"time"

	"rudy-gc-api/internal/dep"
	"rudy-gc-api/internal/model/modelx"
)

const movieCardCacheTTL = 24 * time.Hour

type Service struct {
	repo  *modelx.MovieReadRepo
	cache dep.MovieTypeCache
}

func New(d *dep.Dep) *Service {
	return &Service{
		repo:  modelx.NewMovieReadRepo(d.Conn, d.Config),
		cache: d.MovieTypeCache,
	}
}
