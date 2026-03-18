// internal/domain/loop/a_daily_best_loop.go
package loop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/contracts"
)

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
			case contracts.ProcDailyBest, contracts.ProcSeeds, contracts.ProcSeedByName, contracts.ProcSyncBest, contracts.ProcRefreshOldestDetail, contracts.ProcRebuildCastRank, contracts.ProcRebuildActorRank:
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
						log.Info("ProcessionTriggerLoop: running CrawlDailyBestProcession(isSync=false)")
						err = l.crawlLogic.CrawlDailyBestProcession(procCtx, false)

					case contracts.ProcSyncBest: // ✅ 新增分支
						log.Info("ProcessionTriggerLoop: running CrawlDailyBestProcession(isSync=true)")
						err = l.crawlLogic.CrawlDailyBestProcession(procCtx, true)

					case contracts.ProcSeeds:
						log.Info("ProcessionTriggerLoop: running CrawlBySeedsActiveProcession")
						err = l.crawlLogic.CrawlBySeedsActiveProcession(procCtx)

					case contracts.ProcSeedByName:
						name := strings.TrimSpace(m.Name)
						if name == "" {
							err = fmt.Errorf("empty seed name")
							break
						}
						log.Infof("ProcessionTriggerLoop: running CrawlBySeedName(%s)", name)
						err = l.crawlLogic.CrawlBySeedName(procCtx, name)

					case contracts.ProcRefreshOldestDetail: // ⭐ 新增分支
						log.Infof("ProcessionTriggerLoop: RefreshOldestDetail(num=%d)", m.Number)
						_, err = l.crawlLogic.RefreshOldestDetail(procCtx, m.Number)

					case contracts.ProcRebuildCastRank:
						log.Info("ProcessionTriggerLoop: RebuildAllCastRankStats")
						err = l.crawlLogic.RebuildAllCastRankStats(procCtx)

					case contracts.ProcRebuildActorRank:
						name := strings.TrimSpace(m.ActorName)
						if name == "" {
							err = fmt.Errorf("empty actor name")
							break
						}
						log.Infof("ProcessionTriggerLoop: RebuildCastRankStatsByName(%s)", name)
						err = l.crawlLogic.RebuildCastRankStatsByName(procCtx, name)
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
