// internal/domain/loop/a_film_trigger_loop.go
package loop

import (
	"context"
	"time"

	"rudy_gc/internal/contracts"
)

func (l *FetchLoopService) FilmTriggerLoop(ctx context.Context, trigger <-chan contracts.FilmTriggerMsg) {
	log := l.deps.Log.WithContext(ctx)
	log.Info("FilmTriggerLoop: started")
	defer log.Info("FilmTriggerLoop: stopped")

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-trigger:
			if !ok {
				return
			}

			// 尝试登记一次“触发式大流程”（互斥执行）
			procCtx, ok := l.tryBeginProcess(ctx)
			if !ok {
				log.Warn("FilmTriggerLoop: still running, skip trigger")
				continue
			}

			go func(m contracts.FilmTriggerMsg) {
				defer l.endProcess()

				start := time.Now()

				// ✅ 暂停详情抓取（避免资源竞争）
				log.Info("FilmTriggerLoop: stopping Detail loop...")
				l.StopDetailLoop()

				var err error
				switch m.Kind {
				case contracts.ProcFilmRename:
					log.Info("FilmTriggerLoop: running RenameFilm()")
					err = l.filmSvc.RenameFilm(procCtx)

				case contracts.ProcFilmProcess:
					log.Info("FilmTriggerLoop: running ProcessFilm(ctx)")
					err = l.filmSvc.ProcessFilm(procCtx)
				default:
					log.Warnf("FilmTriggerLoop: unknown kind=%d, skip", m.Kind)
				}

				if err != nil {
					log.Errorf("FilmTriggerLoop: task failed: %v", err)
				} else {
					log.Infof("FilmTriggerLoop: done in %v", time.Since(start))
				}

				// ✅ 恢复详情抓取
				log.Info("FilmTriggerLoop: restarting Detail loop...")
				l.StartDetailLoop(ctx, 10*time.Second, 100)
			}(msg)
		}
	}
}
