package loop

import (
	"context"
	"sync"
	"time"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/fetchqueue"
	"rudy_gc/local/api/internal/service/fetchsehuatang"
	"rudy_gc/local/api/internal/service/fetchsite"
	"rudy_gc/local/api/internal/service/movie"
	"rudy_gc/local/api/internal/service/sc"
	"rudy_gc/local/api/internal/service/spider"
)

const (
	defaultDetailBaseWindow = 10 * time.Second
	defaultDetailMaxBatch   = 100
)

type FetchLoopService struct {
	deps              *dep.Dep
	crawlLogic        *spider.Service
	fetchQueue        *fetchqueue.Service
	fetchSiteSvc      *fetchsite.Service
	fetchSehuatangSvc *fetchsehuatang.Service
	movieSvc          *movie.Service
	scSvc             *sc.Service

	jobs       *managedProgressJobManager
	detailLogs *detailLoopLogHub

	rootMu  sync.Mutex
	rootCtx context.Context
	started bool

	exclusiveMu            sync.Mutex
	exclusiveGroups        map[string]exclusiveTaskSlot
	refreshOldestJobID     string
	refreshOldestAutoPause bool

	detailMu          sync.Mutex
	detailCtx         context.Context
	detailCancel      context.CancelFunc
	detailDone        chan struct{}
	detailPaused      bool
	detailPauseOwners map[string]struct{}
	detailStartedAt   int64
	detailBaseWindow  time.Duration
	detailMaxBatch    int
}

var sharedFetchLoopService struct {
	mu  sync.Mutex
	svc *FetchLoopService
}

func NewFetchLoopService(deps *dep.Dep) *FetchLoopService {
	sharedFetchLoopService.mu.Lock()
	defer sharedFetchLoopService.mu.Unlock()
	if sharedFetchLoopService.svc != nil {
		return sharedFetchLoopService.svc
	}
	sharedFetchLoopService.svc = &FetchLoopService{
		deps:              deps,
		crawlLogic:        spider.NewService(deps),
		fetchQueue:        fetchqueue.NewService(deps),
		fetchSiteSvc:      fetchsite.NewService(deps),
		fetchSehuatangSvc: fetchsehuatang.NewService(deps),
		movieSvc:          movie.NewService(deps),
		scSvc:             sc.NewService(deps),
		jobs:              newManagedProgressJobManager("crawler", 64, pushManagedProgressEvent),
		detailLogs:        newDetailLoopLogHub(64),
		exclusiveGroups:   make(map[string]exclusiveTaskSlot),
		detailBaseWindow:  defaultDetailBaseWindow,
		detailMaxBatch:    defaultDetailMaxBatch,
	}
	return sharedFetchLoopService.svc
}
