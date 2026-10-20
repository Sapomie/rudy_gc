package sc

import (
	"sync"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/movie"
	"rudy_gc/local/api/internal/service/personstat"
)

type Service struct {
	deps          *dep.Dep
	movieSvc      *movie.Service
	personStatSvc *personstat.Service
	copyMu        sync.Mutex
	copyTask      *copyTask
	copyV2Mu      sync.RWMutex
	copyV2Preview map[string][]SmartPickCopyItem
	copyV2Tasks   map[string]*copyTaskV2
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:          d,
		movieSvc:      movie.NewService(d),
		personStatSvc: personstat.NewService(d),
		copyV2Preview: make(map[string][]SmartPickCopyItem),
		copyV2Tasks:   make(map[string]*copyTaskV2),
	}
}
