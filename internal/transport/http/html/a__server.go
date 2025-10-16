package html

import (
	"rudy_gc/internal/svc"

	"rudy_gc/internal/domain/movie"
)

type MovieHTMLHandler struct {
	svc *movie.Service
}

func NewMovieHTMLHandler(deps *svc.Deps) *MovieHTMLHandler {
	return &MovieHTMLHandler{svc: movie.NewMovieService(deps)}
}
