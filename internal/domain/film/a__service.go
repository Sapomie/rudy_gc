package film

import (
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type Service struct {
	deps     *svc.Deps
	movieSvc *movie.MovieService
}

func NewFilmService(deps *svc.Deps) *Service {
	return &Service{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}
