package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/spider"
	"rudy_gc/internal/svc"

	"github.com/gin-gonic/gin"
)

type MovieAPI struct {
	movieSvc   *movie.Service
	crawlLogic *spider.CrawlLogic
	deps       *svc.Deps
}

func NewMovieAPI(deps *svc.Deps) *MovieAPI {
	return &MovieAPI{
		movieSvc:   movie.NewService(deps),
		crawlLogic: spider.NewCrawlLogic(deps),
		deps:       deps,
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

type addMovieCastReq struct {
	Name string `json:"name"`
}

// POST /api/movie/:movie/add-cast
func (h *MovieAPI) AddCast(c *gin.Context) {
	javId := strings.TrimSpace(c.Param("movie"))
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	mv, err := h.deps.MovieModel.FindOneByJavId(c.Request.Context(), javId)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "影片不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if mv == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "影片不存在"})
		return
	}

	var req addMovieCastReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "演员名不能为空"})
		return
	}

	castRow, err := h.deps.CastModel.FindOneByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "演员不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if castRow == nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "演员不存在"})
		return
	}

	now := time.Now().Unix()
	if _, err := h.deps.MovieCastModel.FindOneByMovieJavIdCastId(c.Request.Context(), javId, castRow.Id); err != nil {
		if !errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
		if _, err := h.deps.MovieCastModel.Insert(c.Request.Context(), &moviex.AmrMovieCast{
			MovieJavId: javId,
			CastId:     castRow.Id,
			CreatedOn:  now,
			UpdatedOn:  now,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
	}

	movieNumber, ownedMovieNumber, err := h.deps.CastModel.GetMovieNumbersByID(c.Request.Context(), castRow.Id, consts.OwnedAllNotRemoved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	castRow.MovieNumber = movieNumber
	castRow.OwnedMovieNumber = ownedMovieNumber
	castRow.UpdatedOn = now
	if err := h.deps.CastModel.Update(c.Request.Context(), castRow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if err := h.deps.SyncPersonStatsByIDs(c.Request.Context(), []int64{castRow.PersonId}, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	h.movieSvc.InvalidateMovieType(c.Request.Context(), javId)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
