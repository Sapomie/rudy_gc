package loop

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"rudy_gc/internal/taskctx"
)

func (l *FetchLoopService) StartDetailLoop(parent context.Context, baseWindow time.Duration, maxBatch int) {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()
	l.startDetailLoopLocked(parent, baseWindow, maxBatch)
}

func (l *FetchLoopService) startDetailLoopLocked(parent context.Context, baseWindow time.Duration, maxBatch int) {
	if parent == nil {
		parent = context.Background()
	}
	if baseWindow <= 0 {
		baseWindow = defaultDetailBaseWindow
	}
	if maxBatch <= 0 {
		maxBatch = defaultDetailMaxBatch
	}
	l.detailBaseWindow = baseWindow
	l.detailMaxBatch = maxBatch

	if l.detailCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})
	l.detailCtx = ctx
	l.detailCancel = cancel
	l.detailDone = done
	l.detailPaused = false

	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				l.deps.Log.WithContext(ctx).Errorf("DetailFetchLoopSingle panic: %v\n%s", r, debug.Stack())
			}
			l.detailMu.Lock()
			if l.detailCtx == ctx {
				l.detailCtx = nil
				l.detailCancel = nil
				l.detailDone = nil
				l.detailPaused = false
			}
			l.detailMu.Unlock()
		}()

		l.DetailFetchLoopSingle(ctx, l.deps.DetailJobs, baseWindow, maxBatch)
	}()

	l.deps.Log.WithContext(parent).Info("StartDetailLoop: 启动详情抓取 loop")
}

func (l *FetchLoopService) StopDetailLoop() {
	l.stopDetailLoop()
}

func (l *FetchLoopService) stopDetailLoop() {
	l.detailMu.Lock()
	cancel := l.detailCancel
	done := l.detailDone
	if cancel == nil {
		l.detailMu.Unlock()
		return
	}
	l.detailCancel = nil
	l.detailMu.Unlock()

	l.deps.Log.Info("StopDetailLoop: 正在停止详情抓取 loop ...")
	cancel()
	if done != nil {
		<-done
	}
	l.deps.Log.Info("StopDetailLoop: 详情抓取 loop 已停止")
}

func (l *FetchLoopService) StopDetailLoopAsync() {
	go l.stopDetailLoop()
}

func (l *FetchLoopService) IsDetailLoopRunning() bool {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()
	return l.detailCancel != nil
}

func (l *FetchLoopService) RestartDetailLoop(parent context.Context, baseWindow time.Duration, maxBatch int) {
	l.stopDetailLoop()
	l.StartDetailLoop(parent, baseWindow, maxBatch)
}

func (l *FetchLoopService) pauseDetailLoop() {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()
	l.detailPaused = true
}

func (l *FetchLoopService) resumeDetailLoop() {
	l.detailMu.Lock()
	defer l.detailMu.Unlock()
	l.detailPaused = false
}

func (l *FetchLoopService) waitDetailLoopActive(ctx context.Context) error {
	for {
		l.detailMu.Lock()
		paused := l.detailPaused
		l.detailMu.Unlock()
		if !paused {
			return nil
		}
		if err := taskctx.Sleep(ctx, 200*time.Millisecond); err != nil {
			return err
		}
	}
}

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
			if err := l.waitDetailLoopActive(ctx); err != nil {
				l.flushBatch(ctx, &buf)
				return
			}

			id = strings.TrimSpace(id)
			if tryAddID(&buf, id, seen, ttl) && len(buf) >= maxBatch {
				log.Infof("DetailFetchLoopSingle: 达到批次上限(%d)，立即 flush", maxBatch)
				l.flushBatch(ctx, &buf)
				window = baseWindow / 2
				resetTimer(timer, window)
			}
		case <-timer.C:
			if err := l.waitDetailLoopActive(ctx); err != nil {
				l.flushBatch(ctx, &buf)
				return
			}
			beforeFlush := len(buf)
			if beforeFlush > 0 {
				log.Infof("DetailFetchLoopSingle: 定时 flush，缓冲数量=%d", beforeFlush)
			}
			l.flushBatch(ctx, &buf)

			if beforeFlush == 0 {
				window = minDuration(baseWindow*2, 5*time.Second)
			} else {
				window = baseWindow
			}
			resetTimer(timer, window)
		}
	}
}

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
