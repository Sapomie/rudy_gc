package vfilm

import (
	"sync"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/movie"
)

type Service struct {
	deps     *dep.Dep
	movieSvc *movie.Service
	mu       sync.Mutex
}

type FilmService = Service
type DirectoryService = Service

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:     d,
		movieSvc: movie.NewService(d),
	}
}

func NewFilmService(d *dep.Dep) *FilmService {
	return NewService(d)
}

func NewDirectoryService(d *dep.Dep) *DirectoryService {
	return NewService(d)
}
