package wdir

import (
	"sync"

	"rudy_gc/local/api/internal/dep"
	mediaservice "rudy_gc/local/api/internal/service/media"
	"rudy_gc/local/api/internal/service/movie"
)

type Service struct {
	deps     *dep.Dep
	movieSvc *movie.Service
	mediaSvc *mediaservice.Service

	normalizeMu   sync.Mutex
	normalizedAll bool
}

type DirectoryService = Service

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:     d,
		movieSvc: movie.NewService(d),
		mediaSvc: mediaservice.NewService(d),
	}
}

func NewDirectoryService(d *dep.Dep) *DirectoryService {
	return NewService(d)
}
