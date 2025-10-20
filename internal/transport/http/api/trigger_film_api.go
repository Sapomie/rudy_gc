package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/contracts"
	"rudy_gc/internal/svc"
)

type FilmTriggerAPI struct {
	ch chan contracts.FilmTriggerMsg
}

func NewFilmTriggerAPI(deps *svc.Deps) *FilmTriggerAPI {
	return &FilmTriggerAPI{ch: deps.FilmTrigger}
}

// POST /api/triggers/film/rename
func (h *FilmTriggerAPI) Rename(c *gin.Context) {
	h.enqueue(c, contracts.FilmTriggerMsg{Kind: contracts.ProcFilmRename})
}

// POST /api/triggers/film/process
func (h *FilmTriggerAPI) Process(c *gin.Context) {
	h.enqueue(c, contracts.FilmTriggerMsg{Kind: contracts.ProcFilmProcess})
}

func (h *FilmTriggerAPI) enqueue(c *gin.Context, msg contracts.FilmTriggerMsg) {
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted) // 202：接受，后台执行
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "film trigger queue is full"})
	}
}
