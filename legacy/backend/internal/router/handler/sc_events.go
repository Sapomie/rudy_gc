package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/types"
)

type scEventListQuery struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

func (h *MovieHTMLHandler) ListScEventsPage(c *gin.Context) {
	var q scEventListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 24
	}
	q.Sort = normalizeScEventSortField(q.Sort)
	q.Order = normalizeScEventSortOrder(q.Order)

	resp, err := h.scSvc.ListEventPage(c.Request.Context(), int(q.Page), int(q.PageSize), q.Sort, q.Order)
	if err != nil {
		c.String(http.StatusInternalServerError, "加载 SC 事件失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, resp.Total, q.Page, q.PageSize, pageWindow)
	sortQ := buildScEventSortQuery(c, q.Sort, q.Order)
	viewQ := buildScEventViewQuery(c)

	c.HTML(http.StatusOK, "page.sc_event_list", gin.H{
		"Title":       "SC Events",
		"Items":       resp.Items,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ScSortQuery": sortQ,
		"ScViewQuery": viewQ,
	})
}

func (h *MovieHTMLHandler) ListScEventsCardPage(c *gin.Context) {
	var q scEventListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 18
	}
	q.Sort = normalizeScEventSortField(q.Sort)
	q.Order = normalizeScEventSortOrder(q.Order)

	resp, err := h.scSvc.ListEventCardPage(c.Request.Context(), int(q.Page), int(q.PageSize), q.Sort, q.Order)
	if err != nil {
		c.String(http.StatusInternalServerError, "加载 SC 事件失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, resp.Total, q.Page, q.PageSize, pageWindow)
	sortQ := buildScEventSortQuery(c, q.Sort, q.Order)
	viewQ := buildScEventViewQuery(c)

	c.HTML(http.StatusOK, "page.sc_event_cards", gin.H{
		"Title":       "SC Event Cards",
		"Items":       resp.Items,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ScSortQuery": sortQ,
		"ScViewQuery": viewQ,
	})
}

func (h *MovieHTMLHandler) ScEventDetailPage(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.String(http.StatusBadRequest, "缺少参数: name")
		return
	}
	movieFilter := normalizeScEventDetailMovieFilter(c.Query("mf"))

	detail, err := h.scSvc.GetEventDetail(c.Request.Context(), name, movieFilter)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.String(http.StatusNotFound, "未找到 SC 事件: %s", name)
			return
		}
		c.String(http.StatusInternalServerError, "加载 SC 事件失败: %v", err)
		return
	}
	if detail == nil || detail.Event == nil {
		c.String(http.StatusNotFound, "未找到 SC 事件: %s", name)
		return
	}

	c.HTML(http.StatusOK, "page.sc_event_detail", gin.H{
		"Title":                     detail.Event.Name,
		"Event":                     detail.Event,
		"Items":                     detail.Items,
		"FailedItems":               detail.FailedItems,
		"ComeCount":                 detail.ComeCount,
		"EditImageName":             detail.EditImageName,
		"EditMovieCount":            detail.EditMovieCount,
		"EditScMovieCount":          detail.EditScMovieCount,
		"EditComeMovieJavId":        detail.EditComeMovieJavId,
		"EditComeMovieOptions":      detail.EditComeMovieOptions,
		"EditCurrentMovieCastNames": detail.EditCurrentMovieCastNames,
		"MovieFilter":               movieFilter,
		"MovieFilterAllHref":        buildScEventDetailMovieFilterHref(c, "all"),
		"MovieFilterScHref":         buildScEventDetailMovieFilterHref(c, "sc"),
		"MovieFilterNoScHref":       buildScEventDetailMovieFilterHref(c, "nosc"),
	})
}

func normalizeScEventDetailMovieFilter(v string) string {
	switch strings.TrimSpace(v) {
	case "all":
		return "all"
	case "nosc":
		return "nosc"
	default:
		return "sc"
	}
}

func buildScEventDetailMovieFilterHref(c *gin.Context, movieFilter string) string {
	path := c.Request.URL.Path
	query := c.Request.URL.Query()
	query.Set("mf", normalizeScEventDetailMovieFilter(movieFilter))
	encoded := query.Encode()
	if encoded == "" {
		return path
	}
	return path + "?" + encoded
}
