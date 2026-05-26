package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/svc"
)

type crawlerJobActionResponse struct {
	JobID    string `json:"job_id,omitempty"`
	TaskType string `json:"task_type,omitempty"`
	Message  string `json:"message,omitempty"`
}

func newCrawlerRuntime(deps *svc.Deps) *loop.FetchLoopService {
	return loop.NewFetchLoopService(deps)
}

func writeCrawlerError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": strings.TrimSpace(message)})
}

func writeCrawlerJobStarted(c *gin.Context, jobID, taskType string) {
	c.JSON(http.StatusAccepted, crawlerJobActionResponse{
		JobID:    jobID,
		TaskType: taskType,
		Message:  "任务已启动",
	})
}

func writeCrawlerJobAction(c *gin.Context, jobID, message string) {
	c.JSON(http.StatusOK, crawlerJobActionResponse{
		JobID:   jobID,
		Message: message,
	})
}
