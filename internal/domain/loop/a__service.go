// internal/domain/loop/a__service.go
package loop

import (
	"rudy_gc/internal/domain/spider/logic"
	"rudy_gc/internal/svc"
)

type FetchLoopService struct {
	deps       *svc.Deps
	crawlLogic *logic.CrawlLogic
}

func NewFetchLoopService(deps *svc.Deps) *FetchLoopService {
	return &FetchLoopService{
		deps:       deps,
		crawlLogic: logic.NewCrawlLogic(deps),
	}
}
