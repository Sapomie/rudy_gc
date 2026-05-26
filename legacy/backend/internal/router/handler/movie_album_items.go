package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/movie"
)

type movieAlbumItemsListQuery struct {
	Page      int64  `form:"p"`
	PageSize  int64  `form:"ps"`
	Sort      string `form:"sort"`
	Order     string `form:"order"`
	AlbumName string `form:"album_name"`
	Keyword   string `form:"keyword"`
}

type movieAlbumItemsSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type movieAlbumItemsSortQuery struct {
	ByMovieName movieAlbumItemsSortLink
	ByJavID     movieAlbumItemsSortLink
	BySortNo    movieAlbumItemsSortLink
	ByCreated   movieAlbumItemsSortLink
}

func (h *MovieHTMLHandler) MovieAlbumItemsPage(c *gin.Context) {
	var q movieAlbumItemsListQuery
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
	q.Sort = normalizeMovieAlbumItemsSortField(q.Sort)
	q.Order = normalizeMovieAlbumItemsSortOrder(q.Order)

	result, err := h.movieSvc.BuildMovieAlbumItemsPage(c.Request.Context(), movie.MovieAlbumItemsPageQuery{
		AlbumName: strings.TrimSpace(q.AlbumName),
		Page:      q.Page,
		PageSize:  q.PageSize,
		Sort:      q.Sort,
		Order:     q.Order,
		Keyword:   strings.TrimSpace(q.Keyword),
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "电影相册加载失败: %v", err)
		return
	}

	pageInfo := BuildPageInfo(c, result.Total, result.Page, result.PageSize, pageWindow)
	c.HTML(http.StatusOK, "page.movie_album_items_list", gin.H{
		"Title":             "电影相册",
		"PageTitle":         "电影相册列表",
		"PageNote":          "管理按 movie_jav_id 收纳的电影相册。",
		"QuickNavCurrent":   "movie_albums",
		"Query":             q,
		"Rows":              result.Items,
		"Albums":            result.Albums,
		"SelectedAlbumID":   result.SelectedAlbumID,
		"SelectedAlbumName": result.SelectedAlbumName,
		"Total":             result.Total,
		"PageInfo":          pageInfo,
		"SortQuery":         buildMovieAlbumItemsSortQuery(c, q.Sort, q.Order),
	})
}

func normalizeMovieAlbumItemsSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "movie_jav_id", "sort_no", "created_on":
		return strings.TrimSpace(raw)
	default:
		return "created_on"
	}
}

func normalizeMovieAlbumItemsSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func buildMovieAlbumItemsSortQuery(c *gin.Context, currentField string, currentOrder string) *movieAlbumItemsSortQuery {
	makeHref := func(field string) movieAlbumItemsSortLink {
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
		return movieAlbumItemsSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &movieAlbumItemsSortQuery{
		ByMovieName: makeHref("movie_name"),
		ByJavID:     makeHref("movie_jav_id"),
		BySortNo:    makeHref("sort_no"),
		ByCreated:   makeHref("created_on"),
	}
}
