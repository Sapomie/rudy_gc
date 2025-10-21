package loop

import (
	"context"
	"rudy_gc/internal/domain/sc"
	"rudy_gc/internal/domain/vfilm"
	"runtime/debug"
	"sync"
	"time"

	"rudy_gc/internal/domain/spider/logic"
	"rudy_gc/internal/svc"
)

// FetchLoopService 管理后台抓取任务：
// - 详情抓取 loop（DetailFetchLoopSingle）的启动/停止/查询
// - 触发式大流程（每日榜 / seeds）的启动登记（仅供 ProcessionTriggerLoop 使用）
type FetchLoopService struct {
	deps       *svc.Deps
	crawlLogic *logic.CrawlLogic
	filmSvc    *vfilm.FilmService //
	scSvc      *sc.ScService

	// ====== 详情抓取 loop 控制 ======
	detailMu     sync.Mutex
	detailCtx    context.Context
	detailCancel context.CancelFunc
	detailWG     sync.WaitGroup

	// ====== 当前“触发式大流程”控制（仅登记状态） ======
	procMu  sync.Mutex
	procCtx context.Context
}

// NewFetchLoopService 构造函数
func NewFetchLoopService(deps *svc.Deps) *FetchLoopService {
	return &FetchLoopService{
		deps:       deps,
		crawlLogic: logic.NewCrawlLogic(deps),
		filmSvc:    vfilm.NewFilmService(deps),
		scSvc:      sc.NewScService(deps),
	}
}

//
// ====== 一、详情抓取 loop 控制（Start / Stop / Query / Restart） ======
//

// StartDetailLoop 启动详情抓取 loop（幂等）
func (l *FetchLoopService) StartDetailLoop(parent context.Context, baseWindow time.Duration, maxBatch int) {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()

	if l.detailCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(parent)
	l.detailCtx = ctx
	l.detailCancel = cancel

	l.detailWG.Add(1)
	go func() {
		defer l.detailWG.Done()
		defer func() {
			if r := recover(); r != nil {
				l.deps.Log.WithContext(ctx).Errorf("DetailFetchLoopSingle panic: %v\n%s", r, debug.Stack())
			}
			l.detailMu.Lock()
			l.detailCtx = nil
			l.detailCancel = nil
			l.detailMu.Unlock()
		}()

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
		l.detailMu.Lock()
		l.detailCtx = nil
		l.detailMu.Unlock()
		l.deps.Log.Info("StopDetailLoop: 详情抓取 loop 已停止")
	}
}

// StopDetailLoopAsync 非阻塞停止
func (l *FetchLoopService) StopDetailLoopAsync() {
	go l.StopDetailLoop()
}

// IsDetailLoopRunning 查询 loop 是否运行中
func (l *FetchLoopService) IsDetailLoopRunning() bool {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()
	return l.detailCancel != nil
}

// RestartDetailLoop 重启 loop（先停再启）
func (l *FetchLoopService) RestartDetailLoop(parent context.Context, baseWindow time.Duration, maxBatch int) {
	l.StopDetailLoop()
	l.StartDetailLoop(parent, baseWindow, maxBatch)
}

//
// ====== 二、触发式大流程登记 ======
//

// tryBeginProcess 登记新流程，返回 ctx 和是否成功（false 表示已有流程在跑）
func (l *FetchLoopService) tryBeginProcess(parent context.Context) (context.Context, bool) {
	l.procMu.Lock()
	defer l.procMu.Unlock()
	if l.procCtx != nil {
		return nil, false
	}
	ctx, _ := context.WithCancel(parent)
	l.procCtx = ctx
	return ctx, true
}

// endProcess 清理当前流程登记
func (l *FetchLoopService) endProcess() {
	l.procMu.Lock()
	l.procCtx = nil
	l.procMu.Unlock()
}
