package html

import (
	"net/http"
	"rudy_gc/internal/consts"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/domain/vfilm"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type DirectoryHTML struct {
	dirSvc *vfilm.DirectoryService
}

func NewDirectoryHTMLHandler(deps *svc.Deps) *DirectoryHTML {
	return &DirectoryHTML{dirSvc: vfilm.NewDirectoryService(deps)}
}

func (h *DirectoryHTML) DirDetail(c *gin.Context) {
	// 1) 构造请求并绑定 query
	var req types.DirPageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	// 2) 从 path 决定是否 root
	if c.Param("id") == "root" || c.Param("id") == "" {
		req.UseRoot = true
	} else {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		req.DirID = id
	}

	// 3) 设置默认值（一次即可）
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > maxPageSize {
		req.PageSize = defaultPageSize
	}
	if req.SortField == "" {
		req.SortField = consts.OrderByBirthTime
	}
	if req.ChildrenPage <= 0 {
		req.ChildrenPage = 1
	}
	if req.ChildrenSize <= 0 {
		req.ChildrenSize = 60
	}
	req.Recursive = true

	// 4) 调 domain
	out, err := h.dirSvc.GetDirectoryPage(c.Request.Context(), &req)
	if err != nil {
		c.String(http.StatusInternalServerError, "加载目录失败: %v", err)
		return
	}
	if out == nil || out.Detail == nil || out.Detail.Directory == nil {
		c.String(http.StatusNotFound, "目录不存在")
		return
	}

	// 5) 模板上下文
	q := buildQueryKeep(c, []string{"od", "asc", "p", "ps"})
	dirID := out.Detail.Directory.Id
	orderStr := "desc"
	if req.Asc {
		orderStr = "asc"
	}

	ctx := gin.H{
		"Title":       out.Detail.Directory.Name,
		"Detail":      out.Detail,
		"Recursive":   req.Recursive,
		"CurrentSort": req.SortField,
		"sortQuery":   gin.H{"sort": req.SortField, "order": orderStr},
		"Pagination": gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
			"total":     out.Total,
			"base_path": "/dir/" + strconv.FormatInt(dirID, 10),
			"query":     q,
		},
		"Query": q,
		"Children": gin.H{
			"Items":   out.Children,
			"MoreURL": "",
		},
	}

	c.HTML(http.StatusOK, "page.dir_detail", ctx)
}

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
