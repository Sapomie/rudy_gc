package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/sehuatang"
)

type sehuatangMagnetPageQuery struct {
	Page       int64  `form:"p"`
	PageSize   int64  `form:"ps"`
	Sort       string `form:"sort"`
	Order      string `form:"order"`
	Keyword    string `form:"keyword"`
	MovieJavID string `form:"movie_jav_id"`
	InfoHash   string `form:"info_hash"`
}

type sehuatangMagnetSortLink struct {
	Href   string
	Active bool
	Desc   bool
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

	result, err := h.sehuatang.ListPage(c.Request.Context(), sehuatang.ListRequest{
		Page:       q.Page,
		PageSize:   q.PageSize,
		Sort:       q.Sort,
		Order:      q.Order,
		Keyword:    strings.TrimSpace(q.Keyword),
		MovieJavID: strings.TrimSpace(q.MovieJavID),
		InfoHash:   strings.TrimSpace(q.InfoHash),
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Sehuatang 列表加载失败: %v", err)
		return
	}

	pageInfo := buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize)
	c.HTML(http.StatusOK, "page.sehuatang_magnet_list", gin.H{
		"Title":        "Sehuatang Magnet 列表",
		"PageTitle":    "Sehuatang Magnet 列表",
		"PageNote":     "展示 t_sehuatang_magnet 已入库的数据，支持关键词、movie_jav_id、info_hash 筛选以及分页排序。",
		"Query":        q,
		"Rows":         result.Items,
		"Total":        result.Total,
		"PageInfo":     pageInfo,
		"SortQuery":    buildSehuatangMagnetSortQuery(c, q.Sort, q.Order),
		"TaskPageURL":  "/triggers/fetch-sehuatang",
		"TasksPageURL": "/crawler/tasks",
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
