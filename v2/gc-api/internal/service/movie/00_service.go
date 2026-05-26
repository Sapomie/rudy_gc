package movie

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
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

func isNotFound(err error) bool {
	return errors.Is(err, sqlc.ErrNotFound) || errors.Is(err, modelx.ErrNotFound)
}
