package dailybest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (h *DailyBestAPI) StartFetch(c *gin.Context) {
	var request contracts.DailyBestFetchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求内容错误")
		return
	}
	run, err := h.service.Enqueue(c.Request.Context(), request, dailybesttypes.TriggerManual)
	if err != nil {
		writeError(c, http.StatusBadRequest, "fetch_rejected", err.Error())
		return
	}
	c.JSON(http.StatusAccepted, contracts.DailyBestFetchAcceptedResponse{
		RunID: run.ID, Status: run.Status, Message: "DailyBest 抓取任务已创建",
	})
}
