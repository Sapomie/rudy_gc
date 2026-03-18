package sc

import (
	"sync"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/movie"
)

type Service struct {
	deps     *dep.Dep
	movieSvc *movie.Service
	copyMu   sync.Mutex
	copyTask *copyTask
}

type ScService = Service

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:     d,
		movieSvc: movie.NewService(d),
	}
}
