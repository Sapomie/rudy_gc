package dailybest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"rudy_gc/remote/api/internal/notify"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type recordingNotifier struct {
	messages []notify.Message
}

func (n *recordingNotifier) Notify(_ context.Context, message notify.Message) error {
	n.messages = append(n.messages, message)
	return nil
}

func TestNotifyFetchResultRules(t *testing.T) {
	fetchErr := errors.New("source request failed")
	tests := []struct {
		name        string
		triggerType string
		collectErr  error
		wantCount   int
		wantSubject string
	}{
		{name: "scheduled failure", triggerType: dailybesttypes.TriggerScheduled, collectErr: fetchErr, wantCount: 1, wantSubject: "抓取失败"},
		{name: "recovery failure deduplicated", triggerType: dailybesttypes.TriggerRecovery, collectErr: fetchErr, wantCount: 0},
		{name: "manual failure ignored", triggerType: dailybesttypes.TriggerManual, collectErr: fetchErr, wantCount: 0},
		{name: "scheduled success ignored", triggerType: dailybesttypes.TriggerScheduled, wantCount: 0},
		{name: "recovery success", triggerType: dailybesttypes.TriggerRecovery, wantCount: 1, wantSubject: "已恢复"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := &recordingNotifier{}
			service := New(
				context.Background(), nil, nil, time.UTC,
				WithNotifier(recorder, "https://remote.example.com"),
			)
			service.notifyFetchResult(
				context.Background(), 42, "2026-07-28", contracts.DailyBestCategoryMonth,
				test.triggerType, time.Date(2026, 7, 29, 8, 31, 0, 0, time.UTC).Unix(), test.collectErr,
			)
			if len(recorder.messages) != test.wantCount {
				t.Fatalf("messages=%+v", recorder.messages)
			}
			if test.wantCount == 0 {
				return
			}
			message := recorder.messages[0]
			if !strings.Contains(message.Subject, test.wantSubject) {
				t.Fatalf("subject=%q", message.Subject)
			}
			for _, expected := range []string{
				"任务 ID：42",
				"业务日期：2026-07-28",
				"榜单分类：月榜（1）",
				"https://remote.example.com/dailybestfetchrundetail/42",
			} {
				if !strings.Contains(message.Body, expected) {
					t.Fatalf("body missing %q: %q", expected, message.Body)
				}
			}
		})
	}
}

type notificationExecuteRepository struct {
	Repository
	failed bool
}

func (r *notificationExecuteRepository) MarkFetchRunRunning(context.Context, int64, int64) error {
	return nil
}

func (r *notificationExecuteRepository) MarkFetchRunFailed(context.Context, int64, int64, string) error {
	r.failed = true
	return nil
}

func TestExecuteScheduledFailureSendsNotification(t *testing.T) {
	repository := &notificationExecuteRepository{}
	recorder := &recordingNotifier{}
	service := New(
		context.Background(), repository, failingRecoverySource{}, time.UTC,
		WithNotifier(recorder, "https://remote.example.com"),
	)
	service.execute(
		context.Background(), 43, "2026-07-28", contracts.DailyBestCategoryMonth,
		dailybesttypes.TriggerScheduled,
	)
	if !repository.failed {
		t.Fatal("抓取任务未标记失败")
	}
	if len(recorder.messages) != 1 || !strings.Contains(recorder.messages[0].Body, "test fetch failed") {
		t.Fatalf("messages=%+v", recorder.messages)
	}
}
