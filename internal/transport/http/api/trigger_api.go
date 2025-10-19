// internal/transport/http/api/trigger_api.go
package api

import (
	"net/http"
	"rudy_gc/internal/contracts"
	"strings"

	"rudy_gc/internal/svc"

	"github.com/gin-gonic/gin"
)

type TriggerAPI struct {
	ch chan contracts.TriggerMsg
}

func NewTriggerAPI(deps *svc.Deps) *TriggerAPI {
	return &TriggerAPI{ch: deps.BestTrigger} // 你的 trigger 通道
}

// 管理页（可选：极简 HTML）
func (h *TriggerAPI) Page(c *gin.Context) {
	c.HTML(http.StatusOK, "page.admin_triggers", gin.H{})
}

func (h *TriggerAPI) DailyBest(c *gin.Context) {
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcDailyBest})
}

type seedsReq struct {
	Seeds []string `json:"seeds"` // 也支持 form:"seeds[]" 看你前端
}

func (h *TriggerAPI) Seeds(c *gin.Context) {
	var req seedsReq
	_ = c.ShouldBind(&req) // 允许空，走默认 seeds
	// 去重+清洗
	uniq := make([]string, 0, len(req.Seeds))
	seen := map[string]struct{}{}
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
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcSeeds, Seeds: uniq})
}

func (h *TriggerAPI) Both(c *gin.Context) {
	h.enqueue(c, contracts.TriggerMsg{Kind: contracts.ProcBoth})
}

func (h *TriggerAPI) enqueue(c *gin.Context, msg contracts.TriggerMsg) {
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted) // 202：已接收，后台执行
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "trigger queue is full"})
	}
}
