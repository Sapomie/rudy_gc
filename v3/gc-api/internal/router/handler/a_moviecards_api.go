package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func (h *MovieHTMLHandler) ListMovieCardFullAPI(c *gin.Context) {
	h.movieCardAPI(c,
		types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate},
		"MovieCard",
		"/cards",
	)
}

func (h *MovieHTMLHandler) movieCardAPI(c *gin.Context, base types.ListMovieFullRequest, title, legacyPath string) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = legacyPath
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	req, err := parseMovieCardRequest(c, base)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误: " + err.Error()})
		return
	}

	data, err := h.loadMovieCardPageData(c, req, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":           title,
		"movies":          data.Movies,
		"total":           data.Total,
		"PageInfo":        data.PageInfo,
		"pageInfo":        data.PageInfo,
		"ownedQuery":      data.OwnedQuery,
		"sortQuery":       data.SortQuery,
		"CurrentSort":     data.CurrentSort,
		"MovieCardFilter": data.MovieCardFilter,
	})
}
