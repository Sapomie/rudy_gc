package api

import (
	"net/http"
	"rudy_gc/internal/domain/movie"

	"rudy_gc/internal/svc"

	"github.com/gin-gonic/gin"
)

type MovieAPI struct {
	svc *movie.Service
}

func NewAPI(deps *svc.Deps) *MovieAPI {
	return &MovieAPI{svc: movie.NewMovieService(deps)}
}

// 添加“稍后下载”
func (h *MovieAPI) Add(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.svc.AddToDownloadLater(c.Request.Context(), javId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}

// 移除“稍后下载”
func (h *MovieAPI) Remove(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.svc.RemoveFromDownloadLater(c.Request.Context(), javId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}
