package dailybest

import (
	"context"
	"errors"

	contracts "rudy_gc/shared/dailybest"
)

var ErrBlankPage = errors.New("dailybest blank page")

const (
	RunStatusQueued    = "queued"
	RunStatusRunning   = "running"
	RunStatusSucceeded = "succeeded"
	RunStatusFailed    = "failed"
	RunStatusNoChange  = "no_change"

	TriggerScheduled = "scheduled"
	TriggerManual    = "manual"
)

type SnapshotDraft struct {
	BusinessDate string
	Category     int64
	Pages        []contracts.DailyBestSnapshotPage
	Items        []contracts.DailyBestSnapshotItem
	Checksum     string
}

type PublishResult struct {
	Snapshot contracts.DailyBestSnapshotResponse
	Created  bool
}

type PageSource interface {
	FetchPage(ctx context.Context, page int64) (contracts.DailyBestSnapshotPage, error)
}
