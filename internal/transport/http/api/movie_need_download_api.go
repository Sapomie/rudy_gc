package api

import (
	"net/http"

	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/domain/spider/logic"
	"rudy_gc/internal/svc"

	"github.com/gin-gonic/gin"
)

type MovieAPI struct {
	movieSvc   *movie.MovieService
	crawlLogic *logic.CrawlLogic
}

func NewMovieAPI(deps *svc.Deps) *MovieAPI {
	return &MovieAPI{
		movieSvc:   movie.NewMovieService(deps),
		crawlLogic: logic.NewCrawlLogic(deps),
	}
}

func (h *MovieAPI) AddToDownloadLater(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.movieSvc.AddToDownloadLater(c.Request.Context(), javId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}

func (h *MovieAPI) RemoveFromDownloadLater(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.movieSvc.RemoveFromDownloadLater(c.Request.Context(), javId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}

// ===== 新增：下载当前电影封面 =====
// POST /api/movie/:movie/download-cover
func (h *MovieAPI) DownloadCoverNow(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	if err := h.crawlLogic.DownloadPictureOfMovieByJavId(c.Request.Context(), javId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
