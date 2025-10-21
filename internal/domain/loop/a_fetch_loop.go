package loop

import (
	"context"
	"fmt"
	"rudy_gc/internal/contracts"
	"strings"
	"time"
)

// package loop

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
			case contracts.ProcDailyBest, contracts.ProcSeeds, contracts.ProcSeedByName: // ✅ 加入新类型
				procCtx, ok := l.tryBeginProcess(ctx)
				if !ok {
					log.Warn("ProcessionTriggerLoop: still running, skip trigger")
					continue
				}

				go func(m contracts.TriggerMsg) {
					defer l.endProcess()

					start := time.Now()
					log.Info("ProcessionTriggerLoop: stopping Detail loop...")
					l.StopDetailLoop()

					var err error
					switch m.Kind {
					case contracts.ProcDailyBest:
						log.Info("ProcessionTriggerLoop: running CrawlDailyBestProcession")
						err = l.crawlLogic.CrawlDailyBestProcession(procCtx)

					case contracts.ProcSeeds:
						log.Info("ProcessionTriggerLoop: running CrawlBySeedsActiveProcession")
						err = l.crawlLogic.CrawlBySeedsActiveProcession(procCtx)

					case contracts.ProcSeedByName: // ✅ 新增分支
						if strings.TrimSpace(m.Name) == "" {
							err = fmt.Errorf("empty seed name")
							break
						}
						log.Infof("ProcessionTriggerLoop: running CrawlBySeedName(%s)", m.Name)
						err = l.crawlLogic.CrawlBySeedName(procCtx, m.Name)
					}

					if err != nil {
						log.Errorf("ProcessionTriggerLoop: procession failed: %v", err)
					} else {
						log.Infof("ProcessionTriggerLoop: done in %v", time.Since(start))
					}

					log.Info("ProcessionTriggerLoop: restarting Detail loop...")
					l.StartDetailLoop(ctx, 10*time.Second, 100)
				}(msg)

			default:
				log.Warnf("ProcessionTriggerLoop: unknown kind=%d, skip", msg.Kind)
			}
		}
	}
}
