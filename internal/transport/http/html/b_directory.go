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
	// 1) 绑定查询参数
	var req types.DirPageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	// 2) 从 path 决定是否 root
	if c.Param("id") == "root" || c.Param("id") == "" {
		req.UseRoot = true
	} else {
		if id, err := strconv.ParseInt(c.Param("id"), 10, 64); err == nil {
			req.DirID = id
		}
	}

	// 3) 默认值（仅此一处）
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

	// 5) 分页信息（⚠️ 用 int64 形参）
	pi := BuildPageInfo(c, out.Total, req.Page, req.PageSize, pageWindow)

	// 6) 模板上下文
	q := buildQueryKeep(c, []string{"od", "asc", "p", "ps"})
	// 🔹 新增：目录页排序栏对象（基于当前 od）
	sortQ := buildDirSortQuery(c, req.SortField)

	ctx := gin.H{
		"Title":       out.Detail.Directory.Name,
		"Detail":      out.Detail,
		"Recursive":   req.Recursive,
		"CurrentSort": req.SortField,
		"sortQuery":   gin.H{"sort": req.SortField},

		// 供 pagination partial 使用
		"PageInfo": pi,
		"pageInfo": pi,
		"Pagination": gin.H{
			"PageInfo": pi,
		},
		"Sort": sortQ,

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
