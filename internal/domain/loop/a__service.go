// internal/domain/loop/a__service.go
package loop

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"rudy_gc/internal/domain/spider/logic"
	"rudy_gc/internal/svc"
)

// FetchLoopService 管理后台抓取任务：
// - 详情抓取 loop（DetailFetchLoopSingle）的启动/停止/查询
// - 触发式大流程（每日榜 / seeds）的登记、取消（可被 StopCurrentProcess 取消）
type FetchLoopService struct {
	deps       *svc.Deps
	crawlLogic *logic.CrawlLogic

	// ====== 详情抓取 loop 控制 ======
	detailMu     sync.Mutex
	detailCtx    context.Context
	detailCancel context.CancelFunc
	detailWG     sync.WaitGroup

	// ====== 当前“触发式大流程”控制 ======
	procMu     sync.Mutex
	procCtx    context.Context
	procCancel context.CancelFunc
}

// NewFetchLoopService 构造函数
func NewFetchLoopService(deps *svc.Deps) *FetchLoopService {
	return &FetchLoopService{
		deps:       deps,
		crawlLogic: logic.NewCrawlLogic(deps),
	}
}

//
// ====== 一、详情抓取 loop 控制（Start / Stop / Query / Restart） ======
//

// StartDetailLoop 启动详情抓取 loop（幂等）
// parent: 一般传入顶层 ctx（例如 main 的 ctx），baseWindow/maxBatch 传给 DetailFetchLoopSingle
func (l *FetchLoopService) StartDetailLoop(parent context.Context, baseWindow time.Duration, maxBatch int) {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()

	// 已经在跑就不重复启动（幂等）
	if l.detailCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(parent)
	l.detailCtx = ctx
	l.detailCancel = cancel

	l.detailWG.Add(1)
	go func() {
		defer l.detailWG.Done()

		// 防止 goroutine panic 静默退出
		defer func() {
			if r := recover(); r != nil {
				l.deps.Log.WithContext(ctx).Errorf("DetailFetchLoopSingle panic: %v\n%s", r, debug.Stack())
			}
			// 退出时清理状态（双保险）
			l.detailMu.Lock()
			l.detailCtx = nil
			l.detailCancel = nil
			l.detailMu.Unlock()
		}()

		// 调用实际 loop（文件内另有实现 DetailFetchLoopSingle）
		l.DetailFetchLoopSingle(ctx, l.deps.DetailJobs, baseWindow, maxBatch)
	}()

	l.deps.Log.WithContext(parent).Info("StartDetailLoop: 启动详情抓取 loop")
}

// StopDetailLoop 停止详情抓取 loop（阻塞直到退出）
// 幂等：无论是否在跑都可以安全调用
func (l *FetchLoopService) StopDetailLoop() {
	l.detailMu.Lock()
	cancel := l.detailCancel
	// 置为 nil 表示正在停止 / 已停止（避免竞争）
	l.detailCancel = nil
	l.detailMu.Unlock()

	if cancel != nil {
		l.deps.Log.Info("StopDetailLoop: 正在停止详情抓取 loop ...")
		cancel()
		l.detailWG.Wait()

		// 二次清理（双保险）
		l.detailMu.Lock()
		l.detailCtx = nil
		l.detailMu.Unlock()

		l.deps.Log.Info("StopDetailLoop: 详情抓取 loop 已停止")
	}
}

// StopDetailLoopAsync 非阻塞停止（快速返回）
func (l *FetchLoopService) StopDetailLoopAsync() {
	go l.StopDetailLoop()
}

// IsDetailLoopRunning 查询详情 loop 是否在运行（仅供状态展示）
func (l *FetchLoopService) IsDetailLoopRunning() bool {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()
	return l.detailCancel != nil
}

// RestartDetailLoop 便捷重启（若在跑则先停）
func (l *FetchLoopService) RestartDetailLoop(parent context.Context, baseWindow time.Duration, maxBatch int) {
	l.StopDetailLoop()
	l.StartDetailLoop(parent, baseWindow, maxBatch)
}

//
// ====== 二、触发式大流程（DailyBest / Seeds）控制 ======
//

// tryBeginProcess 尝试登记并开始一次大流程；返回 procCtx 和 true 表示成功开始，false 表示已有流程在跑
func (l *FetchLoopService) tryBeginProcess(parent context.Context) (context.Context, bool) {
	l.procMu.Lock()
	defer l.procMu.Unlock()
	if l.procCancel != nil {
		return nil, false // 已有在跑
	}
	ctx, cancel := context.WithCancel(parent)
	l.procCtx = ctx
	l.procCancel = cancel
	return ctx, true
}

// endProcess 清理当前流程登记（在大流程 goroutine 结束时调用）
func (l *FetchLoopService) endProcess() {
	l.procMu.Lock()
	l.procCtx = nil
	l.procCancel = nil
	l.procMu.Unlock()
}

// StopCurrentProcess 外部接口：停止当前正在运行的大流程（返回是否有任务被取消）
func (l *FetchLoopService) StopCurrentProcess() bool {
	l.procMu.Lock()
	cancel := l.procCancel
	l.procMu.Unlock()

	if cancel == nil {
		l.deps.Log.Info("StopCurrentProcess: no active process")
		return false
	}
	l.deps.Log.Info("StopCurrentProcess: cancel current process ...")
	cancel()
	return true
}
