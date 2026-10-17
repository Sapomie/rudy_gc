package dailybest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

const probeTargetURL = "https://www.c97k.com/cn/vl_bestrated.php?&mode=&page=1"

type targetProbeRepository struct {
	Repository
	RuntimeRepository
	SourceConfigRepository
	completed chan dailybesttypes.TargetProbeResult
}

func (r *targetProbeRepository) FindSourceConfig(context.Context) (dailybesttypes.SourceConfig, error) {
	return dailybesttypes.SourceConfig{JavAddress: "www.c97k.com"}, nil
}

func (r *targetProbeRepository) CreateTargetProbe(context.Context, dailybesttypes.TargetProbeDraft) (dailybesttypes.TargetProbe, error) {
	return dailybesttypes.TargetProbe{ID: 1, CreatedOn: time.Now().Unix()}, nil
}

func (r *targetProbeRepository) CompleteTargetProbe(_ context.Context, _ int64, result dailybesttypes.TargetProbeResult) error {
	r.completed <- result
	return nil
}

func (r *targetProbeRepository) CreateRuntimeEvent(context.Context, dailybesttypes.RuntimeEventDraft) (dailybesttypes.RuntimeEvent, error) {
	return dailybesttypes.RuntimeEvent{}, nil
}

type targetProbeAgent struct{}

func (targetProbeAgent) Request(context.Context, string, string, string) (browseragenttypes.FetchResult, int, error) {
	return browseragenttypes.FetchResult{
		StatusCode: 200, FinalURL: probeTargetURL, BrowserState: browseragenttypes.StateFetching,
		Body:      `<html><body><div id="rightcolumn"><div class="videothumblist"></div></div></body></html>`,
		StartedAt: 10, FinishedAt: 11,
	}, http.StatusOK, nil
}

func (targetProbeAgent) Status(context.Context) (browseragenttypes.Status, error) {
	return browseragenttypes.Status{}, nil
}

func TestTargetProbeSchedulerRunsImmediately(t *testing.T) {
	repository := &targetProbeRepository{completed: make(chan dailybesttypes.TargetProbeResult, 1)}
	service := New(context.Background(), repository, nil, time.UTC, WithAgent(targetProbeAgent{}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := service.StartTargetProbeScheduler(ctx, time.Hour); err != nil {
		t.Fatal(err)
	}
	select {
	case result := <-repository.completed:
		if result.Status != dailybesttypes.TargetProbeStatusSucceeded {
			t.Fatalf("result=%+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("启动后未立即执行目标站探测")
	}
	cancel()
	service.Wait()
}

func TestEvaluateTargetProbeSuccess(t *testing.T) {
	result := evaluateTargetProbe(probeTargetURL, http.StatusOK, browseragenttypes.FetchResult{
		StatusCode: 200, FinalURL: probeTargetURL, BrowserState: browseragenttypes.StateFetching,
		Body:      `<html><head><title>排行</title></head><body><div id="rightcolumn"><div class="videothumblist"></div></div></body></html>`,
		StartedAt: 10, FinishedAt: 11,
	}, nil)
	if result.Status != dailybesttypes.TargetProbeStatusSucceeded || result.ErrorMessage != "" {
		t.Fatalf("result=%+v", result)
	}
}

func TestEvaluateTargetProbeRejectsCloudflare(t *testing.T) {
	result := evaluateTargetProbe(probeTargetURL, http.StatusBadGateway, browseragenttypes.FetchResult{
		StatusCode: 403, FinalURL: probeTargetURL, ErrorType: "cloudflare", ErrorMessage: "challenge",
	}, errors.New("challenge"))
	if result.Status != dailybesttypes.TargetProbeStatusFailed || result.ErrorType != "cloudflare" {
		t.Fatalf("result=%+v", result)
	}
}

func TestEvaluateTargetProbeRejectsMissingRankList(t *testing.T) {
	result := evaluateTargetProbe(probeTargetURL, http.StatusOK, browseragenttypes.FetchResult{
		StatusCode: 200, FinalURL: probeTargetURL, Body: `<div id="rightcolumn"></div>`,
	}, nil)
	if result.Status != dailybesttypes.TargetProbeStatusFailed || result.ErrorType != "missing_rank_list" {
		t.Fatalf("result=%+v", result)
	}
}
