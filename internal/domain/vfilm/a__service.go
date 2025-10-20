package vfilm

import (
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type FilmService struct {
	deps     *svc.Deps
	movieSvc *movie.MovieService
}

func NewFilmService(deps *svc.Deps) *FilmService {
	return &FilmService{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}
