package loop

import (
	"context"
	"rudy_gc/internal/domain/spider/logic"
	"sync"
	"time"

	"rudy_gc/internal/svc"
)

// FetchLoopService 负责管理各类后台循环任务（loop）
// 包含详情抓取、每日榜处理等循环任务的调度与控制
type FetchLoopService struct {
	deps       *svc.Deps
	crawlLogic *logic.CrawlLogic

	// ====== 详情抓取 loop 控制 ======
	detailMu     sync.Mutex
	detailCtx    context.Context
	detailCancel context.CancelFunc
	detailWG     sync.WaitGroup
}

// NewFetchLoopService 创建服务实例
func NewFetchLoopService(deps *svc.Deps) *FetchLoopService {
	return &FetchLoopService{
		deps:       deps,
		crawlLogic: logic.NewCrawlLogic(deps),
	}
}

//
// ====== 控制 DetailFetchLoopSingle 的生命周期 ======
//

// StartDetailLoop 启动详情抓取 loop（幂等）
func (l *FetchLoopService) StartDetailLoop(parent context.Context, baseWindow time.Duration, maxBatch int) {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()

	// 若已启动则直接返回
	if l.detailCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(parent)
	l.detailCtx = ctx
	l.detailCancel = cancel

	l.detailWG.Add(1)
	go func() {
		defer l.detailWG.Done()
		l.DetailFetchLoopSingle(ctx, l.deps.DetailJobs, baseWindow, maxBatch)
	}()
	l.deps.Log.WithContext(parent).Info("StartDetailLoop: 启动详情抓取 loop")
}

// StopDetailLoop 停止详情抓取 loop（阻塞直到退出）
func (l *FetchLoopService) StopDetailLoop() {
	l.detailMu.Lock()
	cancel := l.detailCancel
	l.detailCancel = nil
	l.detailMu.Unlock()

	if cancel != nil {
		l.deps.Log.Info("StopDetailLoop: 正在停止详情抓取 loop ...")
		cancel()
		l.detailWG.Wait()
		l.deps.Log.Info("StopDetailLoop: 详情抓取 loop 已停止")
	}
}
