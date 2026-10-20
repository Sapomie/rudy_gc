package dailybest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/router/handler/common"
	dailybestsync "rudy_gc/local/api/internal/service/dailybest"
	"rudy_gc/local/api/internal/service/loop"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type RankDayAPI struct {
	service *dailybestsync.Service
	runtime *loop.FetchLoopService
}

func NewRankDayAPI(deps *dep.Dep) *RankDayAPI {
	return &RankDayAPI{
		service: deps.DailyBestSync,
		runtime: common.NewCrawlerRuntime(deps),
	}
}

func (h *RankDayAPI) List(c *gin.Context) {
	defer common.WithFrontendPath(c, "/remotedailybestrankdaystats")()
	query, err := parseRankDayQuery(c)
	if err != nil {
		common.WriteCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.service.ListRankDays(c.Request.Context(), query)
	if err != nil {
		common.WriteCrawlerError(c, http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Days":  result.Days,
		"Total": result.Total,
		"Query": gin.H{
			"BusinessDateStart": query.BusinessDateStart,
			"BusinessDateEnd":   query.BusinessDateEnd,
			"Page":              result.Page,
			"PageSize":          result.PageSize,
		},
		"PageInfo": common.BuildPageInfo(c, result.Total, result.Page, result.PageSize, common.PageWindow),
	})
}

func (h *RankDayAPI) Sync(c *gin.Context) {
	var request dailybesttypes.SyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.WriteCrawlerError(c, http.StatusBadRequest, "同步请求格式错误")
		return
	}
	jobID, err := h.runtime.StartTask(loop.StartTaskRequest{
		TaskType:         loop.TaskSpiderDailyBestSync,
		AutoFetchSite:    "false",
		DailyBestTargets: request.Targets,
	})
	if err != nil {
		common.WriteCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	common.WriteCrawlerJobStarted(c, jobID, loop.TaskSpiderDailyBestSync)
}

func parseRankDayQuery(c *gin.Context) (dailybesttypes.RankDayQuery, error) {
	page, err := parsePositiveQuery(c.Query("p"), 1, "p")
	if err != nil {
		return dailybesttypes.RankDayQuery{}, err
	}
	pageSize, err := parsePositiveQuery(c.Query("ps"), 24, "ps")
	if err != nil {
		return dailybesttypes.RankDayQuery{}, err
	}
	dayNumber, err := parsePositiveQuery(c.Query("day_number"), 0, "day_number")
	if err != nil {
		return dailybesttypes.RankDayQuery{}, err
	}
	start := strings.TrimSpace(c.Query("business_date_start"))
	end := strings.TrimSpace(c.Query("business_date_end"))
	startDay, err := parseBusinessDate(start, "business_date_start")
	if err != nil {
		return dailybesttypes.RankDayQuery{}, err
	}
	endDay, err := parseBusinessDate(end, "business_date_end")
	if err != nil {
		return dailybesttypes.RankDayQuery{}, err
	}
	if startDay > 0 && endDay > 0 && startDay > endDay {
		return dailybesttypes.RankDayQuery{}, &queryError{message: "开始日期不能晚于结束日期"}
	}
	return dailybesttypes.RankDayQuery{
		BusinessDateStart: start, BusinessDateEnd: end, DayNumber: dayNumber, Page: page, PageSize: pageSize,
	}, nil
}

func parsePositiveQuery(raw string, fallback int64, field string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, &queryError{message: field + " 必须为正整数"}
	}
	return value, nil
}

func parseBusinessDate(raw, field string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	dayNumber, err := contracts.RankDayNumber(raw)
	if err != nil || dayNumber <= 0 {
		return 0, &queryError{message: field + " 必须为有效日期"}
	}
	return dayNumber, nil
}

type queryError struct {
	message string
}

func (e *queryError) Error() string {
	return e.message
}
