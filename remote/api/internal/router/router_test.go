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

type fakeRepository struct {
	createFetchRun func(context.Context, string, int64, string) (contracts.DailyBestFetchRun, error)
	listFetchLogs  func(context.Context, dailybesttypes.FetchLogFilter) ([]dailybesttypes.FetchLog, int64, error)
}

func (f fakeRepository) CreateFetchRun(ctx context.Context, businessDate string, category int64, triggerType string) (contracts.DailyBestFetchRun, error) {
	if f.createFetchRun != nil {
		return f.createFetchRun(ctx, businessDate, category, triggerType)
	}
	return contracts.DailyBestFetchRun{ID: 1, Status: dailybesttypes.RunStatusQueued}, nil
}
func (fakeRepository) MarkFetchRunRunning(context.Context, int64, int64) error        { return nil }
func (fakeRepository) MarkFetchRunFailed(context.Context, int64, int64, string) error { return nil }
func (fakeRepository) InterruptActiveFetches(context.Context, int64, string) (int64, int64, error) {
	return 0, 0, nil
}
func (fakeRepository) MarkFetchRunNoChange(context.Context, int64, int64, int64, int64, string, []dailybesttypes.MovieDraft) error {
	return nil
}
func (fakeRepository) PublishRanks(context.Context, int64, dailybesttypes.RankDraft, contracts.DailyBestRankDayStat, int64) error {
	return nil
}
func (fakeRepository) ListRankDayStatDayNumbers(context.Context) ([]int64, error) {
	return nil, nil
}
func (fakeRepository) InsertRankDayStatsIfMissing(context.Context, []contracts.DailyBestRankDayStat) error {
	return nil
}
func (fakeRepository) ListAllRanks(context.Context) ([]contracts.DailyBestRank, error) {
	return nil, nil
}
func (fakeRepository) ListRankDays(context.Context, dailybesttypes.RankDayFilter) ([]contracts.DailyBestRankDay, int64, error) {
	return []contracts.DailyBestRankDay{{
		BusinessDate: "2026-07-19", DayNumber: 1296, RankCount: 465,
		Categories: []contracts.DailyBestRankDayCategory{{FetchRunID: 3, Category: 1, RankCount: 465}},
	}}, 25, nil
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
func (fakeRepository) ListFetchRuns(context.Context, dailybesttypes.PageQuery) ([]contracts.DailyBestFetchRun, int64, error) {
	return []contracts.DailyBestFetchRun{{ID: 1}}, 25, nil
}
func (f fakeRepository) ListFetchLogs(ctx context.Context, filter dailybesttypes.FetchLogFilter) ([]dailybesttypes.FetchLog, int64, error) {
	if f.listFetchLogs != nil {
		return f.listFetchLogs(ctx, filter)
	}
	return []dailybesttypes.FetchLog{{
		ID: 9, FetchRunID: 7, PageNumber: 3, AttemptNumber: 1,
		RequestURL: "https://example.com/page=3", Status: dailybesttypes.FetchLogStatusSucceeded,
		HTTPStatus: 200, DurationMS: 128, CreatedOn: 10, UpdatedOn: 11,
	}}, 25, nil
}
func (fakeRepository) FindFetchRun(context.Context, int64) (contracts.DailyBestFetchRun, error) {
	return contracts.DailyBestFetchRun{}, nil
}
func (fakeRepository) Ping(context.Context) error { return nil }

func TestRankDayAPIIsAvailableWithoutCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/rank-days?p=1&ps=24&business_date_start=2026-07-01", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"days":[{"business_date":"2026-07-19"`) {
		t.Fatalf("missing rank days body=%s", response.Body.String())
	}
}

type blankPageSource struct{}

func (blankPageSource) FetchPage(context.Context, int64, int64) (dailybesttypes.SourcePage, error) {
	return dailybesttypes.SourcePage{}, dailybesttypes.ErrBlankPage
}

func TestManualFetchAcceptsEmptyRequestAndResolvesInternalFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	created := make(chan contracts.DailyBestFetchRun, 1)
	repository := fakeRepository{createFetchRun: func(
		_ context.Context,
		businessDate string,
		category int64,
		triggerType string,
	) (contracts.DailyBestFetchRun, error) {
		run := contracts.DailyBestFetchRun{
			ID: 7, BusinessDate: businessDate, Category: category,
			TriggerType: triggerType, Status: dailybesttypes.RunStatusQueued,
		}
		created <- run
		return run, nil
	}}
	service := dailybestservice.New(context.Background(), repository, blankPageSource{}, time.UTC)
	engine := New(service)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/dailybest/fetch", nil))
	service.Wait()

	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	run := <-created
	if run.BusinessDate == "" || run.Category != contracts.DailyBestCategoryMonth || run.TriggerType != dailybesttypes.TriggerManual {
		t.Fatalf("run=%+v", run)
	}
}

func TestManualFetchRejectsSecondActiveRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{
		createFetchRun: func(context.Context, string, int64, string) (contracts.DailyBestFetchRun, error) {
			return contracts.DailyBestFetchRun{}, dailybesttypes.ErrActiveFetchRun
		},
	}, blankPageSource{}, time.UTC)
	response := httptest.NewRecorder()
	New(service).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/dailybest/fetch", nil))
	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"code":"active_fetch_run_exists"`) {
		t.Fatalf("body=%s", response.Body.String())
	}
}

func TestRankTableAPIs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	for _, path := range []string{
		"/api/v1/dailybest/rank-days?p=2&ps=24",
		"/api/v1/dailybest/ranks?p=2&ps=24",
		"/api/v1/dailybest/rank-day-stats?p=2&ps=24",
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

func TestFetchLogAPIUsesFiltersAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	filters := make(chan dailybesttypes.FetchLogFilter, 1)
	service := dailybestservice.New(context.Background(), fakeRepository{
		listFetchLogs: func(_ context.Context, filter dailybesttypes.FetchLogFilter) ([]dailybesttypes.FetchLog, int64, error) {
			filters <- filter
			return []dailybesttypes.FetchLog{{
				ID: 9, FetchRunID: 7, PageNumber: 3, AttemptNumber: 1,
				RequestURL: "https://example.com/page=3", Status: dailybesttypes.FetchLogStatusFailed,
				HTTPStatus: 503, DurationMS: 128, ErrorMessage: "HTTP 状态码 503", CreatedOn: 10, UpdatedOn: 11,
			}}, 25, nil
		},
	}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(
		http.MethodGet,
		"/api/v1/dailybest/fetch-logs?fetch_run_id=7&page_number=3&status=failed&p=2&ps=24",
		nil,
	))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	filter := <-filters
	if filter.FetchRunID != 7 || filter.PageNumber != 3 || filter.Status != dailybesttypes.FetchLogStatusFailed || filter.Page != 2 || filter.PageSize != 24 {
		t.Fatalf("filter=%+v", filter)
	}
	if !strings.Contains(response.Body.String(), `"logs":[{"id":9,"fetch_run_id":7,"page_number":3`) {
		t.Fatalf("body=%s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"page_info":{"total":25,"page":2,"page_size":24,"page_total":2}`) {
		t.Fatalf("body=%s", response.Body.String())
	}
}

func TestFetchLogAPIRejectsUnknownStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/fetch-logs?status=unknown", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestRankSyncManifestAPIIsRemoved(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/rank-sync-manifests", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestRankDayAPIRejectsInvertedDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/dailybest/rank-days?business_date_start=2026-07-21&business_date_end=2026-07-01", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestRankDayAPIRequiresPaginationOnlyForLists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	for path, expected := range map[string]int{
		"/api/v1/dailybest/rank-days":                 http.StatusBadRequest,
		"/api/v1/dailybest/rank-days?day_number=1296": http.StatusOK,
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != expected {
			t.Fatalf("path=%s status=%d want=%d body=%s", path, response.Code, expected, response.Body.String())
		}
	}
}

func TestRankPeriodAPIsAreRemoved(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := dailybestservice.New(context.Background(), fakeRepository{}, nil, time.UTC)
	router := New(service)
	for _, path := range []string{
		"/api/v1/dailybest/rank-periods",
		"/api/v1/dailybest/rank-period-items",
		"/api/v1/dailybest/rank-period-stats",
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("path=%s status=%d body=%s", path, response.Code, response.Body.String())
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
