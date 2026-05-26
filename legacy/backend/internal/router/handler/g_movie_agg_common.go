package handler

import (
	"strconv"

	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/moviereleaseagg"
	"rudy_gc/internal/svc"
)

const defaultAggPageSize = 100

type MovieAggHTMLHandler struct {
	deps          *svc.Deps
	movieSvc      *movie.Service
	releaseAggSvc *moviereleaseagg.Service
}

func NewMovieAggHTMLHandler(deps *svc.Deps) *MovieAggHTMLHandler {
	return &MovieAggHTMLHandler{
		deps:          deps,
		movieSvc:      movie.NewService(deps),
		releaseAggSvc: moviereleaseagg.NewService(deps),
	}
}

func atoiDef(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func clampPageSize(size int) int {
	if size <= 0 {
		return defaultAggPageSize
	}
	if size > maxPageSize {
		return maxPageSize
	}
	return size
}
