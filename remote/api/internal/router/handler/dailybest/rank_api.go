package dailybest

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (h *DailyBestAPI) ListRankSyncManifests(c *gin.Context) {
	rows, err := h.service.ListRankSyncManifests(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_rank_sync_manifests_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestRankSyncManifestListResponse{Manifests: rows})
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
