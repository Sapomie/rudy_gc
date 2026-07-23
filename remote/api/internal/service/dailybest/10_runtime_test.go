package dailybest

import (
	"context"
	"testing"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

type startupRepository struct {
	Repository
	RuntimeRepository
	events []dailybesttypes.RuntimeEventDraft
}

func (r *startupRepository) InterruptActiveFetches(context.Context, int64, string) (int64, int64, error) {
	return 2, 3, nil
}

func (r *startupRepository) InterruptRunningTargetProbes(context.Context, int64, string) (int64, error) {
	return 1, nil
}

func (r *startupRepository) CreateRuntimeEvent(_ context.Context, draft dailybesttypes.RuntimeEventDraft) (dailybesttypes.RuntimeEvent, error) {
	r.events = append(r.events, draft)
	return dailybesttypes.RuntimeEvent{}, nil
}

func (r *startupRepository) FindLatestTargetProbe(context.Context) (dailybesttypes.TargetProbe, error) {
	return dailybesttypes.TargetProbe{}, ErrNotFound
}

func TestPrepareRuntimeInterruptsStaleFetchState(t *testing.T) {
	repository := &startupRepository{}
	service := New(context.Background(), repository, nil, time.UTC)
	if err := service.PrepareRuntime(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(repository.events) != 2 {
		t.Fatalf("events=%+v", repository.events)
	}
	if repository.events[0].EventType != "fetch_interrupted" || repository.events[1].EventType != "target_probe_interrupted" {
		t.Fatalf("events=%+v", repository.events)
	}
}
