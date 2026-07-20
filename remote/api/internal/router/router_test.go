package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type fakeRepository struct{}

func (fakeRepository) CreateFetchRun(context.Context, string, int64, string) (contracts.DailyBestFetchRun, error) {
	return contracts.DailyBestFetchRun{}, nil
}
func (fakeRepository) MarkFetchRunRunning(context.Context, int64, int64) error        { return nil }
func (fakeRepository) MarkFetchRunFailed(context.Context, int64, int64, string) error { return nil }
func (fakeRepository) MarkFetchRunNoChange(context.Context, int64, int64, int64, int64, string, []dailybesttypes.MovieDraft) error {
	return nil
}
func (fakeRepository) PublishRanks(context.Context, int64, dailybesttypes.RankDraft, dailybesttypes.RankProjection, int64) error {
	return nil
}
func (fakeRepository) RankProjectionComplete(context.Context) (bool, error) { return true, nil }
func (fakeRepository) SyncRankProjection(context.Context, dailybesttypes.RankProjection) error {
	return nil
}
func (fakeRepository) ListAllRanks(context.Context) ([]contracts.DailyBestRank, error) {
	return nil, nil
}
func (fakeRepository) ListRankSyncManifests(context.Context) ([]contracts.DailyBestRankSyncManifest, error) {
	return []contracts.DailyBestRankSyncManifest{{FetchRunID: 3, BusinessDate: "2026-07-19", Category: 1}}, nil
}
func (fakeRepository) ListRanks(context.Context, dailybesttypes.RankFilter) ([]contracts.DailyBestRank, int64, error) {
	return []contracts.DailyBestRank{{
		ID: 1, RankPos: 1, MovieJavID: "jav-1",
		Movie: contracts.DailyBestMovieBasic{
			MovieName: "ABC-001", MovieCode: "ABC-001", Title: "ABC-001 完整标题",
			CoverURL: "https://example.com/cover.jpg",
		},
	}}, 25, nil
}
func (fakeRepository) ListRankDayStats(context.Context, dailybesttypes.RankDayStatFilter) ([]contracts.DailyBestRankDayStat, int64, error) {
	return []contracts.DailyBestRankDayStat{{ID: 1}}, 25, nil
}
func (fakeRepository) ListRankPeriods(context.Context, dailybesttypes.RankPeriodFilter) ([]contracts.DailyBestRankPeriod, int64, error) {
	return []contracts.DailyBestRankPeriod{{ID: 1}}, 25, nil
}
func (fakeRepository) ListRankPeriodItems(context.Context, dailybesttypes.RankPeriodItemFilter) ([]contracts.DailyBestRankPeriodItem, int64, error) {
	return []contracts.DailyBestRankPeriodItem{{ID: 1}}, 25, nil
}
func (fakeRepository) ListRankPeriodStats(context.Context, dailybesttypes.RankPeriodStatFilter) ([]contracts.DailyBestRankPeriodStat, int64, error) {
	return []contracts.DailyBestRankPeriodStat{{ID: 1}}, 25, nil
}
func (fakeRepository) ListFetchRuns(context.Context, dailybesttypes.PageQuery) ([]contracts.DailyBestFetchRun, int64, error) {
	return []contracts.DailyBestFetchRun{{ID: 1}}, 25, nil
}
func (fakeRepository) FindFetchRun(context.Context, int64) (contracts.DailyBestFetchRun, error) {
	return contracts.DailyBestFetchRun{}, nil
}
func (fakeRepository) Ping(context.Context) error { return nil }

func TestRankSyncManifestAPIIsAvailableWithoutCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/rank-sync-manifests", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), `"page_info"`) {
		t.Fatalf("sync response must remain unpaged body=%s", response.Body.String())
	}
}

func TestRankTableAPIs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	for _, path := range []string{
		"/api/v1/dailybest/ranks?p=2&ps=24",
		"/api/v1/dailybest/rank-day-stats?p=2&ps=24",
		"/api/v1/dailybest/rank-periods?p=2&ps=24",
		"/api/v1/dailybest/rank-period-items?p=2&ps=24",
		"/api/v1/dailybest/rank-period-stats?p=2&ps=24",
		"/api/v1/dailybest/fetch-runs?p=2&ps=24",
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("path=%s status=%d body=%s", path, response.Code, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), `"page_info":{"total":25,"page":2,"page_size":24,"page_total":2}`) {
			t.Fatalf("path=%s missing page_info body=%s", path, response.Body.String())
		}
	}
}

func TestRankAPIIncludesMovieBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/ranks", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	for _, expected := range []string{`"movie_name":"ABC-001"`, `"title":"ABC-001 完整标题"`, `"cover_url":"https://example.com/cover.jpg"`} {
		if !strings.Contains(response.Body.String(), expected) {
			t.Fatalf("missing %s body=%s", expected, response.Body.String())
		}
	}
}

func TestRankTableAPIRejectsMalformedPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/ranks?p=bad", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
