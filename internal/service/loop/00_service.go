package loop

import (
	"context"
	"sync"
	"time"

	"rudy_gc/internal/service/fetchqueue"
	"rudy_gc/internal/service/fetchsehuatang"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/service/spider"
	"rudy_gc/internal/service/vfilm"
	"rudy_gc/internal/svc"
)

const (
	defaultDetailBaseWindow = 10 * time.Second
	defaultDetailMaxBatch   = 100
)

type FetchLoopService struct {
	deps              *svc.Deps
	crawlLogic        *spider.CrawlLogic
	fetchQueue        *fetchqueue.Service
	fetchSehuatangSvc *fetchsehuatang.Service
	movieSvc          *movie.Service
	filmSvc           *vfilm.FilmService
	scSvc             *sc.ScService

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

func NewFetchLoopService(deps *svc.Deps) *FetchLoopService {
	sharedFetchLoopService.mu.Lock()
	defer sharedFetchLoopService.mu.Unlock()
	if sharedFetchLoopService.svc != nil {
		return sharedFetchLoopService.svc
	}
	sharedFetchLoopService.svc = &FetchLoopService{
		deps:              deps,
		crawlLogic:        spider.NewCrawlLogic(deps),
		fetchQueue:        fetchqueue.NewService(deps),
		fetchSehuatangSvc: fetchsehuatang.NewService(deps),
		movieSvc:          movie.NewService(deps),
		filmSvc:           vfilm.NewFilmService(deps),
		scSvc:             sc.NewService(deps),
		jobs:              newManagedProgressJobManager("crawler", 64, pushManagedProgressEvent),
		detailLogs:        newDetailLoopLogHub(64),
		exclusiveGroups:   make(map[string]exclusiveTaskSlot),
		detailBaseWindow:  defaultDetailBaseWindow,
		detailMaxBatch:    defaultDetailMaxBatch,
	}
	return sharedFetchLoopService.svc
}
