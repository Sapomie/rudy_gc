package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/fetchsite"
)

type fetchSiteSukebeiPageQuery struct {
	Page            int64  `form:"p"`
	PageSize        int64  `form:"ps"`
	Sort            string `form:"sort"`
	Order           string `form:"order"`
	Keyword         string `form:"keyword"`
	Status          string `form:"status"`
	HasError        string `form:"has_error"`
	LastFetchFrom   string `form:"last_fetch_from"`
	LastFetchTo     string `form:"last_fetch_to"`
	ReleaseDateFrom string `form:"release_date_from"`
	ReleaseDateTo   string `form:"release_date_to"`
}

type fetchSiteSukebeiSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type fetchSiteSukebeiSortQuery struct {
	ByMovieCode     fetchSiteSukebeiSortLink
	ByFetchStatus   fetchSiteSukebeiSortLink
	ByLastFetchTime fetchSiteSukebeiSortLink
	ByResultCount   fetchSiteSukebeiSortLink
	ByHashCount     fetchSiteSukebeiSortLink
	ByLatestPublish fetchSiteSukebeiSortLink
}

func (h *CrawlerPages) FetchSiteSukebeiListPageMain(c *gin.Context) {
	var q fetchSiteSukebeiPageQuery
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
	q.Sort = normalizeFetchSiteSukebeiSortField(q.Sort)
	q.Order = normalizeFetchSiteSukebeiSortOrder(q.Order)

	pageQuery := fetchsite.SukebeiPageQuery{
		Page:     q.Page,
		PageSize: q.PageSize,
		Sort:     q.Sort,
		Order:    q.Order,
		Keyword:  strings.TrimSpace(q.Keyword),
	}
	if v, ok, err := parseFetchSiteStatus(q.Status); err != nil {
		c.String(http.StatusBadRequest, "Sukebei 状态错误: %v", err)
		return
	} else if ok {
		pageQuery.Status = v
		pageQuery.StatusSet = true
	}
	switch strings.TrimSpace(q.HasError) {
	case "":
	case "error":
		pageQuery.HasErrorOnly = true
	case "clean":
		pageQuery.HasNoErrorOnly = true
	default:
		c.String(http.StatusBadRequest, "Sukebei 错误筛选参数错误")
		return
	}
	if ts, ok, err := parseOptionalDateStart(q.LastFetchFrom); err != nil {
		c.String(http.StatusBadRequest, "Sukebei 最后抓取开始日期错误: %v", err)
		return
	} else if ok {
		pageQuery.LastFetchFrom = ts
		pageQuery.HasLastFetchFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.LastFetchTo); err != nil {
		c.String(http.StatusBadRequest, "Sukebei 最后抓取结束日期错误: %v", err)
		return
	} else if ok {
		pageQuery.LastFetchTo = ts
		pageQuery.HasLastFetchTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.ReleaseDateFrom); err != nil {
		c.String(http.StatusBadRequest, "Sukebei 发行时间开始日期错误: %v", err)
		return
	} else if ok {
		pageQuery.ReleaseDateFrom = ts
		pageQuery.HasReleaseDateFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.ReleaseDateTo); err != nil {
		c.String(http.StatusBadRequest, "Sukebei 发行时间结束日期错误: %v", err)
		return
	} else if ok {
		pageQuery.ReleaseDateTo = ts
		pageQuery.HasReleaseDateTo = true
	}

	result, err := h.fetchSite.BuildSukebeiPage(c.Request.Context(), pageQuery)
	if err != nil {
		c.String(http.StatusInternalServerError, "Sukebei 抓取列表加载失败: %v", err)
		return
	}

	pageInfo := buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize)
	c.HTML(http.StatusOK, "page.fetch_site_sukebei_list", gin.H{
		"Title":        "Sukebei 抓取列表",
		"PageTitle":    "Sukebei 抓取列表",
		"PageNote":     "只展示 t_sukebei_torrent_fetch，并支持独立筛选、排序、分页和行级触发。",
		"Query":        q,
		"Rows":         result.Items,
		"Total":        result.Total,
		"PageInfo":     pageInfo,
		"SortQuery":    buildFetchSiteSukebeiSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":  "/fetch-site-javbus-list",
		"TaskPageURL":  "/triggers/fetch-site",
		"TasksPageURL": "/crawler/tasks",
	})
}

func buildFetchSiteSukebeiSortQuery(c *gin.Context, currentField string, currentOrder string) *fetchSiteSukebeiSortQuery {
	makeHref := func(field string) fetchSiteSukebeiSortLink {
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
		return fetchSiteSukebeiSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &fetchSiteSukebeiSortQuery{
		ByMovieCode:     makeHref("movie_code"),
		ByFetchStatus:   makeHref("fetch_status"),
		ByLastFetchTime: makeHref("last_fetch_time"),
		ByResultCount:   makeHref("last_result_count"),
		ByHashCount:     makeHref("torrent_hash_count"),
		ByLatestPublish: makeHref("latest_publish_time"),
	}
}

func normalizeFetchSiteSukebeiSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_code", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time":
		return strings.TrimSpace(raw)
	default:
		return "last_fetch_time"
	}
}

func normalizeFetchSiteSukebeiSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}
