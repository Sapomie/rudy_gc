package html

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/domain/vfilm"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type DirectoryHTML struct {
	dirSvc *vfilm.DirectoryService
}

func NewDirectoryHTMLHandler(deps *svc.Deps) *DirectoryHTML {
	return &DirectoryHTML{dirSvc: vfilm.NewDirectoryService(deps)}
}

// GET /dir/root 或 /dir/:id
func (h *DirectoryHTML) DirDetail(c *gin.Context) {
	// 1) 解析参数（沿用 MovieHTML 的节奏）
	recursive := c.DefaultQuery("recursive", "1") == "1"

	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	if page <= 0 {
		page = 1
	}
	size := atoiDefault(c.DefaultQuery("page_size", "24"), defaultPageSize)
	if size <= 0 || size > maxPageSize {
		size = defaultPageSize
	}

	// 排序：页面 query 仍用 sort/order；给“子目录列表”仍用 repo 的 DirSort，
	// 给“影片列表”则用 types.ListDirFilmsRequest 的 SortField(字符串)
	sortField := c.DefaultQuery("sort", "updated_on")
	asc := strings.ToLower(c.DefaultQuery("order", "desc")) == "asc"

	// 2) 目录详情
	var (
		detail *types.DirDetail
		err    error
		dirID  int64
	)

	if c.Param("id") == "root" || c.Param("id") == "" {
		detail, err = h.dirSvc.GetRootDetail(c.Request.Context(), recursive)
		if detail != nil && detail.Directory != nil {
			dirID = detail.Directory.Id
		}
	} else {
		dirID, _ = strconv.ParseInt(c.Param("id"), 10, 64)
		detail, err = h.dirSvc.GetDirDetail(c.Request.Context(), dirID, recursive)
	}

	// 子目录（名称升序，带聚合，最多 60 个）
	children, _, _ := h.dirSvc.ListChildren(
		c.Request.Context(),
		dirID,
		1, 60,
		film_repo.DirSortName,
		true, // asc
		true, // withAgg
	)

	if err != nil {
		c.String(http.StatusInternalServerError, "加载目录失败: %v", err)
		return
	}
	if detail == nil || detail.Directory == nil {
		c.String(http.StatusNotFound, "目录不存在")
		return
	}

	// 3) 电影列表（直属/递归）
	total := int64(0)
	req := &types.ListDirFilmsRequest{
		DirID:     dirID,
		Page:      page,
		PageSize:  size,
		SortField: sortField, // e.g. "updated_on" / "name" / "size" ...
		Asc:       asc,
		Recursive: recursive,
	}
	if mts, tt, e := h.dirSvc.ListFilmsForDirPage(c.Request.Context(), req); e == nil {
		detail.MovieTypes = mts
		total = tt
	}

	// 4) 模板上下文（与现有 partial 兼容）
	q := buildQueryKeep(c, []string{"sort", "order", "page", "page_size"})

	ctx := gin.H{
		"Title":       detail.Directory.Name,
		"Detail":      detail,
		"Recursive":   recursive,
		"CurrentSort": sortField,
		"sortQuery":   gin.H{"sort": sortField, "order": c.DefaultQuery("order", "desc")},
		"Pagination": gin.H{
			"page":      page,
			"page_size": size,
			"total":     total,
			"base_path": "/dir/" + strconv.FormatInt(dirID, 10),
			"query":     q,
		},
		"Query": q,

		// 子目录数据
		"Children": gin.H{
			"Items":   children, // []*types.DirSummary
			"MoreURL": "",
		},
	}

	// 5) 渲染
	c.HTML(http.StatusOK, "page.dir_detail", ctx)
}

/* ================= helpers ================= */

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// 生成类似 "&sort=updated_on&order=desc&page=1&page_size=24"
func buildQueryKeep(c *gin.Context, keys []string) string {
	q := c.Request.URL.Query()
	var sb strings.Builder
	for _, k := range keys {
		if v := q.Get(k); v != "" {
			sb.WriteByte('&')
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	return sb.String()
}
