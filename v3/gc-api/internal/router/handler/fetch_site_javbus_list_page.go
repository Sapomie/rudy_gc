package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/loop"
)

type fetchSiteJavbusPageQuery struct {
	Page            int64    `form:"p"`
	PageSize        int64    `form:"ps"`
	Sort            string   `form:"sort"`
	Order           string   `form:"order"`
	MediaOwned      string   `form:"mowned"`
	Keyword         string   `form:"keyword"`
	Status          string   `form:"status"`
	Statuses        []string `form:"statuses"`
	TriggerSortKey  string   `form:"trigger_sort_key"`
	LastFetchFrom   string   `form:"last_fetch_from"`
	LastFetchTo     string   `form:"last_fetch_to"`
	ReleaseDateFrom string   `form:"release_date_from"`
	ReleaseDateTo   string   `form:"release_date_to"`
	MediaBirthFrom  string   `form:"media_birth_from"`
	MediaBirthTo    string   `form:"media_birth_to"`
}

type fetchSiteJavbusSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type fetchSiteJavbusSortQuery struct {
	ByMovieName      fetchSiteJavbusSortLink
	ByReleaseDate    fetchSiteJavbusSortLink
	ByFetchStatus    fetchSiteJavbusSortLink
	ByLastFetchTime  fetchSiteJavbusSortLink
	ByResultCount    fetchSiteJavbusSortLink
	ByHashCount      fetchSiteJavbusSortLink
	ByLatestPublish  fetchSiteJavbusSortLink
	ByMediaBirthTime fetchSiteJavbusSortLink
}

func (h *CrawlerPages) FetchSiteJavbusListPageMain(c *gin.Context) {
	var q fetchSiteJavbusPageQuery
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
	q.Sort = normalizeFetchSiteJavbusSortField(q.Sort)
	q.Order = normalizeFetchSiteJavbusSortOrder(q.Order)

	pageQuery := fetchsite.JavbusPageQuery{
		Page:     q.Page,
		PageSize: q.PageSize,
		Sort:     q.Sort,
		Order:    q.Order,
		Keyword:  strings.TrimSpace(q.Keyword),
	}
	if v, ok, err := parseOwnedFilterValue(q.MediaOwned); err != nil {
		c.String(http.StatusBadRequest, "WMedia 库存筛选错误: %v", err)
		return
	} else if ok {
		pageQuery.MediaOwned = v
	}
	if values, ok, err := parseFetchSiteStatuses(q.Statuses, q.Status); err != nil {
		c.String(http.StatusBadRequest, "JavBus 状态错误: %v", err)
		return
	} else if ok {
		pageQuery.Statuses = values
		pageQuery.HasStatuses = true
	}
	if ts, ok, err := parseOptionalDateStart(q.LastFetchFrom); err != nil {
		c.String(http.StatusBadRequest, "JavBus 最后抓取开始日期错误: %v", err)
		return
	} else if ok {
		pageQuery.LastFetchFrom = ts
		pageQuery.HasLastFetchFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.LastFetchTo); err != nil {
		c.String(http.StatusBadRequest, "JavBus 最后抓取结束日期错误: %v", err)
		return
	} else if ok {
		pageQuery.LastFetchTo = ts
		pageQuery.HasLastFetchTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.ReleaseDateFrom); err != nil {
		c.String(http.StatusBadRequest, "JavBus 发行时间开始日期错误: %v", err)
		return
	} else if ok {
		pageQuery.ReleaseDateFrom = ts
		pageQuery.HasReleaseDateFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.ReleaseDateTo); err != nil {
		c.String(http.StatusBadRequest, "JavBus 发行时间结束日期错误: %v", err)
		return
	} else if ok {
		pageQuery.ReleaseDateTo = ts
		pageQuery.HasReleaseDateTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.MediaBirthFrom); err != nil {
		c.String(http.StatusBadRequest, "JavBus WMedia 下载时间开始日期错误: %v", err)
		return
	} else if ok {
		pageQuery.MediaBirthFrom = ts
		pageQuery.HasMediaBirthFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.MediaBirthTo); err != nil {
		c.String(http.StatusBadRequest, "JavBus WMedia 下载时间结束日期错误: %v", err)
		return
	} else if ok {
		pageQuery.MediaBirthTo = ts
		pageQuery.HasMediaBirthTo = true
	}

	result, err := h.fetchSite.BuildJavbusPage(c.Request.Context(), pageQuery)
	if err != nil {
		c.String(http.StatusInternalServerError, "JavBus 抓取列表加载失败: %v", err)
		return
	}

	pageInfo := buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize)
	statusSelected := buildFetchSiteSukebeiStatusSelected(q.Statuses, q.Status)
	c.HTML(http.StatusOK, "page.fetch_site_javbus_list", gin.H{
		"Title":                  "JavBus 抓取列表",
		"PageTitle":              "JavBus 抓取列表",
		"PageNote":               "只展示 t_javbus_magnet_fetch，并支持独立筛选、排序、分页和行级触发。",
		"Query":                  q,
		"Rows":                   result.Items,
		"Total":                  result.Total,
		"total":                  result.Total,
		"SuccessCount":           result.SuccessCount,
		"PendingCount":           result.PendingCount,
		"FailedCount":            result.FailedCount,
		"ownedQuery":             buildOwnedFilterInfo(c),
		"PageInfo":               pageInfo,
		"SortQuery":              buildFetchSiteJavbusSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":            "/fetch-site-sukebei-list",
		"TaskPageURL":            "/triggers/fetch-site-javbus-filtered",
		"TasksPageURL":           "/crawler/tasks",
		"FilteredJavbusTaskType": loop.TaskSpiderFetchJavbusFilter,
		"StatusSelected":         statusSelected,
		"StatusAllActive":        len(statusSelected) == 0,
	})
}

func buildFetchSiteJavbusSortQuery(c *gin.Context, currentField string, currentOrder string) *fetchSiteJavbusSortQuery {
	makeHref := func(field string) fetchSiteJavbusSortLink {
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
		return fetchSiteJavbusSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &fetchSiteJavbusSortQuery{
		ByMovieName:      makeHref("movie_name"),
		ByReleaseDate:    makeHref("release_date"),
		ByFetchStatus:    makeHref("fetch_status"),
		ByLastFetchTime:  makeHref("last_fetch_time"),
		ByResultCount:    makeHref("last_result_count"),
		ByHashCount:      makeHref("torrent_hash_count"),
		ByLatestPublish:  makeHref("latest_publish_time"),
		ByMediaBirthTime: makeHref("media_birth_time"),
	}
}

func normalizeFetchSiteJavbusSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "release_date", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time", "media_birth_time":
		return strings.TrimSpace(raw)
	default:
		return "last_fetch_time"
	}
}

func normalizeFetchSiteJavbusSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}
