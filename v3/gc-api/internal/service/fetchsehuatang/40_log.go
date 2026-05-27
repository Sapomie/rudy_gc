package fetchsehuatang

import (
	"context"
	"time"

	"rudy_gc/internal/taskctx"
)

func reportInfoLog(ctx context.Context, message string) {
	reportTaskLog(ctx, "info", message)
}

func reportWarnLog(ctx context.Context, message string) {
	reportTaskLog(ctx, "warn", message)
}

func reportErrorLog(ctx context.Context, message string) {
	reportTaskLog(ctx, "error", message)
}

func reportTaskLog(ctx context.Context, level, message string) {
	taskctx.ReportLog(ctx, taskctx.Log{
		Level:   level,
		Message: message,
		Line:    message,
		At:      time.Now().Unix(),
	})
}
