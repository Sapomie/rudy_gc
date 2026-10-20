package list

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/router/handler/common"
	"rudy_gc/local/api/internal/service/movie"
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

type removeTorrentAlbumItemReq struct {
	ItemID int64 `json:"item_id"`
}

type batchRemoveTorrentAlbumItemReq struct {
	ItemIDs []int64 `json:"item_ids"`
}

type batchMoveTorrentAlbumItemReq struct {
	ItemIDs       []int64 `json:"item_ids"`
	TargetAlbumID int64   `json:"target_album_id"`
}

type createTorrentAlbumReq struct {
	Name string `json:"name"`
}

type updateTorrentAlbumReq struct {
	Name string `json:"name"`
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

func (h *ListPagesAPI) TorrentAlbums(c *gin.Context) {
	defer common.WithFrontendPath(c, "/torrentalbumlist")()

	var q torrentAlbumsListQuery
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
	case "javbus_magnet", "sukebei_torrent", "sh_magnet":
		request.SourceType = strings.TrimSpace(q.SourceType)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "来源类型参数错误"})
		return
	}

	if ts, ok, err := common.ParseOptionalDateStart(q.PublishFrom); err != nil {
		common.JSONBadRequest(c, "发布日期开始参数错误", err)
		return
	} else if ok {
		request.PublishTimeFrom = ts
		request.HasPublishTimeFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.PublishTo); err != nil {
		common.JSONBadRequest(c, "发布日期结束参数错误", err)
		return
	} else if ok {
		request.PublishTimeTo = ts
		request.HasPublishTimeTo = true
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.CreatedFrom); err != nil {
		common.JSONBadRequest(c, "下载中时间开始参数错误", err)
		return
	} else if ok {
		request.CreatedOnFrom = ts
		request.HasCreatedOnFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.CreatedTo); err != nil {
		common.JSONBadRequest(c, "下载中时间结束参数错误", err)
		return
	} else if ok {
		request.CreatedOnTo = ts
		request.HasCreatedOnTo = true
	}

	result, err := h.movieSvc.BuildTorrentAlbumsPage(c.Request.Context(), request)
	if err != nil {
		common.JSONServerError(c, "相册列表加载失败", err)
		return
	}

	pageInfo := common.BuildPageInfo(c, result.Total, result.Page, result.PageSize, common.PageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":             "TorrentAlbums",
		"ViewType":          "torrent-albums",
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

func parseTorrentAlbumID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, movie.ErrInvalidAlbumID
	}
	return id, nil
}

func (h *ListPagesAPI) RemoveTorrentAlbumItem(c *gin.Context) {
	albumID, err := parseTorrentAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req removeTorrentAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	removed, err := h.movieSvc.RemoveAlbumItemByID(c.Request.Context(), albumID, req.ItemID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrInvalidAlbumID), errors.Is(err, movie.ErrInvalidAlbumItem), errors.Is(err, movie.ErrAlbumMismatch):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"removed": removed,
		"message": "移除完成",
	})
}

func (h *ListPagesAPI) BatchRemoveTorrentAlbumItems(c *gin.Context) {
	albumID, err := parseTorrentAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req batchRemoveTorrentAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	removedCount, failedIDs, err := h.movieSvc.RemoveAlbumItemsByIDs(c.Request.Context(), albumID, req.ItemIDs)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrInvalidAlbumID):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"removed_count": removedCount,
		"failed_ids":    failedIDs,
		"message":       "批量移除完成",
	})
}

func (h *ListPagesAPI) BatchMoveTorrentAlbumItems(c *gin.Context) {
	albumID, err := parseTorrentAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req batchMoveTorrentAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	movedCount, failedIDs, err := h.movieSvc.MoveAlbumItemsByIDs(c.Request.Context(), albumID, req.TargetAlbumID, req.ItemIDs)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrInvalidAlbumID), errors.Is(err, movie.ErrTargetAlbumID), errors.Is(err, movie.ErrTargetAlbumSame):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":           true,
		"moved_count":  movedCount,
		"failed_ids":   failedIDs,
		"target_album": req.TargetAlbumID,
		"message":      "批量移动完成",
	})
}

func (h *ListPagesAPI) CreateTorrentAlbum(c *gin.Context) {
	var req createTorrentAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	album, err := h.movieSvc.CreateAlbum(c.Request.Context(), req.Name)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrAlbumNameEmpty), errors.Is(err, movie.ErrAlbumNameTooLong), errors.Is(err, movie.ErrAlbumNameExists):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"album_id":   album.ID,
		"album_name": album.Name,
		"message":    "相册创建成功",
	})
}

func (h *ListPagesAPI) DeleteTorrentAlbum(c *gin.Context) {
	albumID, err := parseTorrentAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	result, err := h.movieSvc.DeleteAlbum(c.Request.Context(), albumID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrAlbumInvalidID), errors.Is(err, movie.ErrAlbumDeleteBlocked), errors.Is(err, movie.ErrAlbumNotEmpty):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		case errors.Is(err, movie.ErrAlbumNotFound):
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"deleted":    true,
		"album_id":   result.AlbumID,
		"album_name": result.AlbumName,
		"message":    "相册删除成功",
	})
}

func (h *ListPagesAPI) UpdateTorrentAlbum(c *gin.Context) {
	albumID, err := parseTorrentAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req updateTorrentAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}
	result, err := h.movieSvc.UpdateAlbum(c.Request.Context(), albumID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrAlbumInvalidID), errors.Is(err, movie.ErrAlbumNameEmpty), errors.Is(err, movie.ErrAlbumNameTooLong), errors.Is(err, movie.ErrAlbumNameExists), errors.Is(err, movie.ErrAlbumUpdateBlocked):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		case errors.Is(err, movie.ErrAlbumNotFound):
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"album_id":   result.AlbumID,
		"album_name": result.AlbumName,
		"message":    "相册名称更新成功",
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
		q := common.CloneValues(c)
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
