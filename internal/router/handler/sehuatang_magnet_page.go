package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/sehuatang"
	"rudy_gc/internal/service/wkv"
)

type sehuatangMagnetPageQuery struct {
	Page         int64  `form:"p"`
	PageSize     int64  `form:"ps"`
	Sort         string `form:"sort"`
	Order        string `form:"order"`
	Keyword      string `form:"keyword"`
	MovieJavID   string `form:"movie_jav_id"`
	InfoHash     string `form:"info_hash"`
	Tag          string `form:"tag"`
	PostTimeFrom string `form:"post_time_from"`
	PostTimeTo   string `form:"post_time_to"`
}

type sehuatangMagnetSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type sehuatangTagQuickFilter struct {
	Label  string
	Href   string
	Active bool
}

type sehuatangMagnetSortQuery struct {
	ByMovieJavID   sehuatangMagnetSortLink
	ByMovieName    sehuatangMagnetSortLink
	ByPostTime     sehuatangMagnetSortLink
	ByPostDate     sehuatangMagnetSortLink
	ByLastSeenTime sehuatangMagnetSortLink
	ByCreatedOn    sehuatangMagnetSortLink
	ByUpdatedOn    sehuatangMagnetSortLink
}

func (h *CrawlerPages) FetchSiteSehuatangListPageMain(c *gin.Context) {
	var q sehuatangMagnetPageQuery
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
	q.Sort = normalizeSehuatangMagnetSortField(q.Sort)
	q.Order = normalizeSehuatangMagnetSortOrder(q.Order)

	request := sehuatang.ListRequest{
		Page:       q.Page,
		PageSize:   q.PageSize,
		Sort:       q.Sort,
		Order:      q.Order,
		Keyword:    strings.TrimSpace(q.Keyword),
		MovieJavID: strings.TrimSpace(q.MovieJavID),
		InfoHash:   strings.TrimSpace(q.InfoHash),
		Tag:        strings.TrimSpace(q.Tag),
	}
	if ts, ok, err := parseOptionalDateStart(q.PostTimeFrom); err != nil {
		c.String(http.StatusBadRequest, "发帖时间开始日期错误: %v", err)
		return
	} else if ok {
		request.PostTimeFrom = ts
		request.HasPostTimeFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.PostTimeTo); err != nil {
		c.String(http.StatusBadRequest, "发帖时间结束日期错误: %v", err)
		return
	} else if ok {
		request.PostTimeTo = ts
		request.HasPostTimeTo = true
	}

	result, err := h.sehuatang.ListPage(c.Request.Context(), request)
	if err != nil {
		c.String(http.StatusInternalServerError, "Sehuatang 列表加载失败: %v", err)
		return
	}
	shtTimeValue, err := h.wkvSvc.GetValue(c.Request.Context(), wkv.ItemKeySHTTime)
	if err != nil {
		shtTimeValue = ""
	}

	pageInfo := buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize)
	c.HTML(http.StatusOK, "page.sehuatang_magnet_list", gin.H{
		"Title":          "Sehuatang Magnet 列表",
		"PageTitle":      "Sehuatang Magnet 列表",
		"PageNote":       "展示 t_sehuatang_magnet 已入库的数据，支持关键词、movie_jav_id、tag、info_hash 筛选以及分页排序。",
		"Query":          q,
		"Rows":           result.Items,
		"Albums":         result.Albums,
		"Total":          result.Total,
		"PageInfo":       pageInfo,
		"SortQuery":      buildSehuatangMagnetSortQuery(c, q.Sort, q.Order),
		"QuickTags":      buildSehuatangTagQuickFilters(c, q.Tag),
		"TaskPageURL":    "/triggers/fetch-sehuatang",
		"TasksPageURL":   "/crawler/tasks",
		"ShtTimeValue":   shtTimeValue,
		"ShtTimeDisplay": displayWKvDateValue(shtTimeValue),
	})
}

func buildSehuatangMagnetSortQuery(c *gin.Context, currentField string, currentOrder string) *sehuatangMagnetSortQuery {
	makeHref := func(field string) sehuatangMagnetSortLink {
		q := cloneValues(c)
		q.Set("p", "1")
		q.Set("sort", field)
		if currentField == field && currentOrder == "desc" {
			q.Set("order", "asc")
		} else {
			q.Set("order", "desc")
		}
		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		return sehuatangMagnetSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &sehuatangMagnetSortQuery{
		ByMovieJavID:   makeHref("movie_jav_id"),
		ByMovieName:    makeHref("movie_name"),
		ByPostTime:     makeHref("post_time"),
		ByPostDate:     makeHref("post_date"),
		ByLastSeenTime: makeHref("last_seen_time"),
		ByCreatedOn:    makeHref("created_on"),
		ByUpdatedOn:    makeHref("updated_on"),
	}
}

func normalizeSehuatangMagnetSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_jav_id", "movie_name", "post_time", "post_date", "last_seen_time", "created_on", "updated_on":
		return strings.TrimSpace(raw)
	default:
		return "post_time"
	}
}

func normalizeSehuatangMagnetSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func buildSehuatangTagQuickFilters(c *gin.Context, currentTag string) []*sehuatangTagQuickFilter {
	current := strings.TrimSpace(currentTag)
	return []*sehuatangTagQuickFilter{
		{
			Label:  "全部",
			Href:   buildSehuatangTagFilterHref(c, ""),
			Active: current == "",
		},
		{
			Label:  "自提征用",
			Href:   buildSehuatangTagFilterHref(c, "自提征用"),
			Active: current == "自提征用",
		},
		{
			Label:  "FC2PPV",
			Href:   buildSehuatangTagFilterHref(c, "FC2PPV"),
			Active: strings.EqualFold(current, "FC2PPV"),
		},
	}
}

func buildSehuatangTagFilterHref(c *gin.Context, tag string) string {
	q := cloneValues(c)
	q.Set("p", "1")
	if strings.TrimSpace(tag) == "" {
		q.Del("tag")
	} else {
		q.Set("tag", strings.TrimSpace(tag))
	}
	href := c.Request.URL.Path
	if enc := q.Encode(); enc != "" {
		href += "?" + enc
	}
	return href
}
