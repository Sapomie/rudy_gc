package handler

import (
	"net/http"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/movie"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/wdir"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type WDirectoryHTML struct {
	dirSvc     *wdir.DirectoryService
	movieSvc   *movie.Service
	detailJobs chan string
}

func NewWDirectoryHTMLHandler(deps *svc.Deps) *WDirectoryHTML {
	return &WDirectoryHTML{
		dirSvc:     wdir.NewDirectoryService(deps),
		movieSvc:   movie.NewService(deps),
		detailJobs: deps.DetailJobs,
	}
}

func (h *WDirectoryHTML) RootList(c *gin.Context) {
	if rawID := strings.TrimSpace(c.Query("id")); rawID != "" && rawID != "root" {
		h.DirDetail(c)
		return
	}

	var req struct {
		Page     int64 `form:"p"`
		PageSize int64 `form:"ps"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > maxPageSize {
		req.PageSize = 60
	}

	items, total, err := h.dirSvc.ListRootPage(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		c.String(http.StatusInternalServerError, "加载目录失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, total, req.Page, req.PageSize, pageWindow)
	ctx := gin.H{
		"Title":    "媒体目录",
		"Total":    total,
		"PageInfo": pi,
		"pageInfo": pi,
		"Pagination": gin.H{
			"PageInfo": pi,
		},
		"Children": gin.H{
			"Title":   "根目录",
			"Items":   items,
			"MoreURL": "",
		},
	}

	c.HTML(http.StatusOK, "page.wdir_list", ctx)
}

func (h *WDirectoryHTML) DirDetail(c *gin.Context) {
	rawID := strings.TrimSpace(c.Query("id"))
	if rawID == "" || rawID == "root" {
		target := "/wdir"
		if raw := strings.TrimSpace(c.Request.URL.RawQuery); raw != "" {
			target += "?" + raw
		}
		c.Redirect(http.StatusFound, target)
		return
	}

	dirID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || dirID <= 0 {
		c.String(http.StatusNotFound, "目录不存在")
		return
	}

	var req types.DirPageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	req.DirID = dirID
	if _, ok := c.GetQuery("recursive"); !ok {
		req.Recursive = true
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > maxPageSize {
		req.PageSize = defaultPageSize
	}
	if req.SortField == "" {
		req.SortField = consts.OrderByMediaBirthTime
	}
	if req.ChildrenPage <= 0 {
		req.ChildrenPage = 1
	}
	if req.ChildrenSize <= 0 {
		req.ChildrenSize = 60
	}

	out, err := h.dirSvc.GetDirectoryPage(c.Request.Context(), &req)
	if err != nil {
		c.String(http.StatusInternalServerError, "加载目录失败: %v", err)
		return
	}
	if out == nil || out.Detail == nil || out.Detail.Directory == nil {
		c.String(http.StatusNotFound, "目录不存在")
		return
	}

	q := buildWDirToggleQuery(c)

	baseReq := types.ListMovieFullRequest{
		MediaOwned: consts.OwnedAllNotRemoved,
		OrderBy:    consts.OrderByMediaBirthTime,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	cardReq, err := parseMovieCardRequest(c, baseReq)
	if err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	cardReq.MediaDirFull = strings.TrimSpace(out.Detail.Directory.Path)
	cardReq.MediaDirSub = req.Recursive

	cardResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), &cardReq)
	if err != nil {
		c.String(http.StatusInternalServerError, "加载目录失败: %v", err)
		return
	}
	h.enqueueJavIDsNonBlocking(cardResp.JavIds)

	cardPageInfo := BuildPageInfo(c, cardResp.Total, cardReq.Page, cardReq.PageSize, pageWindow)
	cardFilter := buildMovieCardFilterView(c, cardReq, cardReq.OrderBy, nil)
	cardFilter.HideDirs = true
	cardFilter.Action = buildWDirFilterAction(dirID, req.Recursive)
	cardFilter.ClearHref = buildWDirFilterAction(dirID, req.Recursive)

	ctx := gin.H{
		"Title":           out.Detail.Directory.Name,
		"Detail":          out.Detail,
		"Recursive":       req.Recursive,
		"CurrentSort":     cardReq.OrderBy,
		"SortQuery":       buildSortQuery(c, cardReq.OrderBy),
		"sortQuery":       buildSortQuery(c, cardReq.OrderBy),
		"Movies":          cardResp.List,
		"movies":          cardResp.List,
		"Total":           cardResp.Total,
		"total":           cardResp.Total,
		"PageInfo":        cardPageInfo,
		"pageInfo":        cardPageInfo,
		"ownedQuery":      buildOwnedFilterInfoWithDefaults(c, "", "3"),
		"MovieCardFilter": cardFilter,
		"Pagination": gin.H{
			"PageInfo": cardPageInfo,
		},
		"Query": q,
		"Children": gin.H{
			"Title":   "子目录",
			"Items":   out.Children,
			"MoreURL": "",
		},
	}

	c.HTML(http.StatusOK, "page.wdir_detail", ctx)
}

func buildWDirFilterAction(dirID int64, recursive bool) string {
	path := "/wdir?id=" + strconv.FormatInt(dirID, 10)
	if recursive {
		return path + "&recursive=1"
	}
	return path + "&recursive=0"
}

func buildWDirToggleQuery(c *gin.Context) string {
	q := cloneValues(c)
	q.Del("id")
	q.Del("recursive")
	if enc := q.Encode(); enc != "" {
		return "&" + enc
	}
	return ""
}

func (h *WDirectoryHTML) enqueueJavIDsNonBlocking(ids []string) {
	if h.detailJobs == nil || len(ids) == 0 {
		return
	}
	for _, id := range uniqueNonEmpty(ids) {
		select {
		case h.detailJobs <- id:
		default:
			continue
		}
	}
}
