package dailybest

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (h *DailyBestAPI) ListRankPeriods(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	periodType, err := parseNonNegativeInt64(c.Query("period_type"), "period_type")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_period_type", err.Error())
		return
	}
	category, err := parseNonNegativeInt64(c.Query("category"), "category")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_category", err.Error())
		return
	}
	rows, total, err := h.service.ListRankPeriods(c.Request.Context(), dailybesttypes.RankPeriodFilter{
		PageQuery: page, PeriodType: periodType, PeriodKey: strings.TrimSpace(c.Query("period_key")), Category: category,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_rank_periods_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestRankPeriodListResponse{Periods: rows, PageInfo: buildPageInfo(total, page)})
}

func (h *DailyBestAPI) ListRankPeriodItems(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	periodID, err := parseNonNegativeInt64(c.Query("period_id"), "period_id")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_period_id", err.Error())
		return
	}
	rows, total, err := h.service.ListRankPeriodItems(c.Request.Context(), dailybesttypes.RankPeriodItemFilter{
		PageQuery: page, PeriodID: periodID, MovieJavID: strings.TrimSpace(c.Query("movie_jav_id")),
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_rank_period_items_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestRankPeriodItemListResponse{Items: rows, PageInfo: buildPageInfo(total, page)})
}

func (h *DailyBestAPI) ListRankPeriodStats(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	periodID, err := parseNonNegativeInt64(c.Query("period_id"), "period_id")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_period_id", err.Error())
		return
	}
	rows, total, err := h.service.ListRankPeriodStats(c.Request.Context(), dailybesttypes.RankPeriodStatFilter{
		PageQuery: page, PeriodID: periodID,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_rank_period_stats_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestRankPeriodStatListResponse{Stats: rows, PageInfo: buildPageInfo(total, page)})
}
