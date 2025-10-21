// internal/transport/http/api/trigger_api.go（或新文件 sc_trigger_api.go）
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/contracts"
	"rudy_gc/internal/svc"
)

type ScTriggerAPI struct {
	ch chan contracts.ScTriggerMsg
}

func NewScTriggerAPI(deps *svc.Deps) *ScTriggerAPI {
	return &ScTriggerAPI{ch: deps.ScTrigger}
}

// POST /api/triggers/sc/move { "scName": "xxx" }
type scMoveReq struct {
	ScName string `json:"scName" form:"scName"`
}

func (h *ScTriggerAPI) Move(c *gin.Context) {
	var req scMoveReq
	_ = c.ShouldBind(&req)
	name := strings.TrimSpace(req.ScName)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scName required"})
		return
	}
	msg := contracts.ScTriggerMsg{Kind: contracts.ScMove, ScName: name}
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted)
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "sc trigger queue is full"})
	}
}

// POST /api/triggers/sc/add { "dir": "/path/to/scan" }
type scAddReq struct {
	Dir string `json:"dir" form:"dir"`
}

func (h *ScTriggerAPI) Add(c *gin.Context) {
	var req scAddReq
	_ = c.ShouldBind(&req)
	dir := strings.TrimSpace(req.Dir)
	if dir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dir required"})
		return
	}
	msg := contracts.ScTriggerMsg{Kind: contracts.ScAdd, Dir: dir}
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted)
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "sc trigger queue is full"})
	}
}
