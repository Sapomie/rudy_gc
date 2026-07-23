package dailybest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (h *DailyBestAPI) ListRankDays(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	filter, err := parseRankDayFilter(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_rank_day_filter", err.Error())
		return
	}
	if filter.DayNumber == 0 && !page.Enabled {
		writeError(c, http.StatusBadRequest, "pagination_required", "rank-days 列表必须提供 p 和 ps")
		return
	}
	filter.PageQuery = page
	rows, total, err := h.service.ListRankDays(c.Request.Context(), filter)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_rank_days_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestRankDayListResponse{Days: rows, PageInfo: buildPageInfo(total, page)})
}

func parseRankDayFilter(c *gin.Context) (dailybesttypes.RankDayFilter, error) {
	dayNumber, err := parseNonNegativeInt64(c.Query("day_number"), "day_number")
	if err != nil {
		return dailybesttypes.RankDayFilter{}, err
	}
	startDayNumber, err := parseRankBusinessDate(c.Query("business_date_start"), "business_date_start")
	if err != nil {
		return dailybesttypes.RankDayFilter{}, err
	}
	endDayNumber, err := parseRankBusinessDate(c.Query("business_date_end"), "business_date_end")
	if err != nil {
		return dailybesttypes.RankDayFilter{}, err
	}
	if startDayNumber > 0 && endDayNumber > 0 && startDayNumber > endDayNumber {
		return dailybesttypes.RankDayFilter{}, fmt.Errorf("business_date_start 不能晚于 business_date_end")
	}
	return dailybesttypes.RankDayFilter{
		DayNumber: dayNumber, StartDayNumber: startDayNumber, EndDayNumber: endDayNumber,
	}, nil
}

func parseRankBusinessDate(raw, field string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	dayNumber, err := contracts.RankDayNumber(raw)
	if err != nil || dayNumber <= 0 {
		return 0, fmt.Errorf("%s 必须为 %s 及之后的 YYYY-MM-DD", field, contracts.DailyBestGenesisDate)
	}
	return dayNumber, nil
}

func (h *DailyBestAPI) ListRanks(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	dayNumber, err := parseDayNumber(c.Query("day_number"), c.Query("business_date"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_rank_date", err.Error())
		return
	}
	category, err := parseNonNegativeInt64(c.Query("category"), "category")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_category", err.Error())
		return
	}
	rankPos, err := parseNonNegativeInt64(c.Query("rank_pos"), "rank_pos")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_rank_pos", err.Error())
		return
	}
	rows, total, err := h.service.ListRanks(c.Request.Context(), dailybesttypes.RankFilter{
		PageQuery: page, DayNumber: dayNumber, Category: category, RankPos: rankPos,
		MovieJavID: strings.TrimSpace(c.Query("movie_jav_id")),
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_ranks_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestRankListResponse{Ranks: rows, PageInfo: buildPageInfo(total, page)})
}

func (h *DailyBestAPI) ListRankDayStats(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	rankDate := strings.TrimSpace(c.Query("rank_date"))
	dayNumber, err := parseDayNumber(c.Query("day_number"), rankDate)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_rank_date", err.Error())
		return
	}
	rows, total, err := h.service.ListRankDayStats(c.Request.Context(), dailybesttypes.RankDayStatFilter{
		PageQuery: page, DayNumber: dayNumber, RankDate: rankDate,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_rank_day_stats_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestRankDayStatListResponse{Stats: rows, PageInfo: buildPageInfo(total, page)})
}
