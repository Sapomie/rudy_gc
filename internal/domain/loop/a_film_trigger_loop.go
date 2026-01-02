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

			procCtx, ok := l.tryBeginFilmProcess(ctx)
			if !ok {
				log.Warn("FilmTriggerLoop: still running, skip trigger")
				continue
			}

			go func(m contracts.FilmTriggerMsg) {
				defer l.endFilmProcess()

				start := time.Now()

				// ✅ 暂停详情抓取（避免资源竞争）
				log.Info("FilmTriggerLoop: stopping Detail loop...")

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
			}(msg)
		}
	}
}

// tryBeginProcess 登记新流程，返回 ctx 和是否成功（false 表示已有流程在跑）
func (l *FetchLoopService) tryBeginFilmProcess(parent context.Context) (context.Context, bool) {
	l.filmProcMu.Lock()
	defer l.filmProcMu.Unlock()
	if l.filmProcCtx != nil {
		return nil, false
	}
	ctx, _ := context.WithCancel(parent)
	l.filmProcCtx = ctx
	return ctx, true
}

// endProcess 清理当前流程登记
func (l *FetchLoopService) endFilmProcess() {
	l.filmProcMu.Lock()
	l.filmProcCtx = nil
	l.filmProcMu.Unlock()
}
