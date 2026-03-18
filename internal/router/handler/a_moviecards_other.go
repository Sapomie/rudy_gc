package handler

import (
	"net/http"
	"rudy_gc/internal/types"

	"github.com/gin-gonic/gin"
)

func (h *MovieHTMLHandler) ListMovieCardRandomPick(c *gin.Context) {
	req := types.ListMovieFullRequest{
		Owned:    3,
		Page:     1,
		PageSize: 10000,
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	curOD := normalizeOrderBy(c.Query("od"), req.OrderBy)
	req.OrderBy = curOD
	movieTypes, err := h.scSvc.PickFromSource(c, &req, 30)
	if err != nil {
		c.String(http.StatusBadRequest, "PickFromSource err: %v", err)
		return
	}
	total := len(movieTypes)

	pi := BuildPageInfo(c, int64(total), req.Page, req.PageSize, pageWindow)
	ownedQ := buildOwnedFilterInfo(c)
	sortQ := buildSortQuery(c, curOD)

	c.HTML(http.StatusOK, "page.list_movie_card", gin.H{
		"Title":       "MovieCard",
		"movies":      movieTypes,
		"Total":       total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ownedQuery":  ownedQ,
		"sortQuery":   sortQ,
		"CurrentSort": curOD,
	})
}
