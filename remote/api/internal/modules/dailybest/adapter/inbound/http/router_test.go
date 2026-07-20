package dailybestremotehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"rudy_gc/remote/api/internal/modules/dailybest/domain"
	contracts "rudy_gc/shared/dailybest"

	"github.com/gin-gonic/gin"
)

type fakeRepository struct{}

func (fakeRepository) CreateFetchRun(context.Context, string, int64, string) (contracts.DailyBestFetchRun, error) {
	return contracts.DailyBestFetchRun{}, nil
}
func (fakeRepository) MarkFetchRunRunning(context.Context, int64, int64) error        { return nil }
func (fakeRepository) MarkFetchRunFailed(context.Context, int64, int64, string) error { return nil }
func (fakeRepository) PublishSnapshot(context.Context, int64, dailybest.SnapshotDraft, int64) (dailybest.PublishResult, error) {
	return dailybest.PublishResult{}, nil
}
func (fakeRepository) ListSnapshotsAfter(context.Context, int64) ([]contracts.DailyBestSnapshotManifest, error) {
	return []contracts.DailyBestSnapshotManifest{{Revision: 3, BusinessDate: "2026-07-19", Category: 1}}, nil
}
func (fakeRepository) FindSnapshot(context.Context, int64, bool) (contracts.DailyBestSnapshotResponse, error) {
	return contracts.DailyBestSnapshotResponse{}, nil
}
func (fakeRepository) FindLatestSnapshotByDate(context.Context, string, int64, bool) (contracts.DailyBestSnapshotResponse, error) {
	return contracts.DailyBestSnapshotResponse{}, nil
}
func (fakeRepository) ListFetchRuns(context.Context) ([]contracts.DailyBestFetchRun, error) {
	return nil, nil
}
func (fakeRepository) FindFetchRun(context.Context, int64) (contracts.DailyBestFetchRun, error) {
	return contracts.DailyBestFetchRun{}, nil
}
func (fakeRepository) Ping(context.Context) error { return nil }

func TestSnapshotAPIRequiresReadToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New(fakeRepository{}, nil, "read-secret", "admin-secret")

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/snapshots?after_revision=0", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d", unauthorized.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/snapshots?after_revision=0", nil)
	req.Header.Set("Authorization", "Bearer read-secret")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, req)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status=%d body=%s", authorized.Code, authorized.Body.String())
	}
}
