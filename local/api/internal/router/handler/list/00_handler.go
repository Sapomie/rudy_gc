package list

import (
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/media"
	"rudy_gc/local/api/internal/service/movie"
)

type ListPagesAPI struct {
	deps     *dep.Dep
	mediaSvc *media.Service
	movieSvc *movie.Service
}

func NewListPagesAPI(deps *dep.Dep) *ListPagesAPI {
	return &ListPagesAPI{
		deps:     deps,
		mediaSvc: media.NewService(deps),
		movieSvc: movie.NewService(deps),
	}
}
