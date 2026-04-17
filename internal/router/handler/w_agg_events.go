package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/model/modelx/moviex"
)

type wAggEventListQuery struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
	AggKey   string `form:"agg_key"`
	FlowKey  string `form:"flow_key"`
	Status   string `form:"status"`
}

type wAggEventListRow struct {
	StartedTime         int64
	AggKey              string
	FlowKey             string
	ScopeCount          int64
	BucketCount         int64
	TopCount            int64
	FinishedTime        int64
	DurationSecondsText string
}

type wAggEventSortLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

type wAggEventHeaderLinks struct {
	StartedTime  wAggEventSortLink
	AggKey       wAggEventSortLink
	FlowKey      wAggEventSortLink
	ScopeCount   wAggEventSortLink
	BucketCount  wAggEventSortLink
	TopCount     wAggEventSortLink
	FinishedTime wAggEventSortLink
	Duration     wAggEventSortLink
}

func (h *MovieHTMLHandler) WAggEventListPage(c *gin.Context) {
	var q wAggEventListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	q.Sort = normalizeWAggEventSort(q.Sort)
	q.Order = normalizeWAggEventOrder(q.Order)
	q.AggKey = strings.TrimSpace(q.AggKey)
	q.FlowKey = strings.TrimSpace(q.FlowKey)
	q.Status = strings.TrimSpace(q.Status)

	rows, total, err := h.deps.WAggEventModel.ListPage(c.Request.Context(), moviex.WAggEventListFilter{
		AggKey:   q.AggKey,
		FlowKey:  q.FlowKey,
		Status:   q.Status,
		Sort:     q.Sort,
		Order:    q.Order,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "w_agg_event 列表加载失败: %v", err)
		return
	}

	pageInfo := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	c.HTML(http.StatusOK, "page.w_agg_event_list", gin.H{
		"Title":       "WAggEvents",
		"PageTitle":   "WAggEvent 列表",
		"PageNote":    "直接查看 w_agg_event 落库记录。",
		"Rows":        buildWAggEventListRows(rows),
		"Total":       total,
		"PageInfo":    pageInfo,
		"pageInfo":    pageInfo,
		"Query":       q,
		"HeaderLinks": buildWAggEventHeaderLinks(c, q.Sort, q.Order),
		"ClearHref":   "/w-agg-events",
		"MenuListID":  "w_agg_events",
	})
}

func buildWAggEventListRows(rows []*moviex.WAggEvent) []*wAggEventListRow {
	if len(rows) == 0 {
		return []*wAggEventListRow{}
	}

	out := make([]*wAggEventListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &wAggEventListRow{
			StartedTime:         row.StartedTime,
			AggKey:              row.AggKey,
			FlowKey:             row.FlowKey,
			ScopeCount:          row.ScopeCount,
			BucketCount:         row.BucketCount,
			TopCount:            row.TopCount,
			FinishedTime:        row.FinishedTime,
			DurationSecondsText: formatWAggEventDurationSeconds(row.DurationMs),
		})
	}
	return out
}

func formatWAggEventDurationSeconds(durationMs int64) string {
	if durationMs <= 0 {
		return "-"
	}
	if durationMs%1000 == 0 {
		return fmt.Sprintf("%d", durationMs/1000)
	}
	seconds := float64(durationMs) / 1000.0
	text := fmt.Sprintf("%.1f", seconds)
	text = strings.TrimRight(text, "0")
	return strings.TrimRight(text, ".")
}

func normalizeWAggEventSort(sortField string) string {
	switch strings.TrimSpace(sortField) {
	case "scope_count":
		return "scope_count"
	case "bucket_count":
		return "bucket_count"
	case "top_count":
		return "top_count"
	case "finished_time":
		return "finished_time"
	case "duration_ms":
		return "duration_ms"
	case "agg_key":
		return "agg_key"
	case "flow_key":
		return "flow_key"
	default:
		return "started_time"
	}
}

func normalizeWAggEventOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "asc"
	}
	return "desc"
}

func buildWAggEventHeaderLinks(c *gin.Context, currentSort string, currentOrder string) wAggEventHeaderLinks {
	return wAggEventHeaderLinks{
		StartedTime:  buildWAggEventSortLink(c, currentSort, currentOrder, "started_time", "StartedTime"),
		AggKey:       buildWAggEventSortLink(c, currentSort, currentOrder, "agg_key", "AggKey"),
		FlowKey:      buildWAggEventSortLink(c, currentSort, currentOrder, "flow_key", "FlowKey"),
		ScopeCount:   buildWAggEventSortLink(c, currentSort, currentOrder, "scope_count", "ScopeCount"),
		BucketCount:  buildWAggEventSortLink(c, currentSort, currentOrder, "bucket_count", "BucketCount"),
		TopCount:     buildWAggEventSortLink(c, currentSort, currentOrder, "top_count", "TopCount"),
		FinishedTime: buildWAggEventSortLink(c, currentSort, currentOrder, "finished_time", "FinishedTime"),
		Duration:     buildWAggEventSortLink(c, currentSort, currentOrder, "duration_ms", "Duration(秒)"),
	}
}

func buildWAggEventSortLink(c *gin.Context, currentSort string, currentOrder string, sortField string, label string) wAggEventSortLink {
	nextOrder := "desc"
	active := currentSort == sortField
	if active && currentOrder == "desc" {
		nextOrder = "asc"
	}

	q := c.Request.URL.Query()
	q.Set("p", "1")
	q.Set("sort", sortField)
	q.Set("order", nextOrder)
	href := c.Request.URL.Path
	if encoded := q.Encode(); encoded != "" {
		href += "?" + encoded
	}

	return wAggEventSortLink{
		Label:  label,
		Href:   href,
		Active: active,
		Desc:   active && currentOrder == "desc",
	}
}
