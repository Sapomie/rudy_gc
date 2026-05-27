package wdir

import (
	"sync"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/movie"
)

type Service struct {
	deps     *dep.Dep
	movieSvc *movie.Service

	normalizeMu   sync.Mutex
	normalizedAll bool
}

type DirectoryService = Service

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:     d,
		movieSvc: movie.NewService(d),
	}
}

func NewDirectoryService(d *dep.Dep) *DirectoryService {
	return NewService(d)
}
