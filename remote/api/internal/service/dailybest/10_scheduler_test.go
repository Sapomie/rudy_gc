package dailybest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type failureRecoveryRepository struct {
	Repository
	SchedulerRepository

	mu       sync.Mutex
	created  bool
	triggers chan string
	failed   chan struct{}
}

func (r *failureRecoveryRepository) FindAutomaticFetchState(context.Context, string, int64) (dailybesttypes.AutomaticFetchState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.created {
		active := contracts.DailyBestFetchRun{ID: 2, Status: dailybesttypes.RunStatusRunning}
		return dailybesttypes.AutomaticFetchState{Active: &active}, nil
	}
	failed := contracts.DailyBestFetchRun{
		ID: 1, TriggerType: dailybesttypes.TriggerRecovery, Status: dailybesttypes.RunStatusFailed,
		FinishedAt: time.Now().Add(-2 * time.Hour).Unix(),
	}
	return dailybesttypes.AutomaticFetchState{Failed: &failed}, nil
}

func (r *failureRecoveryRepository) CreateFetchRun(_ context.Context, businessDate string, category int64, triggerType string) (contracts.DailyBestFetchRun, error) {
	r.mu.Lock()
	r.created = true
	r.mu.Unlock()
	r.triggers <- triggerType
	return contracts.DailyBestFetchRun{
		ID: 2, BusinessDate: businessDate, Category: category,
		TriggerType: triggerType, Status: dailybesttypes.RunStatusQueued,
	}, nil
}

func (r *failureRecoveryRepository) MarkFetchRunRunning(context.Context, int64, int64) error {
	return nil
}

func (r *failureRecoveryRepository) MarkFetchRunFailed(context.Context, int64, int64, string) error {
	close(r.failed)
	return nil
}

type failingRecoverySource struct{}

func (failingRecoverySource) FetchPage(context.Context, int64, int64) (dailybesttypes.SourcePage, error) {
	return dailybesttypes.SourcePage{}, errors.New("test fetch failed")
}

func TestFailureRecoveryCreatesRecoveryRunAfterDelay(t *testing.T) {
	repository := &failureRecoveryRepository{
		triggers: make(chan string, 1),
		failed:   make(chan struct{}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	service := New(ctx, repository, failingRecoverySource{}, time.UTC)

	if err := service.StartFailureRecovery(ctx, time.Hour); err != nil {
		t.Fatal(err)
	}
	select {
	case trigger := <-repository.triggers:
		if trigger != dailybesttypes.TriggerRecovery {
			t.Fatalf("trigger=%q", trigger)
		}
	case <-time.After(time.Second):
		t.Fatal("超过恢复时间后未创建 recovery 任务")
	}
	select {
	case <-repository.failed:
	case <-time.After(time.Second):
		t.Fatal("recovery 任务未完成失败收尾")
	}
	cancel()
	service.Wait()
}
