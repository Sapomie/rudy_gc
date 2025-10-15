package html

import (
	"net/http"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/svc"

	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/types"

	"github.com/gin-gonic/gin"
)

type MovieHTMLHandler struct {
	svc *movie.Service
}

func NewMovieHTMLHandler(deps *svc.Deps) *MovieHTMLHandler {
	return &MovieHTMLHandler{svc: movie.NewMovieService(deps)}
}

func (h *MovieHTMLHandler) ListMovieCardLite(c *gin.Context) {
	var req types.ListMovieLiteRequest
	if err := c.BindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 18
	}

	resp, err := h.svc.ListMovieLite(c.Request.Context(), &req)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}

	// 构建分页信息（window=5 可按需调整）
	pi := BuildPageInfo(c, resp.Total, req.Page, req.PageSize, 5)
	ownedQ := buildOwnedFilterInfo(c)

	data := gin.H{
		"Title":      "MovieCard",
		"movies":     resp.List,
		"Total":      resp.Total,
		"PageInfo":   pi,
		"ownedQuery": ownedQ,     // 供 owned_filter 使用
		"total":      resp.Total, // 你的旧模板里直接用了 .total / .fieldName
		"fieldName":  "Movies",   // 按需替换
	}
	c.HTML(http.StatusOK, "base", data)

}

func (h *MovieHTMLHandler) ListMovieCardFull(c *gin.Context) {
	req := types.ListMovieFullRequest{
		OrderBy: consts.OrderByDetailUpdateTime,
	}

	if err := c.BindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 18
	}

	resp, err := h.svc.ListMovieFull(c.Request.Context(), &req)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}

	// 构建分页信息（window=5 可按需调整）
	pi := BuildPageInfo(c, resp.Total, req.Page, req.PageSize, 5)
	ownedQ := buildOwnedFilterInfo(c)

	data := gin.H{
		"Title":      "MovieCard",
		"movies":     resp.List,
		"Total":      resp.Total,
		"PageInfo":   pi,
		"ownedQuery": ownedQ,     // 供 owned_filter 使用
		"total":      resp.Total, // 你的旧模板里直接用了 .total / .fieldName
		"fieldName":  "Movies",   // 按需替换
	}
	c.HTML(http.StatusOK, "base", data)

}

func (h *MovieHTMLHandler) ListMovieCardHasRank(c *gin.Context) {
	req := types.ListMovieFullRequest{
		OrderBy: consts.OrderByRankDate,
	}

	if err := c.BindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 18
	}

	resp, err := h.svc.ListMovieFull(c.Request.Context(), &req)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}

	// 构建分页信息（window=5 可按需调整）
	pi := BuildPageInfo(c, resp.Total, req.Page, req.PageSize, 5)
	ownedQ := buildOwnedFilterInfo(c)

	data := gin.H{
		"Title":      "MovieCard",
		"movies":     resp.List,
		"Total":      resp.Total,
		"PageInfo":   pi,
		"ownedQuery": ownedQ,     // 供 owned_filter 使用
		"total":      resp.Total, // 你的旧模板里直接用了 .total / .fieldName
		"fieldName":  "Movies",   // 按需替换
	}
	c.HTML(http.StatusOK, "base", data)

}

func (h *MovieHTMLHandler) ListMovieCardOwned(c *gin.Context) {
	req := types.ListMovieFullRequest{
		Owned:   consts.OwnedAllNotRemoved,
		OrderBy: consts.OrderByBirthTime,
	}

	if err := c.BindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 18
	}

	resp, err := h.svc.ListMovieFull(c.Request.Context(), &req)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}

	// 构建分页信息（window=5 可按需调整）
	pi := BuildPageInfo(c, resp.Total, req.Page, req.PageSize, 5)
	ownedQ := buildOwnedFilterInfo(c)

	data := gin.H{
		"Title":      "MovieCard",
		"movies":     resp.List,
		"Total":      resp.Total,
		"PageInfo":   pi,
		"ownedQuery": ownedQ,     // 供 owned_filter 使用
		"total":      resp.Total, // 你的旧模板里直接用了 .total / .fieldName
		"fieldName":  "Movies",   // 按需替换
	}
	c.HTML(http.StatusOK, "base", data)

}
