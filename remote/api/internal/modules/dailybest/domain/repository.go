package dailybest

import (
	"context"
	contracts "rudy_gc/shared/dailybest"
)

type Repository interface {
	CreateFetchRun(ctx context.Context, businessDate string, category int64, triggerType string) (contracts.DailyBestFetchRun, error)
	MarkFetchRunRunning(ctx context.Context, runID int64, startedAt int64) error
	MarkFetchRunFailed(ctx context.Context, runID int64, finishedAt int64, message string) error
	PublishSnapshot(ctx context.Context, runID int64, draft SnapshotDraft, finishedAt int64) (PublishResult, error)
	ListSnapshotsAfter(ctx context.Context, revision int64) ([]contracts.DailyBestSnapshotManifest, error)
	FindSnapshot(ctx context.Context, revision int64, includePages bool) (contracts.DailyBestSnapshotResponse, error)
	FindLatestSnapshotByDate(ctx context.Context, businessDate string, category int64, includePages bool) (contracts.DailyBestSnapshotResponse, error)
	ListFetchRuns(ctx context.Context) ([]contracts.DailyBestFetchRun, error)
	FindFetchRun(ctx context.Context, runID int64) (contracts.DailyBestFetchRun, error)
	Ping(ctx context.Context) error
}
