// internal/domain/loop/a_singleid_detail_fetch_loop.go
package loop

import (
	"context"
	"runtime/debug"
	"strings"
	"time"
)

func (l *FetchLoopService) DetailFetchLoopSingle(ctx context.Context, job <-chan string, baseWindow time.Duration, maxBatch int) {
	log := l.deps.Log.WithContext(ctx)
	log.Info("DetailFetchLoopSingle: started")
	defer log.Info("DetailFetchLoopSingle: stopped")

	ttl := 10 * time.Minute
	seen := make(map[string]time.Time)
	buf := make([]string, 0, maxBatch)

	window := baseWindow
	timer := time.NewTimer(window)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			l.flushBatch(ctx, &buf)
			return

		case id, ok := <-job:
			if !ok {
				l.flushBatch(ctx, &buf)
				return
			}

			id = strings.TrimSpace(id)
			if tryAddID(&buf, id, seen, ttl) {

				if len(buf) >= maxBatch {
					log.Infof("DetailFetchLoopSingle: 达到批次上限(%d)，立即 flush", maxBatch)
					l.flushBatch(ctx, &buf)
					window = baseWindow / 2
					resetTimer(timer, window)
				}
			}

		case <-timer.C:
			if len(buf) > 0 {
				log.Infof("DetailFetchLoopSingle: 定时 flush，缓冲数量=%d", len(buf))
			}
			l.flushBatch(ctx, &buf)

			if len(buf) == 0 {
				window = minDuration(baseWindow*2, 5*time.Second)
			} else {
				window = baseWindow
			}
			resetTimer(timer, window)
		}
	}
}

// --- 新增：独立的 flush 函数 ---
func (l *FetchLoopService) flushBatch(ctx context.Context, buf *[]string) {
	defer func() {
		if r := recover(); r != nil {
			l.deps.Log.WithContext(ctx).
				Errorf("flushBatch panic: %v\n%s", r, debug.Stack())
		}
	}()
	if len(*buf) == 0 {
		return
	}
	batch := uniqueNonEmpty(*buf)
	*buf = (*buf)[:0]
	if len(batch) == 0 {
		return
	}

	log := l.deps.Log.WithContext(ctx)
	start := time.Now()
	n, err := l.crawlLogic.HandleFetchDetailsById(ctx, batch)
	if err != nil {
		log.Errorf("DetailFetchLoopSingle: batch failed (n=%d): %v", n, err)
		return
	}
	log.Infof("DetailFetchLoopSingle: processed %d ids (took %v)", len(batch), time.Since(start))
}

// --- 其余辅助函数保持不变 ---
func tryAddID(buf *[]string, id string, seen map[string]time.Time, ttl time.Duration) bool {
	if id == "" {
		return false
	}
	if t, ok := seen[id]; ok && time.Since(t) < ttl {
		return false
	}
	seen[id] = time.Now()
	*buf = append(*buf, id)
	return true
}

func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}

func uniqueNonEmpty(in []string) []string {
	m := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		if _, ok := m[s]; ok {
			continue
		}
		m[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
