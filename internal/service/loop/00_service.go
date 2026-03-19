package loop

import (
	"context"
	"sync"
	"time"

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
	deps       *svc.Deps
	crawlLogic *spider.CrawlLogic
	movieSvc   *movie.Service
	filmSvc    *vfilm.FilmService
	scSvc      *sc.ScService

	jobs       *managedProgressJobManager
	detailLogs *detailLoopLogHub

	rootMu  sync.Mutex
	rootCtx context.Context
	started bool

	exclusiveMu       sync.Mutex
	exclusiveJobID    string
	exclusiveTaskType string

	detailMu         sync.Mutex
	detailCtx        context.Context
	detailCancel     context.CancelFunc
	detailDone       chan struct{}
	detailPaused     bool
	detailStartedAt  int64
	detailBaseWindow time.Duration
	detailMaxBatch   int
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
		deps:             deps,
		crawlLogic:       spider.NewCrawlLogic(deps),
		movieSvc:         movie.NewService(deps),
		filmSvc:          vfilm.NewFilmService(deps),
		scSvc:            sc.NewService(deps),
		jobs:             newManagedProgressJobManager("crawler", 64, pushManagedProgressEvent),
		detailLogs:       newDetailLoopLogHub(64),
		detailBaseWindow: defaultDetailBaseWindow,
		detailMaxBatch:   defaultDetailMaxBatch,
	}
	return sharedFetchLoopService.svc
}
