// internal/domain/loop/a_daily_best_loop.go
package loop

import (
	"context"
	"errors"
	"rudy_gc/internal/contracts"
	"sync/atomic"
	"time"
)

func (l *FetchLoopService) ProcessionTriggerLoop(ctx context.Context, trigger <-chan contracts.TriggerMsg) {
	log := l.deps.Log.WithContext(ctx)
	log.Info("ProcessionTriggerLoop: started")
	defer log.Info("ProcessionTriggerLoop: stopped")

	var running atomic.Bool

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-trigger:
			if !ok {
				return
			}
			if !running.CompareAndSwap(false, true) {
				log.Warn("ProcessionTriggerLoop: still running, skip trigger")
				continue
			}

			go func(m contracts.TriggerMsg) {
				defer running.Store(false)

				start := time.Now()
				// ✅ 先暂停详情抓取
				log.Info("ProcessionTriggerLoop: stopping Detail loop...")
				l.StopDetailLoop()

				// 根据指令执行相应 procession
				var err error
				switch m.Kind {
				case contracts.ProcDailyBest:
					log.Info("ProcessionTriggerLoop: running CrawlDailyBestProcession")
					err = l.crawlLogic.CrawlDailyBestProcession(ctx)

				case contracts.ProcSeeds:
					log.Info("ProcessionTriggerLoop: running CrawlBySeedsProcession")
					err = l.crawlLogic.CrawlBySeedsProcession(ctx)

				case contracts.ProcBoth:
					log.Info("ProcessionTriggerLoop: running BOTH (DailyBest → Seeds)")
					// 这里按你需要的顺序：先每日榜再种子（也可反过来）
					if e1 := l.crawlLogic.CrawlDailyBestProcession(ctx); e1 != nil {
						err = e1
					}
					if e2 := l.crawlLogic.CrawlBySeedsProcession(ctx); e2 != nil {
						err = errors.Join(err, e2)
					}

				default:
					log.Warnf("ProcessionTriggerLoop: unknown kind=%d, skip", m.Kind)
				}

				if err != nil {
					log.Errorf("ProcessionTriggerLoop: procession failed: %v", err)
				} else {
					log.Infof("ProcessionTriggerLoop: done in %v", time.Since(start))
				}

				// ✅ 再恢复详情抓取
				log.Info("ProcessionTriggerLoop: restarting Detail loop...")
				l.StartDetailLoop(ctx, 10*time.Second, 100)
			}(msg)
		}
	}
}
