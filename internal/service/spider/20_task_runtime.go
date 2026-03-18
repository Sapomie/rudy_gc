package spider

import (
	"context"
	"time"

	"rudy_gc/internal/taskctx"
)

func (l *CrawlLogic) waitIfPaused(ctx context.Context) error {
	return taskctx.WaitIfPaused(ctx)
}

func (l *CrawlLogic) reportProgress(ctx context.Context, stage, message string, handled, success, failed, queued int) {
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        stage,
		Message:      message,
		HandledCount: handled,
		SuccessCount: success,
		FailedCount:  failed,
		QueuedCount:  queued,
	})
}

func (l *CrawlLogic) sleepWithContext(ctx context.Context, d time.Duration) error {
	return taskctx.Sleep(ctx, d)
}
