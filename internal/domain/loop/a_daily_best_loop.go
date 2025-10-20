package loop

import (
	"context"
	"rudy_gc/internal/contracts"
	"time"
)

// ProcessionTriggerLoop
// 监听 trigger 通道，根据触发类型执行相应流程（DailyBest / Seeds）
// 不再支持 StopCurrentProcess。
func (l *FetchLoopService) ProcessionTriggerLoop(ctx context.Context, trigger <-chan contracts.TriggerMsg) {
	log := l.deps.Log.WithContext(ctx)
	log.Info("ProcessionTriggerLoop: started")
	defer log.Info("ProcessionTriggerLoop: stopped")

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-trigger:
			if !ok {
				return
			}

			switch msg.Kind {
			case contracts.ProcDailyBest, contracts.ProcSeeds:
				// 尝试登记一个新流程（已在跑就跳过）
				procCtx, ok := l.tryBeginProcess(ctx)
				if !ok {
					log.Warn("ProcessionTriggerLoop: still running, skip trigger")
					continue
				}

				go func(m contracts.TriggerMsg) {
					defer l.endProcess()

					start := time.Now()

					// 暂停详情抓取
					log.Info("ProcessionTriggerLoop: stopping Detail loop...")
					l.StopDetailLoop()

					var err error
					switch m.Kind {
					case contracts.ProcDailyBest:
						log.Info("ProcessionTriggerLoop: running CrawlDailyBestProcession")
						err = l.crawlLogic.CrawlDailyBestProcession(procCtx)

					case contracts.ProcSeeds:
						log.Info("ProcessionTriggerLoop: running CrawlBySeedsProcession")
						err = l.crawlLogic.CrawlBySeedsProcession(procCtx)
					}

					if err != nil {
						log.Errorf("ProcessionTriggerLoop: procession failed: %v", err)
					} else {
						log.Infof("ProcessionTriggerLoop: done in %v", time.Since(start))
					}

					// 恢复详情抓取
					log.Info("ProcessionTriggerLoop: restarting Detail loop...")
					l.StartDetailLoop(ctx, 10*time.Second, 100)
				}(msg)

			default:
				log.Warnf("ProcessionTriggerLoop: unknown kind=%d, skip", msg.Kind)
			}
		}
	}
}
