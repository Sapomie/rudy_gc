package taskctx

import (
	"context"
	"time"
)

type Progress struct {
	Stage             string
	Message           string
	HandledCount      int
	SuccessCount      int
	FailedCount       int
	QueuedCount       int
	CurrentPhaseKey   string
	PhaseKey          string
	PhaseHandledCount int
	PhaseTotalCount   int
	PhaseSuccessCount int
	PhaseFailedCount  int
}

type Log struct {
	Level   string
	Message string
	Line    string
	At      int64
}

type pauseWaiterKey struct{}
type progressReporterKey struct{}
type logReporterKey struct{}

type pauseWaiter func(context.Context) error
type progressReporter func(Progress)
type logReporter func(Log)

func WithPauseWaiter(ctx context.Context, fn func(context.Context) error) context.Context {
	if fn == nil {
		return ctx
	}
	return context.WithValue(ctx, pauseWaiterKey{}, pauseWaiter(fn))
}

func WaitIfPaused(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	raw := ctx.Value(pauseWaiterKey{})
	if raw == nil {
		return nil
	}
	fn, ok := raw.(pauseWaiter)
	if !ok || fn == nil {
		return nil
	}
	return fn(ctx)
}

func WithProgressReporter(ctx context.Context, fn func(Progress)) context.Context {
	if fn == nil {
		return ctx
	}
	return context.WithValue(ctx, progressReporterKey{}, progressReporter(fn))
}

func ReportProgress(ctx context.Context, progress Progress) {
	if ctx == nil {
		return
	}
	raw := ctx.Value(progressReporterKey{})
	if raw == nil {
		return
	}
	fn, ok := raw.(progressReporter)
	if !ok || fn == nil {
		return
	}
	fn(progress)
}

func WithLogReporter(ctx context.Context, fn func(Log)) context.Context {
	if fn == nil {
		return ctx
	}
	return context.WithValue(ctx, logReporterKey{}, logReporter(fn))
}

func ReportLog(ctx context.Context, event Log) {
	if ctx == nil {
		return
	}
	raw := ctx.Value(logReporterKey{})
	if raw == nil {
		return
	}
	fn, ok := raw.(logReporter)
	if !ok || fn == nil {
		return
	}
	fn(event)
}

func Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
