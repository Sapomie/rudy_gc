package api

import (
	"net/http"

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

// 管理页（按钮页）
func (h *TriggerAPI) Page(c *gin.Context) {
	c.HTML(http.StatusOK, "page.admin_triggers", gin.H{})
}

// === 触发：DailyBest ===
func (h *TriggerAPI) DailyBest(c *gin.Context) {
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcDailyBest})
}

// === 触发：Seeds ===
func (h *TriggerAPI) Seeds(c *gin.Context) {
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcSeeds})
}

// --- 内部统一入队 ---
func (h *TriggerAPI) enqueue(c *gin.Context, msg contracts.TriggerMsg) {
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted) // 202: 已接收后台执行
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "trigger queue is full"})
	}
}
