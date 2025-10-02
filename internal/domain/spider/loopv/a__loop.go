package loopv

import (
	"rudy_gc/internal/svc"
	"sync"
)

type LoopServer struct {
	deps               *svc.Deps
	goingOnCrawlJavLib bool
	refDetailSemaphore chan struct{}
	mu                 sync.Mutex
}

func NewLoopServer(deps *svc.Deps) *LoopServer {
	return &LoopServer{
		deps:               deps,
		refDetailSemaphore: make(chan struct{}, 1),
	}
}

func (m *LoopServer) Loop() {
	//threading.GoSafe(m.CrawlJavInvLoop)
	//
	//threading.GoSafe(m.CrawlDetailLoop)
	//
	//threading.GoSafe(m.RefDetailByMovieLoop)
}
