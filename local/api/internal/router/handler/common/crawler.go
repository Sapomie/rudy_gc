package common

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/loop"
)

func NewCrawlerRuntime(deps *dep.Dep) *loop.FetchLoopService {
	return loop.NewFetchLoopService(deps)
}

func WriteCrawlerError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": strings.TrimSpace(message)})
}

func WriteCrawlerJobStarted(c *gin.Context, jobID, taskType string) {
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":    jobID,
		"task_type": taskType,
		"message":   "任务已启动",
	})
}
