package dailybestsource

import (
	"context"
	"errors"
	"net/http"
	"testing"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

type recordingRequestLogger struct {
	createErr   error
	completeErr error
	drafts      []dailybesttypes.FetchLogDraft
	results     []dailybesttypes.FetchLogResult
}

func (l *recordingRequestLogger) CreateFetchLog(_ context.Context, draft dailybesttypes.FetchLogDraft) (int64, error) {
	if l.createErr != nil {
		return 0, l.createErr
	}
	l.drafts = append(l.drafts, draft)
	return int64(len(l.drafts)), nil
}

func (l *recordingRequestLogger) CompleteFetchLog(_ context.Context, _ int64, result dailybesttypes.FetchLogResult) error {
	if l.completeErr != nil {
		return l.completeErr
	}
	l.results = append(l.results, result)
	return nil
}

type fakePageClient struct {
	response Response
	err      error
	called   bool
}

func (c *fakePageClient) Get(_ context.Context, _ string) (Response, error) {
	c.called = true
	return c.response, c.err
}

func TestClientRecordsSuccessfulRequest(t *testing.T) {
	logger := &recordingRequestLogger{}
	pageClient := &fakePageClient{response: Response{StatusCode: http.StatusOK, Body: []byte("ok")}}
	client := NewClient(pageClient, logger)

	response, err := client.Get(context.Background(), 17, 3, "https://example.com/page=3")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(response.Body) != "ok" {
		t.Fatalf("response=%+v", response)
	}
	assertRequestLog(t, logger, 17, 3, dailybesttypes.FetchLogStatusSucceeded, http.StatusOK)
}

func TestClientRecordsHTTPFailure(t *testing.T) {
	logger := &recordingRequestLogger{}
	pageClient := &fakePageClient{response: Response{StatusCode: http.StatusServiceUnavailable, Body: []byte("unavailable")}}
	client := NewClient(pageClient, logger)

	response, err := client.Get(context.Background(), 18, 4, "https://example.com/page=4")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", response.StatusCode)
	}
	assertRequestLog(t, logger, 18, 4, dailybesttypes.FetchLogStatusFailed, http.StatusServiceUnavailable)
	if logger.results[0].ErrorMessage == "" {
		t.Fatal("HTTP 失败日志缺少错误信息")
	}
}

func TestClientRecordsNetworkFailure(t *testing.T) {
	logger := &recordingRequestLogger{}
	pageClient := &fakePageClient{err: errors.New("browser unavailable")}
	client := NewClient(pageClient, logger)

	_, err := client.Get(context.Background(), 19, 5, "https://example.com/page=5")
	if err == nil {
		t.Fatal("expected network error")
	}
	assertRequestLog(t, logger, 19, 5, dailybesttypes.FetchLogStatusFailed, 0)
	if logger.results[0].ErrorMessage == "" {
		t.Fatal("网络失败日志缺少错误信息")
	}
}

func TestClientRecordsBrowserFailureHTTPStatus(t *testing.T) {
	logger := &recordingRequestLogger{}
	pageClient := &fakePageClient{
		response: Response{StatusCode: http.StatusForbidden},
		err:      errors.New("challenge not completed"),
	}
	client := NewClient(pageClient, logger)

	_, err := client.Get(context.Background(), 21, 7, "https://example.com/page=7")
	if err == nil {
		t.Fatal("expected browser error")
	}
	assertRequestLog(t, logger, 21, 7, dailybesttypes.FetchLogStatusFailed, http.StatusForbidden)
}

func TestClientDoesNotRequestWhenInitialLogFails(t *testing.T) {
	logger := &recordingRequestLogger{createErr: errors.New("database unavailable")}
	pageClient := &fakePageClient{err: errors.New("unexpected request")}
	client := NewClient(pageClient, logger)

	_, err := client.Get(context.Background(), 20, 6, "https://example.com/page=6")
	if err == nil {
		t.Fatal("expected log creation error")
	}
	if pageClient.called {
		t.Fatal("日志创建失败后仍然发出了远程请求")
	}
}

func assertRequestLog(
	t *testing.T,
	logger *recordingRequestLogger,
	fetchRunID, pageNumber int64,
	status string,
	httpStatus int,
) {
	t.Helper()
	if len(logger.drafts) != 1 || len(logger.results) != 1 {
		t.Fatalf("drafts=%d results=%d", len(logger.drafts), len(logger.results))
	}
	draft := logger.drafts[0]
	if draft.FetchRunID != fetchRunID || draft.PageNumber != pageNumber || draft.AttemptNumber != 1 || draft.RequestURL == "" {
		t.Fatalf("draft=%+v", draft)
	}
	result := logger.results[0]
	if result.Status != status || result.HTTPStatus != int64(httpStatus) || result.DurationMS < 0 {
		t.Fatalf("result=%+v", result)
	}
}
