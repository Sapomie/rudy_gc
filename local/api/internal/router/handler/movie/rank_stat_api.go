package movie

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/router/handler/common"
	moviesvc "rudy_gc/local/api/internal/service/movie"
)

const (
	rankStatDefaultPageSize int64 = 24
)

type RankStatAPI struct {
	movieSvc *moviesvc.Service
}

func NewRankStatAPI(deps *dep.Dep) *RankStatAPI {
	return &RankStatAPI{movieSvc: moviesvc.NewService(deps)}
}

type rankStatListQuery struct {
	Page      int64  `form:"p"`
	PageSize  int64  `form:"ps"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Type      string `form:"type"`
	Sort      string `form:"sort"`
	Order     string `form:"order"`
}

type rankStatSortLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

func (h *RankStatAPI) ListDayStats(c *gin.Context) {
	defer common.WithFrontendPath(c, "/rankdaystats")()

	var query rankStatListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
		return
	}
	startDayNumber, endDayNumber, err := parseRankStatDateRange(query.StartDate, query.EndDate)
	if err != nil {
		common.JSONBadRequest(c, "日期筛选参数错误", err)
		return
	}
	query.Sort = normalizeRankDayStatSort(query.Sort)
	query.Order = normalizeRankStatOrder(query.Order)
	normalizeRankStatPage(&query)

	result, err := h.movieSvc.ListRankDayStats(c.Request.Context(), moviesvc.RankDayStatListRequest{
		StartDayNumber:    startDayNumber,
		EndDayNumber:      endDayNumber,
		HasStartDayNumber: strings.TrimSpace(query.StartDate) != "",
		HasEndDayNumber:   strings.TrimSpace(query.EndDate) != "",
		Page:              query.Page,
		PageSize:          query.PageSize,
		Sort:              query.Sort,
		Order:             query.Order,
	})
	if err != nil {
		common.JSONServerError(c, "加载日排行统计失败", err)
		return
	}
	pageInfo := common.BuildPageInfo(c, result.Total, result.Page, query.PageSize, common.PageWindow)

	c.JSON(http.StatusOK, gin.H{
		"Title":       "Rank Day Stats",
		"ViewType":    "rank-day-stats",
		"PageTitle":   "日排行统计",
		"PageNote":    "直接查看 c_rank_day_stat 落库记录。",
		"Rows":        result.Items,
		"Total":       result.Total,
		"PageInfo":    pageInfo,
		"pageInfo":    pageInfo,
		"Summary":     result.Summary,
		"Query":       query,
		"HeaderLinks": buildRankDayStatHeaderLinks(c, query.Sort, query.Order),
		"ClearHref":   "/rankdaystats",
	})
}

func (h *RankStatAPI) ListPeriodStats(c *gin.Context) {
	defer common.WithFrontendPath(c, "/rankperiodstats")()

	var query rankStatListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
		return
	}
	query.Type = normalizeRankPeriodStatType(query.Type)
	startDayNumber, endDayNumber, err := parseRankStatDateRange(query.StartDate, query.EndDate)
	if err != nil {
		common.JSONBadRequest(c, "日期筛选参数错误", err)
		return
	}
	query.Sort = normalizeRankPeriodStatSort(query.Sort)
	query.Order = normalizeRankStatOrder(query.Order)
	normalizeRankStatPage(&query)

	result, err := h.movieSvc.ListRankPeriodStats(c.Request.Context(), moviesvc.RankPeriodStatListRequest{
		PeriodType:        consts.RankPeriodTypeFromName(query.Type),
		StartDayNumber:    startDayNumber,
		EndDayNumber:      endDayNumber,
		HasStartDayNumber: strings.TrimSpace(query.StartDate) != "",
		HasEndDayNumber:   strings.TrimSpace(query.EndDate) != "",
		Page:              query.Page,
		PageSize:          query.PageSize,
		Sort:              query.Sort,
		Order:             query.Order,
	})
	if err != nil {
		common.JSONServerError(c, "加载周期排行统计失败", err)
		return
	}
	pageInfo := common.BuildPageInfo(c, result.Total, result.Page, query.PageSize, common.PageWindow)

	c.JSON(http.StatusOK, gin.H{
		"Title":           "Rank Period Stats",
		"ViewType":        "rank-period-stats",
		"PageTitle":       "周期排行统计",
		"PageNote":        "汇总 c_rank 与 c_rank_day_stat 的周期统计记录。",
		"Rows":            result.Items,
		"Total":           result.Total,
		"PageInfo":        pageInfo,
		"pageInfo":        pageInfo,
		"Summary":         result.Summary,
		"Query":           query,
		"TypeLinks":       buildRankPeriodStatTypeLinks(c, query.Type),
		"PeriodTypeLabel": consts.RankPeriodTypeLabel(consts.RankPeriodTypeFromName(query.Type)),
		"HeaderLinks":     buildRankPeriodStatHeaderLinks(c, query.Sort, query.Order),
		"ClearHref":       "/rankperiodstats",
	})
}

func normalizeRankStatPage(query *rankStatListQuery) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = rankStatDefaultPageSize
	}
}

func normalizeRankPeriodStatType(value string) string {
	switch strings.TrimSpace(value) {
	case "week", "month", "quarter", "year":
		return strings.TrimSpace(value)
	default:
		return consts.RankPeriodTypeName(consts.RankPeriodTypeMonth)
	}
}

func parseRankStatDateRange(startDate, endDate string) (int64, int64, error) {
	startDayNumber, err := parseOptionalRankStatDate(startDate)
	if err != nil {
		return 0, 0, err
	}
	endDayNumber, err := parseOptionalRankStatDate(endDate)
	if err != nil {
		return 0, 0, err
	}
	if strings.TrimSpace(startDate) != "" && strings.TrimSpace(endDate) != "" && startDayNumber > endDayNumber {
		return 0, 0, &rankStatDateRangeError{}
	}
	return startDayNumber, endDayNumber, nil
}

func parseOptionalRankStatDate(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, err := time.ParseInLocation(time.DateOnly, value, time.Local)
	if err != nil {
		return 0, err
	}
	return consts.GetRankDayNumber(parsed.Format(time.DateOnly)), nil
}

type rankStatDateRangeError struct{}

func (*rankStatDateRangeError) Error() string { return "起始日期不能晚于结束日期" }

func normalizeRankStatOrder(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "asc"
	}
	return "desc"
}

func normalizeRankDayStatSort(value string) string {
	switch strings.TrimSpace(value) {
	case "rank_row_count",
		"movie_count",
		"continued_movie_count",
		"first_seen_movie_count",
		"returning_movie_count",
		"dropped_movie_count",
		"calculated_on":
		return strings.TrimSpace(value)
	default:
		return "day_number"
	}
}

func normalizeRankPeriodStatSort(value string) string {
	switch strings.TrimSpace(value) {
	case "source_rank_row_count",
		"source_rank_day_count",
		"source_movie_count",
		"missing_rank_day_count",
		"period_movie_count",
		"new_entry_movie_count",
		"source_fetched_start_on",
		"source_fetched_end_on",
		"calculated_on":
		return strings.TrimSpace(value)
	default:
		return "start_day_number"
	}
}

func buildRankDayStatHeaderLinks(c *gin.Context, currentSort, currentOrder string) map[string]rankStatSortLink {
	return map[string]rankStatSortLink{
		"DayNumber":           buildRankStatSortLink(c, currentSort, currentOrder, "day_number", "排行日期"),
		"RankRowCount":        buildRankStatSortLink(c, currentSort, currentOrder, "rank_row_count", "排行明细"),
		"MovieCount":          buildRankStatSortLink(c, currentSort, currentOrder, "movie_count", "去重影片"),
		"ContinuedMovieCount": buildRankStatSortLink(c, currentSort, currentOrder, "continued_movie_count", "连续"),
		"FirstSeenMovieCount": buildRankStatSortLink(c, currentSort, currentOrder, "first_seen_movie_count", "首次"),
		"ReturningMovieCount": buildRankStatSortLink(c, currentSort, currentOrder, "returning_movie_count", "回归"),
		"DroppedMovieCount":   buildRankStatSortLink(c, currentSort, currentOrder, "dropped_movie_count", "掉榜"),
		"CalculatedOn":        buildRankStatSortLink(c, currentSort, currentOrder, "calculated_on", "计算时间"),
	}
}

func buildRankPeriodStatHeaderLinks(c *gin.Context, currentSort, currentOrder string) map[string]rankStatSortLink {
	return map[string]rankStatSortLink{
		"StartDayNumber":      buildRankStatSortLink(c, currentSort, currentOrder, "start_day_number", "周期"),
		"SourceRankDayCount":  buildRankStatSortLink(c, currentSort, currentOrder, "source_rank_day_count", "排行日"),
		"MissingRankDayCount": buildRankStatSortLink(c, currentSort, currentOrder, "missing_rank_day_count", "缺失日"),
		"PeriodMovieCount":    buildRankStatSortLink(c, currentSort, currentOrder, "period_movie_count", "排行数"),
		"NewEntryMovieCount":  buildRankStatSortLink(c, currentSort, currentOrder, "new_entry_movie_count", "新上榜"),
		"SourceMovieCount":    buildRankStatSortLink(c, currentSort, currentOrder, "source_movie_count", "源影片"),
		"SourceFetchedStartOn": buildRankStatSortLink(
			c, currentSort, currentOrder, "source_fetched_start_on", "抓取开始",
		),
		"SourceFetchedEndOn": buildRankStatSortLink(
			c, currentSort, currentOrder, "source_fetched_end_on", "抓取结束",
		),
		"CalculatedOn": buildRankStatSortLink(c, currentSort, currentOrder, "calculated_on", "计算时间"),
	}
}

func buildRankStatSortLink(c *gin.Context, currentSort, currentOrder, sortField, label string) rankStatSortLink {
	nextOrder := "desc"
	active := currentSort == sortField
	if active && currentOrder == "desc" {
		nextOrder = "asc"
	}

	q := c.Request.URL.Query()
	q.Set("p", "1")
	q.Set("sort", sortField)
	q.Set("order", nextOrder)
	return rankStatSortLink{
		Label:  label,
		Href:   rankStatHref(c.Request.URL.Path, q),
		Active: active,
		Desc:   active && currentOrder == "desc",
	}
}

func buildRankPeriodStatTypeLinks(c *gin.Context, currentType string) []rankPeriodSwitch {
	periodTypes := []int64{
		consts.RankPeriodTypeWeek,
		consts.RankPeriodTypeMonth,
		consts.RankPeriodTypeQuarter,
		consts.RankPeriodTypeYear,
	}
	out := make([]rankPeriodSwitch, 0, len(periodTypes))
	for _, periodType := range periodTypes {
		typeName := consts.RankPeriodTypeName(periodType)
		q := c.Request.URL.Query()
		q.Set("p", "1")
		q.Set("type", typeName)
		out = append(out, rankPeriodSwitch{
			Label:  consts.RankPeriodTypeLabel(periodType),
			Href:   rankStatHref(c.Request.URL.Path, q),
			Active: typeName == currentType,
		})
	}
	return out
}

func rankStatHref(path string, query url.Values) string {
	if encoded := query.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
