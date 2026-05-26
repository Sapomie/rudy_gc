package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/loop"
)

type fetchSiteSukebeiPageQuery struct {
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

type fetchSiteSukebeiSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type fetchSiteSukebeiSortQuery struct {
	ByMovieName      fetchSiteSukebeiSortLink
	ByReleaseDate    fetchSiteSukebeiSortLink
	ByFetchStatus    fetchSiteSukebeiSortLink
	ByLastFetchTime  fetchSiteSukebeiSortLink
	ByResultCount    fetchSiteSukebeiSortLink
	ByHashCount      fetchSiteSukebeiSortLink
	ByLatestPublish  fetchSiteSukebeiSortLink
	ByMediaBirthTime fetchSiteSukebeiSortLink
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
	if v, ok, err := parseOwnedFilterValue(q.MediaOwned); err != nil {
		c.String(http.StatusBadRequest, "WMedia 库存筛选错误: %v", err)
		return
	} else if ok {
		pageQuery.MediaOwned = v
	}
	if values, ok, err := parseFetchSiteStatuses(q.Statuses, q.Status); err != nil {
		c.String(http.StatusBadRequest, "Sukebei 状态错误: %v", err)
		return
	} else if ok {
		pageQuery.Statuses = values
		pageQuery.HasStatuses = true
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
	if ts, ok, err := parseOptionalDateStart(q.MediaBirthFrom); err != nil {
		c.String(http.StatusBadRequest, "Sukebei WMedia 下载时间开始日期错误: %v", err)
		return
	} else if ok {
		pageQuery.MediaBirthFrom = ts
		pageQuery.HasMediaBirthFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.MediaBirthTo); err != nil {
		c.String(http.StatusBadRequest, "Sukebei WMedia 下载时间结束日期错误: %v", err)
		return
	} else if ok {
		pageQuery.MediaBirthTo = ts
		pageQuery.HasMediaBirthTo = true
	}

	result, err := h.fetchSite.BuildSukebeiPage(c.Request.Context(), pageQuery)
	if err != nil {
		c.String(http.StatusInternalServerError, "Sukebei 抓取列表加载失败: %v", err)
		return
	}

	pageInfo := buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize)
	statusSelected := buildFetchSiteSukebeiStatusSelected(q.Statuses, q.Status)
	c.HTML(http.StatusOK, "page.fetch_site_sukebei_list", gin.H{
		"Title":                   "Sukebei 抓取列表",
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
		"PageInfo":                pageInfo,
		"SortQuery":               buildFetchSiteSukebeiSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":             "/fetch-site-javbus-list",
		"TaskPageURL":             "/triggers/fetch-site-sukebei-filtered",
		"TasksPageURL":            "/crawler/tasks",
		"FilteredSukebeiTaskType": loop.TaskSpiderFetchSukebeiFilter,
		"StatusSelected":          statusSelected,
		"StatusAllActive":         len(statusSelected) == 0,
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

func normalizeFetchSiteSukebeiSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "release_date", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time", "media_birth_time":
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

func buildFetchSiteSukebeiStatusSelected(raws []string, legacy string) map[string]bool {
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
