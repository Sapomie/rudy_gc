package movie

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

func TestParseRankStatDateRange(t *testing.T) {
	start, end, err := parseRankStatDateRange("2026-07-01", "2026-07-20")
	if err != nil {
		t.Fatal(err)
	}
	if start <= 0 || end <= start {
		t.Fatalf("unexpected range: %d-%d", start, end)
	}
	if _, _, err := parseRankStatDateRange("2026-07-20", "2026-07-01"); err == nil {
		t.Fatal("expected reversed date range error")
	}
	if _, _, err := parseRankStatDateRange("2023-01-01", "2022-12-26"); err == nil {
		t.Fatal("expected reversed date range error across rank day zero")
	}
}

func TestNormalizeRankStatSortAndOrder(t *testing.T) {
	if got := normalizeRankDayStatSort("movie_count"); got != "movie_count" {
		t.Fatalf("day sort = %q", got)
	}
	if got := normalizeRankDayStatSort("unchanged_movie_count"); got != "day_number" {
		t.Fatalf("removed day sort = %q", got)
	}
	if got := normalizeRankPeriodStatSort("missing_rank_day_count"); got != "missing_rank_day_count" {
		t.Fatalf("period missing day sort = %q", got)
	}
	if got := normalizeRankPeriodStatSort("new_entry_movie_count"); got != "new_entry_movie_count" {
		t.Fatalf("period new entry sort = %q", got)
	}
	if got := normalizeRankPeriodStatSort("source_fetched_end_on"); got != "source_fetched_end_on" {
		t.Fatalf("period fetched end sort = %q", got)
	}
	if got := normalizeRankPeriodStatSort("unsafe"); got != "start_day_number" {
		t.Fatalf("period sort = %q", got)
	}
	if got := normalizeRankStatOrder("ASC"); got != "asc" {
		t.Fatalf("order = %q", got)
	}
	if got := normalizeRankPeriodStatType(""); got != "month" {
		t.Fatalf("default period type = %q", got)
	}
	if got := normalizeRankPeriodStatType("week"); got != "week" {
		t.Fatalf("week period type = %q", got)
	}
	if got := normalizeRankPeriodStatType("unsafe"); got != "month" {
		t.Fatalf("invalid period type = %q", got)
	}
	query := rankStatListQuery{Page: -1, PageSize: 1000}
	normalizeRankStatPage(&query)
	if query.Page != 1 || query.PageSize != 1000 {
		t.Fatalf("pagination = page %d, size %d", query.Page, query.PageSize)
	}
	query = rankStatListQuery{}
	normalizeRankStatPage(&query)
	if query.Page != 1 || query.PageSize != rankStatDefaultPageSize {
		t.Fatalf("default pagination = page %d, size %d", query.Page, query.PageSize)
	}
}

func TestRankStatLinksPreserveFiltersAndResetPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(
		http.MethodGet,
		"/rankperiodstats?p=8&ps=24&type=month&start_date=2026-07-01&sort=period_movie_count&order=desc",
		nil,
	)

	typeLinks := buildRankPeriodStatTypeLinks(c, "month")
	if len(typeLinks) != 4 || !typeLinks[1].Active {
		t.Fatalf("unexpected type links: %+v", typeLinks)
	}
	weekURL, err := url.Parse(typeLinks[0].Href)
	if err != nil {
		t.Fatal(err)
	}
	if weekURL.Query().Get("p") != "1" || weekURL.Query().Get("type") != "week" {
		t.Fatalf("unexpected week href: %s", typeLinks[0].Href)
	}
	if weekURL.Query().Get("start_date") != "2026-07-01" || weekURL.Query().Get("ps") != "24" {
		t.Fatalf("type link lost filters: %s", typeLinks[0].Href)
	}

	sortLink := buildRankStatSortLink(c, "period_movie_count", "desc", "period_movie_count", "排行数")
	sortURL, err := url.Parse(sortLink.Href)
	if err != nil {
		t.Fatal(err)
	}
	if !sortLink.Active || !sortLink.Desc || sortURL.Query().Get("p") != "1" || sortURL.Query().Get("order") != "asc" {
		t.Fatalf("unexpected sort link: %+v", sortLink)
	}

	periodHeaderLinks := buildRankPeriodStatHeaderLinks(c, "period_movie_count", "desc")
	for _, key := range []string{
		"SourceRankDayCount",
		"MissingRankDayCount",
		"PeriodMovieCount",
		"NewEntryMovieCount",
		"SourceMovieCount",
	} {
		if _, ok := periodHeaderLinks[key]; !ok {
			t.Fatalf("period header link %s is missing", key)
		}
	}
	if _, ok := periodHeaderLinks["SourceRankRowCount"]; ok {
		t.Fatal("period source rank row header should be hidden")
	}
	if periodHeaderLinks["PeriodMovieCount"].Label != "排行数" {
		t.Fatalf("period movie count label = %q", periodHeaderLinks["PeriodMovieCount"].Label)
	}

	dayHeaderLinks := buildRankDayStatHeaderLinks(c, "continued_movie_count", "desc")
	for key, label := range map[string]string{
		"ContinuedMovieCount": "连续",
		"FirstSeenMovieCount": "首次",
		"ReturningMovieCount": "回归",
	} {
		if dayHeaderLinks[key].Label != label {
			t.Fatalf("day header %s label = %q", key, dayHeaderLinks[key].Label)
		}
	}
}

type rankStatAPITestResponse struct {
	Rows     []map[string]any `json:"Rows"`
	Total    int64            `json:"Total"`
	PageInfo struct {
		Page      int64 `json:"Page"`
		PageSize  int64 `json:"PageSize"`
		PageTotal int64 `json:"PageTotal"`
	} `json:"PageInfo"`
	Summary map[string]any `json:"Summary"`
}

type rankStatDayModelStub struct {
	query moviex.CRankDayStatListQuery
}

func (*rankStatDayModelStub) InsertMissing(context.Context, []*moviex.CRankDayStat) error {
	return nil
}

func (*rankStatDayModelStub) ListAll(context.Context) ([]*moviex.CRankDayStat, error) {
	return nil, nil
}

func (m *rankStatDayModelStub) ListPage(
	_ context.Context,
	query moviex.CRankDayStatListQuery,
) (*moviex.CRankDayStatPage, error) {
	m.query = query
	return &moviex.CRankDayStatPage{
		Items: []*moviex.CRankDayStat{{DayNumber: 1, RankDate: "2023-01-01"}},
		Total: 1,
		Page:  query.Page,
		Summary: &moviex.CRankDayStatSummary{
			DayCount: 1,
		},
	}, nil
}

type rankStatPeriodModelStub struct {
	query moviex.CRankPeriodStatListQuery
}

func (*rankStatPeriodModelStub) Upsert(context.Context, []*moviex.CRankPeriodStat) error { return nil }

func (m *rankStatPeriodModelStub) ListPage(
	_ context.Context,
	query moviex.CRankPeriodStatListQuery,
) (*moviex.CRankPeriodStatPage, error) {
	m.query = query
	return &moviex.CRankPeriodStatPage{
		Items: []*moviex.CRankPeriodStat{{
			PeriodType:         query.PeriodType,
			PeriodKey:          "2026-07",
			NewEntryMovieCount: 2,
		}},
		Total: 1,
		Page:  query.Page,
		Summary: &moviex.CRankPeriodStatSummary{
			PeriodCount: 1,
		},
	}, nil
}

func TestRankStatAPIUsesExistingListContractAndSinglePeriodDefaults(t *testing.T) {
	dayModel := new(rankStatDayModelStub)
	periodModel := new(rankStatPeriodModelStub)
	api := NewRankStatAPI(&dep.Dep{
		RankDayStatModel:    dayModel,
		RankPeriodStatModel: periodModel,
	})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/rank-day-stats", api.ListDayStats)
	router.GET("/api/rank-period-stats", api.ListPeriodStats)

	requestJSON := func(path string) map[string]any {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status=%d body=%s", path, response.Code, response.Body.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		return payload
	}

	dayPayload := requestJSON("/api/rank-day-stats")
	if _, ok := dayPayload["Rows"]; !ok || dayPayload["Query"] == nil || dayPayload["HeaderLinks"] == nil || dayPayload["PageInfo"] == nil {
		t.Fatalf("day list contract is incomplete: %+v", dayPayload)
	}
	if dayModel.query.Page != 1 || dayModel.query.PageSize != rankStatDefaultPageSize {
		t.Fatalf("day default pagination = page %d size %d", dayModel.query.Page, dayModel.query.PageSize)
	}

	periodPayload := requestJSON("/api/rank-period-stats")
	if periodModel.query.PeriodType != consts.RankPeriodTypeMonth {
		t.Fatalf("period default type = %d", periodModel.query.PeriodType)
	}
	query := periodPayload["Query"].(map[string]any)
	if query["Type"] != "month" {
		t.Fatalf("period response query = %+v", query)
	}
	if len(periodPayload["TypeLinks"].([]any)) != 4 {
		t.Fatalf("period switches are incomplete: %+v", periodPayload)
	}
	if _, ok := periodPayload["CategoryLinks"]; ok {
		t.Fatalf("period response includes category links: %+v", periodPayload)
	}
	rows := periodPayload["Rows"].([]any)
	if len(rows) != 1 || int64(rows[0].(map[string]any)["PeriodType"].(float64)) != consts.RankPeriodTypeMonth {
		t.Fatalf("period rows are not limited to the selected type: %+v", rows)
	}
	if int64(rows[0].(map[string]any)["NewEntryMovieCount"].(float64)) != 2 {
		t.Fatalf("period new entry count is missing: %+v", rows)
	}
}

func TestRankStatAPIAgainstConfiguredDatabase(t *testing.T) {
	dsn := os.Getenv("RANK_STAT_INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("RANK_STAT_INTEGRATION_DSN is empty")
	}

	conn := sqlx.NewMysql(dsn)
	deps := &dep.Dep{
		RankDayStatModel:    moviex.NewCRankDayStatModel(conn),
		RankPeriodStatModel: moviex.NewCRankPeriodStatModel(conn),
	}
	api := NewRankStatAPI(deps)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/rank-day-stats", api.ListDayStats)
	router.GET("/api/rank-period-stats", api.ListPeriodStats)

	requestPage := func(path string) rankStatAPITestResponse {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status=%d body=%s", path, response.Code, response.Body.String())
		}
		var payload rankStatAPITestResponse
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		return payload
	}

	firstDayPage := requestPage("/api/rank-day-stats?p=1&ps=7&order=desc")
	secondDayPage := requestPage("/api/rank-day-stats?p=2&ps=7&order=desc")
	if len(firstDayPage.Rows) != 7 || len(secondDayPage.Rows) != 7 || firstDayPage.Total != secondDayPage.Total {
		t.Fatalf("unexpected day pagination: first=%+v second=%+v", firstDayPage.PageInfo, secondDayPage.PageInfo)
	}
	if firstDayPage.PageInfo.Page != 1 || secondDayPage.PageInfo.Page != 2 || firstDayPage.PageInfo.PageSize != 7 {
		t.Fatalf("unexpected day page info: first=%+v second=%+v", firstDayPage.PageInfo, secondDayPage.PageInfo)
	}
	if firstDayPage.Rows[0]["DayNumber"] == secondDayPage.Rows[0]["DayNumber"] {
		t.Fatal("day pages overlap at the first row")
	}
	if int64(firstDayPage.Summary["DayCount"].(float64)) != firstDayPage.Total {
		t.Fatalf("day summary count=%v total=%d", firstDayPage.Summary["DayCount"], firstDayPage.Total)
	}
	lastDayPage := requestPage("/api/rank-day-stats?p=999999&ps=7&order=desc")
	if lastDayPage.PageInfo.Page != lastDayPage.PageInfo.PageTotal || len(lastDayPage.Rows) == 0 {
		t.Fatalf("day out-of-range page was not clamped: %+v", lastDayPage.PageInfo)
	}

	periodPage := requestPage("/api/rank-period-stats?p=1&ps=5&type=week&order=desc")
	if len(periodPage.Rows) != 5 || periodPage.Total <= int64(len(periodPage.Rows)) {
		t.Fatalf("unexpected period pagination: total=%d rows=%d", periodPage.Total, len(periodPage.Rows))
	}
	if periodPage.PageInfo.Page != 1 || periodPage.PageInfo.PageSize != 5 || periodPage.PageInfo.PageTotal <= 1 {
		t.Fatalf("unexpected period page info: %+v", periodPage.PageInfo)
	}
	if int64(periodPage.Summary["PeriodCount"].(float64)) != periodPage.Total {
		t.Fatalf("period summary count=%v total=%d", periodPage.Summary["PeriodCount"], periodPage.Total)
	}
	lastPeriodPage := requestPage("/api/rank-period-stats?p=999999&ps=5&type=week&order=desc")
	if lastPeriodPage.PageInfo.Page != lastPeriodPage.PageInfo.PageTotal || len(lastPeriodPage.Rows) == 0 {
		t.Fatalf("period out-of-range page was not clamped: %+v", lastPeriodPage.PageInfo)
	}
}
