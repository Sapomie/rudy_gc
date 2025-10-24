package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/domain/vfilm"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/svc"
)

type DirectoryAPI struct {
	dirSvc *vfilm.DirectoryService
}

func NewDirectoryAPI(deps *svc.Deps) *DirectoryAPI {
	return &DirectoryAPI{
		dirSvc: vfilm.NewDirectoryService(deps),
	}
}

// GET /api/dirs/root?recursive=1
// 你的根只有一个：直接返回根目录详情（含统计），而不是列表。
func (h *DirectoryAPI) GetRootDetail(c *gin.Context) {
	recursive := c.DefaultQuery("recursive", "1") == "1"
	detail, err := h.dirSvc.GetRootDetail(c.Request.Context(), recursive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": detail})
}

// GET /api/dirs/:id?recursive=1
func (h *DirectoryAPI) GetDirDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "非法的目录ID"})
		return
	}
	recursive := c.DefaultQuery("recursive", "1") == "1"

	detail, err := h.dirSvc.GetDirDetail(c.Request.Context(), id, recursive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "目录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": detail})
}

// GET /api/dirs/:id/children?agg=1&page=1&page_size=24&sort=name&order=asc
func (h *DirectoryAPI) ListChildren(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "非法的目录ID"})
		return
	}
	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	size := atoiDefault(c.DefaultQuery("page_size", "24"), 24)
	if size <= 0 || size > 200 {
		size = 24
	}
	sort := toDirSort(c.DefaultQuery("sort", "name"))
	asc := strings.ToLower(c.DefaultQuery("order", "asc")) == "asc"
	withAgg := c.DefaultQuery("agg", "1") == "1"

	items, total, err := h.dirSvc.ListChildren(c.Request.Context(), id, page, size, sort, asc, withAgg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	totalPages := int((total + int64(size) - 1) / int64(size))
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"page":        page,
		"page_size":   size,
		"total":       total,
		"total_pages": totalPages,
		"items":       items,
	})
}

// GET /api/dirs/:id/siblings
func (h *DirectoryAPI) ListSiblings(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "非法的目录ID"})
		return
	}
	items, err := h.dirSvc.ListSiblings(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "items": items})
}

// GET /api/dirs/:id/breadcrumbs
func (h *DirectoryAPI) GetBreadcrumbs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "非法的目录ID"})
		return
	}
	items, err := h.dirSvc.GetBreadcrumbs(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "items": items})
}

// GET /api/dirs/:id/stats?recursive=1&bucket=none|month|year
func (h *DirectoryAPI) GetDirStats(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "非法的目录ID"})
		return
	}
	recursive := c.DefaultQuery("recursive", "1") == "1"
	var bucket film_repo.BucketKind
	switch strings.ToLower(c.DefaultQuery("bucket", "none")) {
	case "month":
		bucket = film_repo.BucketMonth
	case "year":
		bucket = film_repo.BucketYear
	default:
		bucket = film_repo.BucketNone
	}

	stats, err := h.dirSvc.GetDirStats(c.Request.Context(), id, recursive, bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": stats})
}

// 路由注册（放到 router 里调用）
func (h *DirectoryAPI) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/api/dirs")
	{
		g.GET("/root", h.GetRootDetail) // 单根：直接返回根详情
		g.GET("/:id", h.GetDirDetail)
		g.GET("/:id/children", h.ListChildren)
		g.GET("/:id/siblings", h.ListSiblings)
		g.GET("/:id/breadcrumbs", h.GetBreadcrumbs)
		g.GET("/:id/stats", h.GetDirStats)
	}
}

// ---- helpers ----
func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func toDirSort(s string) film_repo.DirSort {
	switch strings.ToLower(s) {
	case "updated_on":
		return film_repo.DirSortUpdatedOn
	default:
		return film_repo.DirSortName
	}
}
