package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/movie"
)

type torrentAlbumsListQuery struct {
	Page        int64  `form:"p"`
	PageSize    int64  `form:"ps"`
	Sort        string `form:"sort"`
	Order       string `form:"order"`
	AlbumName   string `form:"album_name"`
	Keyword     string `form:"keyword"`
	SourceType  string `form:"source_type"`
	InfoHash    string `form:"info_hash"`
	PublishFrom string `form:"publish_from"`
	PublishTo   string `form:"publish_to"`
	CreatedFrom string `form:"created_from"`
	CreatedTo   string `form:"created_to"`
}

type torrentAlbumsSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type torrentAlbumsSortQuery struct {
	ByMovieName  torrentAlbumsSortLink
	BySourceType torrentAlbumsSortLink
	ByPublish    torrentAlbumsSortLink
	ByCreated    torrentAlbumsSortLink
}

func (h *MovieHTMLHandler) TorrentAlbumsPage(c *gin.Context) {
	var q torrentAlbumsListQuery
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
	q.Sort = normalizeTorrentAlbumsSortField(q.Sort)
	q.Order = normalizeTorrentAlbumsSortOrder(q.Order)

	request := movie.TorrentAlbumsPageQuery{
		AlbumName: strings.TrimSpace(q.AlbumName),
		Page:      q.Page,
		PageSize:  q.PageSize,
		Sort:      q.Sort,
		Order:     q.Order,
		Keyword:   strings.TrimSpace(q.Keyword),
		InfoHash:  strings.TrimSpace(q.InfoHash),
	}
	switch strings.TrimSpace(q.SourceType) {
	case "":
	case "javbus_magnet", "sukebei_torrent", "sehuatang_magnet":
		request.SourceType = strings.TrimSpace(q.SourceType)
	default:
		c.String(http.StatusBadRequest, "来源类型参数错误")
		return
	}

	if ts, ok, err := parseOptionalDateStart(q.PublishFrom); err != nil {
		c.String(http.StatusBadRequest, "发布日期开始参数错误: %v", err)
		return
	} else if ok {
		request.PublishTimeFrom = ts
		request.HasPublishTimeFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.PublishTo); err != nil {
		c.String(http.StatusBadRequest, "发布日期结束参数错误: %v", err)
		return
	} else if ok {
		request.PublishTimeTo = ts
		request.HasPublishTimeTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.CreatedFrom); err != nil {
		c.String(http.StatusBadRequest, "下载中时间开始参数错误: %v", err)
		return
	} else if ok {
		request.CreatedOnFrom = ts
		request.HasCreatedOnFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.CreatedTo); err != nil {
		c.String(http.StatusBadRequest, "下载中时间结束参数错误: %v", err)
		return
	} else if ok {
		request.CreatedOnTo = ts
		request.HasCreatedOnTo = true
	}

	result, err := h.movieSvc.BuildTorrentAlbumsPage(c.Request.Context(), request)
	if err != nil {
		c.String(http.StatusInternalServerError, "相册列表加载失败: %v", err)
		return
	}

	pageInfo := BuildPageInfo(c, result.Total, result.Page, result.PageSize, pageWindow)
	c.HTML(http.StatusOK, "page.torrent_album_list", gin.H{
		"Title":             "TorrentAlbums",
		"PageTitle":         "TorrentAlbums",
		"PageNote":          "展示相册中的资源条目，支持筛选、排序、分页。",
		"QuickNavCurrent":   "torrent_albums",
		"Query":             q,
		"Rows":              result.Items,
		"Albums":            result.Albums,
		"SelectedAlbumID":   result.SelectedAlbumID,
		"SelectedAlbumName": result.SelectedAlbumName,
		"Total":             result.Total,
		"PageInfo":          pageInfo,
		"SortQuery":         buildTorrentAlbumsSortQuery(c, q.Sort, q.Order),
	})
}

func normalizeTorrentAlbumsSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "source_type", "publish_time", "created_on":
		return strings.TrimSpace(raw)
	default:
		return "created_on"
	}
}

func normalizeTorrentAlbumsSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func buildTorrentAlbumsSortQuery(c *gin.Context, currentField string, currentOrder string) *torrentAlbumsSortQuery {
	makeHref := func(field string) torrentAlbumsSortLink {
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
		return torrentAlbumsSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &torrentAlbumsSortQuery{
		ByMovieName:  makeHref("movie_name"),
		BySourceType: makeHref("source_type"),
		ByPublish:    makeHref("publish_time"),
		ByCreated:    makeHref("created_on"),
	}
}
