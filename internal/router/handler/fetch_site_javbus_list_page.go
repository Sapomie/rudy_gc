package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/fetchsite"
)

type fetchSiteJavbusPageQuery struct {
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

type fetchSiteJavbusSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type fetchSiteJavbusSortQuery struct {
	ByMovieCode     fetchSiteJavbusSortLink
	ByFetchStatus   fetchSiteJavbusSortLink
	ByLastFetchTime fetchSiteJavbusSortLink
	ByResultCount   fetchSiteJavbusSortLink
	ByHashCount     fetchSiteJavbusSortLink
	ByLatestPublish fetchSiteJavbusSortLink
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
	if v, ok, err := parseFetchSiteStatus(q.Status); err != nil {
		c.String(http.StatusBadRequest, "JavBus 状态错误: %v", err)
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
		c.String(http.StatusBadRequest, "JavBus 错误筛选参数错误")
		return
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

	result, err := h.fetchSite.BuildJavbusPage(c.Request.Context(), pageQuery)
	if err != nil {
		c.String(http.StatusInternalServerError, "JavBus 抓取列表加载失败: %v", err)
		return
	}

	pageInfo := buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize)
	c.HTML(http.StatusOK, "page.fetch_site_javbus_list", gin.H{
		"Title":        "JavBus 抓取列表",
		"PageTitle":    "JavBus 抓取列表",
		"PageNote":     "只展示 t_javbus_magnet_fetch，并支持独立筛选、排序、分页和行级触发。",
		"Query":        q,
		"Rows":         result.Items,
		"Total":        result.Total,
		"PageInfo":     pageInfo,
		"SortQuery":    buildFetchSiteJavbusSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":  "/fetch-site-sukebei-list",
		"TaskPageURL":  "/triggers/fetch-site",
		"TasksPageURL": "/crawler/tasks",
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
		ByMovieCode:     makeHref("movie_code"),
		ByFetchStatus:   makeHref("fetch_status"),
		ByLastFetchTime: makeHref("last_fetch_time"),
		ByResultCount:   makeHref("last_result_count"),
		ByHashCount:     makeHref("torrent_hash_count"),
		ByLatestPublish: makeHref("latest_publish_time"),
	}
}

func normalizeFetchSiteJavbusSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_code", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time":
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

func buildFetchSitePageInfo(c *gin.Context, total int64, page int64, pageSize int64) *PageInfo {
	if pageSize <= 0 {
		pageSize = 50
	}
	pageTotal := (total + pageSize - 1) / pageSize
	if pageTotal <= 0 {
		pageTotal = 1
	}
	if page < 1 {
		page = 1
	}
	if page > pageTotal {
		page = pageTotal
	}

	makeHref := func(targetPage int64) string {
		q := cloneValues(c)
		q.Set("p", strconv.FormatInt(targetPage, 10))
		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		return href
	}

	prev := page - 1
	if prev < 1 {
		prev = 1
	}
	next := page + 1
	if next > pageTotal {
		next = pageTotal
	}
	links := []PageLink{
		{Label: "Start", Page: 1, Href: makeHref(1), Disabled: page == 1},
		{Label: "Prev", Page: prev, Href: makeHref(prev), Disabled: page == 1},
	}
	for p := page - 2; p <= page+2; p++ {
		if p < 1 || p > pageTotal {
			continue
		}
		links = append(links, PageLink{
			Label:  strconv.FormatInt(p, 10),
			Page:   p,
			Href:   makeHref(p),
			Active: p == page,
		})
	}
	links = append(links,
		PageLink{Label: "Next", Page: next, Href: makeHref(next), Disabled: page == pageTotal},
		PageLink{Label: "End", Page: pageTotal, Href: makeHref(pageTotal), Disabled: page == pageTotal},
	)

	return &PageInfo{
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		PageTotal: pageTotal,
		StartHref: makeHref(1),
		PrevHref:  makeHref(prev),
		NextHref:  makeHref(next),
		EndHref:   makeHref(pageTotal),
		Links:     links,
	}
}

func parseFetchSiteStatus(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return v, true, nil
}
