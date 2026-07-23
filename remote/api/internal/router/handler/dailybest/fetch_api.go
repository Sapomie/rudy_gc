package dailybest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (h *DailyBestAPI) StartFetch(c *gin.Context) {
	run, err := h.service.Enqueue(c.Request.Context(), dailybesttypes.TriggerManual)
	if err != nil {
		if errors.Is(err, dailybestservice.ErrActiveFetchRun) {
			writeError(c, http.StatusConflict, "active_fetch_run_exists", "当前业务日期已经存在抓取任务")
			return
		}
		writeError(c, http.StatusBadRequest, "fetch_rejected", err.Error())
		return
	}
	c.JSON(http.StatusAccepted, contracts.DailyBestFetchAcceptedResponse{
		RunID: run.ID, Status: run.Status, Message: "DailyBest 抓取任务已创建",
	})
}
