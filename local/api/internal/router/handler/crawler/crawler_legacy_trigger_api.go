package crawler

import (
	"net/http"
	"rudy_gc/local/api/internal/router/handler/common"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/service/loop"
)

func (h *CrawlerAPI) DailyBest(c *gin.Context) {
	h.startLegacyTask(c, loop.StartTaskRequest{TaskType: loop.TaskSpiderDailyBest})
}

func (h *CrawlerAPI) RebuildCastRank(c *gin.Context) {
	h.startLegacyTask(c, loop.StartTaskRequest{TaskType: loop.TaskSpiderRebuildCastRank})
}

func (h *CrawlerAPI) Seeds(c *gin.Context) {
	h.startLegacyTask(c, loop.StartTaskRequest{TaskType: loop.TaskSpiderSeeds})
}

type legacySeedNameReq struct {
	Name string `json:"name"`
}

func (h *CrawlerAPI) SeedByName(c *gin.Context) {
	var req legacySeedNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	h.startLegacyTask(c, loop.StartTaskRequest{
		TaskType: loop.TaskSpiderSeedByName,
		Name:     name,
	})
}

type legacyRefreshOldestReq struct {
	Number int64 `json:"number"`
}

func (h *CrawlerAPI) RefreshOldestDetail(c *gin.Context) {
	var req legacyRefreshOldestReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Number <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid number"})
		return
	}

	h.startLegacyTask(c, loop.StartTaskRequest{
		TaskType: loop.TaskSpiderRefreshOldest,
		Number:   req.Number,
	})
}

func (h *CrawlerAPI) startLegacyTask(c *gin.Context, req loop.StartTaskRequest) {
	jobID, err := h.runtime.StartTask(req)
	if err != nil {
		common.WriteCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	common.WriteCrawlerJobStarted(c, jobID, strings.TrimSpace(req.TaskType))
}
