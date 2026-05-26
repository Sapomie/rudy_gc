package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/svc"
)

type TriggerAPI struct {
	runtime   *loop.FetchLoopService
	scRootDir string
}

func NewTriggerAPI(deps *svc.Deps) *TriggerAPI {
	return &TriggerAPI{
		runtime:   newCrawlerRuntime(deps),
		scRootDir: deps.Config.Film.ScRootDir,
	}
}

// 管理页
func (h *TriggerAPI) Page(c *gin.Context) {
	c.HTML(http.StatusOK, "page.admin_triggers", gin.H{
		"ScRootDir": h.scRootDir,
	})
}

// === DailyBest ===
func (h *TriggerAPI) DailyBest(c *gin.Context) {
	h.startTask(c, loop.StartTaskRequest{TaskType: loop.TaskSpiderDailyBest})
}

func (h *TriggerAPI) DailyBestSync(c *gin.Context) { // ✅ 新增
	h.startTask(c, loop.StartTaskRequest{TaskType: loop.TaskSpiderDailyBestSync})
}

func (h *TriggerAPI) RebuildCastRank(c *gin.Context) {
	h.startTask(c, loop.StartTaskRequest{TaskType: loop.TaskSpiderRebuildCastRank})
}

type actorNameReq struct {
	ActorName string `json:"actorName"`
}

func (h *TriggerAPI) RebuildActorRank(c *gin.Context) {
	var req actorNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	actorName := strings.TrimSpace(req.ActorName)
	if actorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "actorName is required"})
		return
	}

	h.startTask(c, loop.StartTaskRequest{
		TaskType:  loop.TaskSpiderRebuildActorRank,
		ActorName: actorName,
	})
}

// === Seeds ===
func (h *TriggerAPI) Seeds(c *gin.Context) {
	h.startTask(c, loop.StartTaskRequest{TaskType: loop.TaskSpiderSeeds})
}

// === Seed By Name ===
type seedNameReq struct {
	Name string `json:"name"`
}

func (h *TriggerAPI) SeedByName(c *gin.Context) {
	var req seedNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	h.startTask(c, loop.StartTaskRequest{
		TaskType: loop.TaskSpiderSeedByName,
		Name:     name,
	})
}

func (h *TriggerAPI) startTask(c *gin.Context, req loop.StartTaskRequest) {
	jobID, err := h.runtime.StartTask(req)
	if err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeCrawlerJobStarted(c, jobID, req.TaskType)
}

type refreshOldestReq struct {
	Number int64 `json:"number"`
}

func (h *TriggerAPI) RefreshOldestDetail(c *gin.Context) {
	var req refreshOldestReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Number <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid number"})
		return
	}

	h.startTask(c, loop.StartTaskRequest{
		TaskType: loop.TaskSpiderRefreshOldest,
		Number:   req.Number,
	})
}
