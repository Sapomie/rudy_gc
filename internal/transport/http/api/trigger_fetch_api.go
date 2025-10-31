package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/contracts"
	"rudy_gc/internal/svc"
)

type TriggerAPI struct {
	ch chan contracts.TriggerMsg
}

func NewTriggerAPI(deps *svc.Deps) *TriggerAPI {
	return &TriggerAPI{ch: deps.BestTrigger}
}

// 管理页
func (h *TriggerAPI) Page(c *gin.Context) {
	c.HTML(http.StatusOK, "page.admin_triggers", gin.H{})
}

// === DailyBest ===
func (h *TriggerAPI) DailyBest(c *gin.Context) {
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcDailyBest})
}

func (h *TriggerAPI) DailyBestSync(c *gin.Context) { // ✅ 新增
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcSyncBest})
}

// === Seeds ===
func (h *TriggerAPI) Seeds(c *gin.Context) {
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcSeeds})
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

	h.enqueue(c, contracts.TriggerMsg{
		Kind: contracts.ProcSeedByName,
		Name: name,
	})
}

// === 内部通用投递 ===
func (h *TriggerAPI) enqueue(c *gin.Context, msg contracts.TriggerMsg) {
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted) // 202: 已接收后台执行
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "trigger queue is full"})
	}
}
