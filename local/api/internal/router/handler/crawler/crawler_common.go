package crawler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type crawlerJobActionResponse struct {
	JobID   string `json:"job_id,omitempty"`
	Message string `json:"message,omitempty"`
}

func writeCrawlerJobAction(c *gin.Context, jobID, message string) {
	c.JSON(http.StatusOK, crawlerJobActionResponse{
		JobID:   jobID,
		Message: message,
	})
}
