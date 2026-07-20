package dailybest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *DailyBestAPI) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *DailyBestAPI) Ready(c *gin.Context) {
	if err := h.service.Ready(c.Request.Context()); err != nil {
		writeError(c, http.StatusServiceUnavailable, "database_unavailable", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
