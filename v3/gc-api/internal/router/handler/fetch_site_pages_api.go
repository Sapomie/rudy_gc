package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/service/sehuatang"
	"rudy_gc/internal/service/wkv"
)

type fetchSitePageQuery struct {
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

type fetchSiteSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type fetchSiteSortQuery struct {
	ByMovieName      fetchSiteSortLink
	ByReleaseDate    fetchSiteSortLink
	ByFetchStatus    fetchSiteSortLink
	ByLastFetchTime  fetchSiteSortLink
	ByResultCount    fetchSiteSortLink
	ByHashCount      fetchSiteSortLink
	ByLatestPublish  fetchSiteSortLink
	ByMediaBirthTime fetchSiteSortLink
}

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

func (h *CrawlerPagesAPI) FetchSiteJavbusListPageMain(c *gin.Context) {
	defer withFrontendPath(c, "/fetch-site-javbus-list")()

	q, pageQuery, statusSelected, ok := h.parseFetchSitePageQuery(c, "JavBus")
	if !ok {
		return
	}

	result, err := h.fetchSite.BuildJavbusPage(c.Request.Context(), fetchsite.JavbusPageQuery(pageQuery))
	if err != nil {
		jsonServerError(c, "JavBus 抓取列表加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":                  "JavBus 抓取列表",
		"ViewType":               "fetch-site-javbus-list",
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
		"PageInfo":               buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize),
		"SortQuery":              buildFetchSiteSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":            "/fetch-site-sukebei-list",
		"TaskPageURL":            "/triggers/fetch-site-javbus-filtered",
		"TasksPageURL":           "/crawler/tasks",
		"FilteredJavbusTaskType": loop.TaskSpiderFetchJavbusFilter,
		"StatusSelected":         statusSelected,
		"StatusAllActive":        len(statusSelected) == 0,
	})
}

func (h *CrawlerPagesAPI) FetchSiteSukebeiListPageMain(c *gin.Context) {
	defer withFrontendPath(c, "/fetch-site-sukebei-list")()

	q, pageQuery, statusSelected, ok := h.parseFetchSitePageQuery(c, "Sukebei")
	if !ok {
		return
	}

	result, err := h.fetchSite.BuildSukebeiPage(c.Request.Context(), fetchsite.SukebeiPageQuery(pageQuery))
	if err != nil {
		jsonServerError(c, "Sukebei 抓取列表加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":                   "Sukebei 抓取列表",
		"ViewType":                "fetch-site-sukebei-list",
		"PageTitle":               "Sukebei 抓取列表",
		"PageNote":                "只展示 t_sukebei_torrent_fetch，并支持独立筛选、排序、分页和行级触发。",
		"Query":                   q,
		"Rows":                    result.Items,
		"Total":                   result.Total,
		"total":                   result.Total,
		"SuccessCount":            result.SuccessCount,
		"PendingCount":            result.PendingCount,
		"FailedCount":             result.FailedCount,
		"ownedQuery":              buildOwnedFilterInfo(c),
		"PageInfo":                buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize),
		"SortQuery":               buildFetchSiteSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":             "/fetch-site-javbus-list",
		"TaskPageURL":             "/triggers/fetch-site-sukebei-filtered",
		"TasksPageURL":            "/crawler/tasks",
		"FilteredSukebeiTaskType": loop.TaskSpiderFetchSukebeiFilter,
		"StatusSelected":          statusSelected,
		"StatusAllActive":         len(statusSelected) == 0,
	})
}

func (h *CrawlerPagesAPI) parseFetchSitePageQuery(c *gin.Context, label string) (fetchSitePageQuery, fetchsite.SukebeiPageQuery, map[string]bool, bool) {
	var q fetchSitePageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return q, fetchsite.SukebeiPageQuery{}, nil, false
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	q.Sort = normalizeFetchSiteSortField(q.Sort)
	q.Order = normalizeFetchSiteSortOrder(q.Order)

	pageQuery := fetchsite.SukebeiPageQuery{
		Page:     q.Page,
		PageSize: q.PageSize,
		Sort:     q.Sort,
		Order:    q.Order,
		Keyword:  strings.TrimSpace(q.Keyword),
	}
	if v, ok, err := parseOwnedFilterValue(q.MediaOwned); err != nil {
		jsonBadRequest(c, "WMedia 库存筛选错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.MediaOwned = v
	}
	if values, ok, err := parseFetchSiteStatuses(q.Statuses, q.Status); err != nil {
		jsonBadRequest(c, label+" 状态错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.Statuses = values
		pageQuery.HasStatuses = true
	}
	if ts, ok, err := parseOptionalDateStart(q.LastFetchFrom); err != nil {
		jsonBadRequest(c, label+" 最后抓取开始日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.LastFetchFrom = ts
		pageQuery.HasLastFetchFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.LastFetchTo); err != nil {
		jsonBadRequest(c, label+" 最后抓取结束日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.LastFetchTo = ts
		pageQuery.HasLastFetchTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.ReleaseDateFrom); err != nil {
		jsonBadRequest(c, label+" 发行时间开始日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.ReleaseDateFrom = ts
		pageQuery.HasReleaseDateFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.ReleaseDateTo); err != nil {
		jsonBadRequest(c, label+" 发行时间结束日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.ReleaseDateTo = ts
		pageQuery.HasReleaseDateTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.MediaBirthFrom); err != nil {
		jsonBadRequest(c, label+" WMedia 下载时间开始日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.MediaBirthFrom = ts
		pageQuery.HasMediaBirthFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.MediaBirthTo); err != nil {
		jsonBadRequest(c, label+" WMedia 下载时间结束日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.MediaBirthTo = ts
		pageQuery.HasMediaBirthTo = true
	}

	return q, pageQuery, buildFetchSiteStatusSelected(q.Statuses, q.Status), true
}

func (h *CrawlerPagesAPI) FetchSiteSehuatangListPageMain(c *gin.Context) {
	defer withFrontendPath(c, "/fetch-site-sehuatang-list")()

	var q sehuatangMagnetPageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
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
		jsonBadRequest(c, "发帖时间开始日期错误", err)
		return
	} else if ok {
		request.PostTimeFrom = ts
		request.HasPostTimeFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.PostTimeTo); err != nil {
		jsonBadRequest(c, "发帖时间结束日期错误", err)
		return
	} else if ok {
		request.PostTimeTo = ts
		request.HasPostTimeTo = true
	}

	result, err := h.sehuatang.ListPage(c.Request.Context(), request)
	if err != nil {
		jsonServerError(c, "Sehuatang 列表加载失败", err)
		return
	}
	shtTimeValue, err := h.wkvSvc.GetValue(c.Request.Context(), wkv.ItemKeySHTTime)
	if err != nil {
		shtTimeValue = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":          "Sehuatang Magnet 列表",
		"ViewType":       "fetch-site-sehuatang-list",
		"PageTitle":      "Sehuatang Magnet 列表",
		"PageNote":       "展示 t_sehuatang_magnet 已入库的数据，支持关键词、movie_jav_id、tag、info_hash 筛选以及分页排序。",
		"Query":          q,
		"Rows":           result.Items,
		"Albums":         result.Albums,
		"Total":          result.Total,
		"PageInfo":       buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize),
		"SortQuery":      buildSehuatangMagnetSortQuery(c, q.Sort, q.Order),
		"QuickTags":      buildSehuatangTagQuickFilters(c, q.Tag),
		"TaskPageURL":    "/triggers/fetch-sehuatang",
		"TasksPageURL":   "/crawler/tasks",
		"ShtTimeValue":   shtTimeValue,
		"ShtTimeDisplay": displayWKvDateValue(shtTimeValue),
	})
}

func buildFetchSitePageInfo(c *gin.Context, total int64, page int64, pageSize int64) *PageInfo {
	return BuildPageInfo(c, total, page, pageSize, 2)
}

func buildFetchSiteSortQuery(c *gin.Context, currentField string, currentOrder string) *fetchSiteSortQuery {
	makeHref := func(field string) fetchSiteSortLink {
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
		return fetchSiteSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &fetchSiteSortQuery{
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

func normalizeFetchSiteSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "release_date", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time", "media_birth_time":
		return strings.TrimSpace(raw)
	default:
		return "last_fetch_time"
	}
}

func normalizeFetchSiteSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func parseOwnedFilterValue(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	if v < 1 || v > 7 {
		return 0, false, fmt.Errorf("must be between 1 and 7")
	}
	return v, true, nil
}

func parseFetchSiteStatuses(raws []string, legacy string) ([]int64, bool, error) {
	merged := make([]string, 0, len(raws)+1)
	for _, raw := range raws {
		current := strings.TrimSpace(raw)
		if current == "" {
			continue
		}
		merged = append(merged, current)
	}
	if len(merged) == 0 {
		current := strings.TrimSpace(legacy)
		if current != "" {
			merged = append(merged, current)
		}
	}
	if len(merged) == 0 {
		return nil, false, nil
	}

	values := make([]int64, 0, len(merged))
	seen := make(map[int64]struct{}, len(merged))
	for _, raw := range merged {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, false, err
		}
		if v < 1 || v > 4 {
			return nil, false, fmt.Errorf("must be between 1 and 4")
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		values = append(values, v)
	}
	return values, len(values) > 0, nil
}

func buildFetchSiteStatusSelected(raws []string, legacy string) map[string]bool {
	selected := make(map[string]bool)
	for _, raw := range raws {
		current := strings.TrimSpace(raw)
		if current == "" {
			continue
		}
		selected[current] = true
	}
	if len(selected) == 0 {
		current := strings.TrimSpace(legacy)
		if current != "" {
			selected[current] = true
		}
	}
	return selected
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
