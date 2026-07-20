package dailybest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	contracts "rudy_gc/shared/dailybest"
)

func (h *DailyBestAPI) ListFetchRuns(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	runs, total, err := h.service.ListFetchRuns(c.Request.Context(), page)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_fetch_runs_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestFetchRunListResponse{Runs: runs, PageInfo: buildPageInfo(total, page)})
}

func (h *DailyBestAPI) GetFetchRun(c *gin.Context) {
	runID, err := parsePositiveInt64(c.Param("id"), "id")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_run_id", err.Error())
		return
	}
	run, err := h.service.FindFetchRun(c.Request.Context(), runID)
	if err != nil {
		writeServiceError(c, "fetch_run", err)
		return
	}
	c.JSON(http.StatusOK, run)
}
