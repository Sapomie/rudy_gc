package list

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/router/handler/common"
)

type crawlRecordListQuery struct {
	Type     string `form:"type"`
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

const crawlRecordTypeAll = "all"

type crawlRecordListRow struct {
	Name            string
	Type            string
	DetailNumber    int64
	DurationMinText string
	StartedTime     int64
}

type crawlRecordTypeLink struct {
	Label  string
	Href   string
	Active bool
}

type crawlRecordSortLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

type crawlRecordHeaderLinks struct {
	StartedTime  crawlRecordSortLink
	Type         crawlRecordSortLink
	DetailNumber crawlRecordSortLink
	DurationMin  crawlRecordSortLink
	Name         crawlRecordSortLink
}

func (h *ListPagesAPI) CrawlRecordList(c *gin.Context) {
	defer common.WithFrontendPath(c, "/crawlrecordlist")()

	var q crawlRecordListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	if q.Type == "" {
		q.Type = consts.RecordTypeSeedsActive
	}
	q.Sort = normalizeCrawlRecordSort(q.Sort)
	q.Order = normalizeCrawlRecordOrder(q.Order)

	rows, total, err := h.deps.RecordModel.ListPage(c.Request.Context(), moviex.ERecordListFilter{
		Type:     crawlRecordFilterType(q.Type),
		Sort:     q.Sort,
		Order:    q.Order,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		common.JSONServerError(c, "query crawl records failed", err)
		return
	}

	pageInfo := common.BuildPageInfo(c, total, q.Page, q.PageSize, common.PageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":       "CrawlRecords",
		"ViewType":    "crawl-records",
		"PageTitle":   "CrawlRecords 列表",
		"PageNote":    "直接查看抓取流程写入的 e_record 记录。",
		"Rows":        buildCrawlRecordListRows(rows),
		"Total":       total,
		"PageInfo":    pageInfo,
		"pageInfo":    pageInfo,
		"Query":       q,
		"TypeLinks":   buildCrawlRecordTypeLinks(c, q.Type),
		"HeaderLinks": buildCrawlRecordHeaderLinks(c, q.Sort, q.Order),
		"ClearHref":   "/crawlrecordlist",
	})
}

func buildCrawlRecordListRows(rows []*moviex.ERecord) []*crawlRecordListRow {
	if len(rows) == 0 {
		return []*crawlRecordListRow{}
	}

	out := make([]*crawlRecordListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &crawlRecordListRow{
			Name:            row.Name,
			Type:            row.Type,
			DetailNumber:    row.DetailNumber,
			DurationMinText: formatCrawlRecordDurationMinutes(row.StartTime, row.EndTime),
			StartedTime:     row.StartTime,
		})
	}
	return out
}

func formatCrawlRecordDurationMinutes(startTime int64, endTime int64) string {
	duration := endTime - startTime
	if duration <= 0 {
		return "-"
	}
	minutes := (duration + 59) / 60
	return fmt.Sprintf("%d", minutes)
}

func normalizeCrawlRecordSort(sortField string) string {
	switch sortField {
	case "type":
		return "type"
	case "detail_number":
		return "detail_number"
	case "duration":
		return "duration"
	case "name":
		return "name"
	default:
		return "start_time"
	}
}

func normalizeCrawlRecordOrder(sortOrder string) string {
	if sortOrder == "asc" {
		return "asc"
	}
	return "desc"
}

func crawlRecordFilterType(recordType string) string {
	if recordType == crawlRecordTypeAll {
		return ""
	}
	return recordType
}

func buildCrawlRecordTypeLinks(c *gin.Context, currentType string) []crawlRecordTypeLink {
	options := []struct {
		Label string
		Value string
	}{
		{Label: "All", Value: crawlRecordTypeAll},
		{Label: consts.RecordTypeDailyBest, Value: consts.RecordTypeDailyBest},
		{Label: consts.RecordTypeSeedsActive, Value: consts.RecordTypeSeedsActive},
		{Label: consts.RecordTypeSeedName, Value: consts.RecordTypeSeedName},
	}

	out := make([]crawlRecordTypeLink, 0, len(options))
	for _, option := range options {
		q := c.Request.URL.Query()
		q.Set("p", "1")
		q.Set("type", option.Value)

		href := c.Request.URL.Path
		if encoded := q.Encode(); encoded != "" {
			href += "?" + encoded
		}
		out = append(out, crawlRecordTypeLink{
			Label:  option.Label,
			Href:   href,
			Active: currentType == option.Value,
		})
	}
	return out
}

func buildCrawlRecordHeaderLinks(c *gin.Context, currentSort string, currentOrder string) crawlRecordHeaderLinks {
	return crawlRecordHeaderLinks{
		StartedTime:  buildCrawlRecordSortLink(c, currentSort, currentOrder, "start_time", "StartedTime"),
		Type:         buildCrawlRecordSortLink(c, currentSort, currentOrder, "type", "Type"),
		DetailNumber: buildCrawlRecordSortLink(c, currentSort, currentOrder, "detail_number", "DetailNumber"),
		DurationMin:  buildCrawlRecordSortLink(c, currentSort, currentOrder, "duration", "Duration(min)"),
		Name:         buildCrawlRecordSortLink(c, currentSort, currentOrder, "name", "Name"),
	}
}

func buildCrawlRecordSortLink(c *gin.Context, currentSort string, currentOrder string, sortField string, label string) crawlRecordSortLink {
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

	return crawlRecordSortLink{
		Label:  label,
		Href:   href,
		Active: active,
		Desc:   active && currentOrder == "desc",
	}
}
