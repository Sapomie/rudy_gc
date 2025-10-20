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

// 管理页（按钮页）
func (h *TriggerAPI) Page(c *gin.Context) {
	c.HTML(http.StatusOK, "page.admin_triggers", gin.H{})
}

// === 触发：DailyBest ===
func (h *TriggerAPI) DailyBest(c *gin.Context) {
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcDailyBest})
}

// === 触发：Seeds（可带 seeds 参数）===
type seedsReq struct {
	Seeds []string `json:"seeds"`
}

func (h *TriggerAPI) Seeds(c *gin.Context) {
	var req seedsReq
	_ = c.ShouldBind(&req)

	seen := make(map[string]struct{}, len(req.Seeds))
	uniq := make([]string, 0, len(req.Seeds))
	for _, s := range req.Seeds {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		uniq = append(uniq, s)
	}

	h.enqueue(c, contracts.TriggerMsg{
		Kind:  contracts.ProcSeeds,
		Seeds: uniq,
	})
}

func (h *TriggerAPI) enqueue(c *gin.Context, msg contracts.TriggerMsg) {
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted) // 202
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "trigger queue is full"})
	}
}
