// internal/domain/loop/a_sc_trigger_loop.go
package loop

import (
	"context"
	"strings"
	"time"

	"rudy_gc/internal/contracts"
)

func (l *FetchLoopService) ScTriggerLoop(ctx context.Context, ch <-chan contracts.ScTriggerMsg) {
	log := l.deps.Log.WithContext(ctx)
	log.Info("ScTriggerLoop: started")
	defer log.Info("ScTriggerLoop: stopped")

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-ch:
			if !ok {
				return
			}
			// 串行处理：一个消息一个 goroutine 也行，这里直接串行，保证顺序与资源占用可控
			start := time.Now()
			switch msg.Kind {
			case contracts.ScMove:
				scName := strings.TrimSpace(msg.ScName)
				if scName == "" {
					log.Warn("ScTriggerLoop: empty scName, skip")
					continue
				}
				log.Infof("ScTriggerLoop: MoveScFilm(scName=%s) begin", scName)
				if err := l.scSvc.MoveScFilm(ctx, scName); err != nil {
					log.Errorf("ScTriggerLoop: MoveScFilm failed: %v", err)
				} else {
					log.Infof("ScTriggerLoop: MoveScFilm ok in %v", time.Since(start))
				}

			case contracts.ScAdd:
				dir := strings.TrimSpace(msg.Dir)
				if dir == "" {
					log.Warn("ScTriggerLoop: empty dir, skip")
					continue
				}
				log.Infof("ScTriggerLoop: AddSc(dir=%s) begin", dir)
				if err := l.scSvc.AddSc(ctx, dir); err != nil {
					log.Errorf("ScTriggerLoop: AddSc failed: %v", err)
				} else {
					log.Infof("ScTriggerLoop: AddSc ok in %v", time.Since(start))
				}

			default:
				log.Warnf("ScTriggerLoop: unknown kind=%d, skip", msg.Kind)
			}
		}
	}
}
