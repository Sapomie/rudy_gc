package loop

import (
	"context"
	"rudy_gc/internal/spider/logic"
	"rudy_gc/internal/spider/types"
	"sync/atomic"
	"time"
)

func (m *LoopServer) CrawlJavInvLoop() {
	m.logInfo("CrawlJavInvLoop started")
	defer m.logInfo("CrawlJavInvLoop stopped")

	for {
		select {
		case <-m.ctx.Done():
			return
		case n := <-m.InvCh:
			if n == nil {
				continue
			}
			m.handleCrawlInv(n)
		}
	}
}

func (m *LoopServer) handleCrawlInv(n *types.Notification) {
	// 信号量：限制并发
	select {
	case m.refInvSemaphore <- struct{}{}:
		// ok
	case <-m.ctx.Done():
		return
	}
	defer func() { <-m.refInvSemaphore }()

	// 运行互斥：正在跑则合并触发
	if !m.startInv() {
		m.logWarn("inventory run skipped: already running", "info", n.Action)
		return
	}
	go func() {
		defer m.stopInv()
		m.executeCrawl(n)
		m.notifyCrawlDetailIfNecessary(n)
	}()
}

func (m *LoopServer) startInv() bool {
	return atomic.CompareAndSwapInt32(&m.goingOnInv, 0, 1)
}

func (m *LoopServer) stopInv() {
	atomic.StoreInt32(&m.goingOnInv, 0)
}

func (m *LoopServer) executeCrawl(n *types.Notification) {
	// 每次执行都带超时/trace ctx（这里先简化）
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Minute)
	defer cancel()

	l := logic.NewCrawlLogic(ctx, m.deps) // 这里的 deps 就是你 main 里准备好的依赖聚合

	var err error
	switch n.Action {
	case types.ActionActiveQueries:
		err = l.CrawlActiveQueries()
	case types.ActionDailyBestinv:
		err = l.CrawlDailyBestinv()
	case types.ActionSyncDailyBestinv:
		err = l.SyncDailyBestinv()
	default:
		m.logWarn("inventory: unknown action", "action", n.Action)
	}

	if err != nil {
		m.logErr(err, "phase", "inventory", "info", n.Action)
	} else {
		m.logInfo("inventory: done", "info", n.Action)
	}
}

func (m *LoopServer) notifyCrawlDetailIfNecessary(n *types.Notification) {
	select {
	case m.DetailCh <- n:
		m.logInfo("detail notified")
	case <-time.After(100 * time.Millisecond):
		// 不阻塞：避免上游被拖慢；这里也可以做“最新值覆盖”的coalescing策略
		m.logWarn("detail notify dropped (channel full)")
	}
}
