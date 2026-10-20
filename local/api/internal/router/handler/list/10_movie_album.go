package list

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/router/handler/common"
	"rudy_gc/local/api/internal/service/movie"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

type movieAlbumItemsListQuery struct {
	Page          int64  `form:"p"`
	PageSize      int64  `form:"ps"`
	Sort          string `form:"sort"`
	Order         string `form:"order"`
	AlbumName     string `form:"album_name"`
	Keyword       string `form:"keyword"`
	ReleasingFrom string `form:"releasing_from"`
	ReleasingTo   string `form:"releasing_to"`
}

type movieAlbumItemsSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type movieAlbumItemsSortQuery struct {
	ByMovieName     movieAlbumItemsSortLink
	ByJavID         movieAlbumItemsSortLink
	BySortNo        movieAlbumItemsSortLink
	ByReleasingDate movieAlbumItemsSortLink
	ByCreated       movieAlbumItemsSortLink
}

type removeMovieAlbumItemReq struct {
	ItemID int64 `json:"item_id"`
}

type batchMovieAlbumItemReq struct {
	ItemIDs []int64 `json:"item_ids"`
}

type batchTransferMovieAlbumItemReq struct {
	ItemIDs       []int64 `json:"item_ids"`
	TargetAlbumID int64   `json:"target_album_id"`
}

type updateMovieAlbumReq struct {
	Name string `json:"name"`
}

func (h *ListPagesAPI) MovieAlbumItems(c *gin.Context) {
	defer common.WithFrontendPath(c, "/moviealbumlist")()

	var q movieAlbumItemsListQuery
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
	q.Sort = normalizeMovieAlbumItemsSortField(q.Sort)
	q.Order = normalizeMovieAlbumItemsSortOrder(q.Order)

	pageQuery := movie.MovieAlbumItemsPageQuery{
		AlbumName: strings.TrimSpace(q.AlbumName),
		Page:      q.Page,
		PageSize:  q.PageSize,
		Sort:      q.Sort,
		Order:     q.Order,
		Keyword:   strings.TrimSpace(q.Keyword),
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.ReleasingFrom); err != nil {
		common.JSONBadRequest(c, "上映开始日期错误", err)
		return
	} else if ok {
		pageQuery.ReleasingDateFrom = ts
		pageQuery.HasReleasingDateFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.ReleasingTo); err != nil {
		common.JSONBadRequest(c, "上映结束日期错误", err)
		return
	} else if ok {
		pageQuery.ReleasingDateTo = ts
		pageQuery.HasReleasingDateTo = true
	}

	result, err := h.movieSvc.BuildMovieAlbumItemsPage(c.Request.Context(), pageQuery)
	if err != nil {
		common.JSONServerError(c, "电影相册加载失败", err)
		return
	}

	pageInfo := common.BuildPageInfo(c, result.Total, result.Page, result.PageSize, common.PageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":             "电影相册",
		"ViewType":          "movie-albums",
		"PageTitle":         "电影相册列表",
		"PageNote":          "统一管理 JAV 与 FC2 影片收藏。",
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

func parseMovieAlbumID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, movie.ErrMovieAlbumInvalidID
	}
	return id, nil
}

func (h *ListPagesAPI) UpdateMovieAlbum(c *gin.Context) {
	albumID, err := parseMovieAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req updateMovieAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	result, err := h.movieSvc.UpdateMovieAlbum(c.Request.Context(), albumID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumInvalidID),
			errors.Is(err, movie.ErrMovieAlbumNameEmpty),
			errors.Is(err, movie.ErrMovieAlbumNameTooLong),
			errors.Is(err, movie.ErrMovieAlbumNameExists),
			errors.Is(err, movie.ErrMovieAlbumUpdateBlocked):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		case errors.Is(err, movie.ErrMovieAlbumNotFound):
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"album_id":   result.ID,
		"album_name": result.Name,
		"message":    "电影相册更新成功",
	})
}

func (h *ListPagesAPI) DeleteMovieAlbum(c *gin.Context) {
	albumID, err := parseMovieAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	result, err := h.movieSvc.DeleteMovieAlbum(c.Request.Context(), albumID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumInvalidID),
			errors.Is(err, movie.ErrMovieAlbumDeleteBlocked),
			errors.Is(err, movie.ErrMovieAlbumNotEmpty):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		case errors.Is(err, movie.ErrMovieAlbumNotFound):
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"album_id":   result.ID,
		"album_name": result.Name,
		"message":    "电影相册删除成功",
	})
}

func (h *ListPagesAPI) RemoveMovieAlbumItem(c *gin.Context) {
	albumID, err := parseMovieAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req removeMovieAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	removed, err := h.movieSvc.RemoveMovieAlbumItemByID(c.Request.Context(), albumID, req.ItemID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumInvalidID), errors.Is(err, movie.ErrMovieAlbumItemInvalid):
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

func (h *ListPagesAPI) BatchMoveMovieAlbumItems(c *gin.Context) {
	h.batchTransferMovieAlbumItems(c, true)
}

func (h *ListPagesAPI) BatchCopyMovieAlbumItems(c *gin.Context) {
	h.batchTransferMovieAlbumItems(c, false)
}

func (h *ListPagesAPI) batchTransferMovieAlbumItems(c *gin.Context, move bool) {
	albumID, err := parseMovieAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req batchTransferMovieAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}
	if len(req.ItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少待处理条目"})
		return
	}

	var result *movie.TransferMovieAlbumItemsResult
	if move {
		result, err = h.movieSvc.MoveMovieAlbumItemsByIDs(c.Request.Context(), albumID, req.TargetAlbumID, req.ItemIDs)
	} else {
		result, err = h.movieSvc.CopyMovieAlbumItemsByIDs(c.Request.Context(), albumID, req.TargetAlbumID, req.ItemIDs)
	}
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumInvalidID), errors.Is(err, movie.ErrMovieAlbumTargetInvalid), errors.Is(err, movie.ErrMovieAlbumTargetSame):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	action := "复制"
	countKey := "copied_count"
	if move {
		action = "移动"
		countKey = "moved_count"
	}
	response := gin.H{
		"ok":              true,
		"processed_count": result.ProcessedCount,
		"skipped_count":   result.SkippedCount,
		"failed_ids":      result.FailedIDs,
		"target_album_id": req.TargetAlbumID,
		"message":         "批量" + action + "完成",
	}
	response[countKey] = result.ProcessedCount
	c.JSON(http.StatusOK, response)
}

func (h *ListPagesAPI) ExecuteMovieAlbumRemove(c *gin.Context) {
	albumID, err := parseMovieAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req batchMovieAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}
	if len(req.ItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少待删除条目"})
		return
	}

	albumRow, err := h.deps.MovieAlbumModel.FindOne(c.Request.Context(), albumID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if albumRow == nil || strings.TrimSpace(albumRow.Name) != consts.MovieDeleteAlbumName {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "当前电影相册不支持统一删除"})
		return
	}

	successJavIDs := make([]string, 0, len(req.ItemIDs))
	failedIDs := make([]int64, 0)
	seen := make(map[int64]struct{}, len(req.ItemIDs))
	removedCount := int64(0)
	removedJAVCount := int64(0)
	removedFC2Count := int64(0)
	for _, itemID := range req.ItemIDs {
		if itemID <= 0 {
			continue
		}
		if _, ok := seen[itemID]; ok {
			continue
		}
		seen[itemID] = struct{}{}

		itemRow, findErr := h.deps.MovieAlbumItemModel.FindOne(c.Request.Context(), itemID)
		if findErr != nil || itemRow == nil || itemRow.AlbumId != albumID {
			failedIDs = append(failedIDs, itemID)
			continue
		}
		movieType := strings.TrimSpace(itemRow.MovieType)
		if movieType == "" && strings.TrimSpace(itemRow.FC2Key) != "" {
			movieType = moviedomain.SehuatangThreadMovieTypeFC2PPV
		}
		if movieType == "" {
			movieType = moviedomain.SehuatangThreadMovieTypeJAV
		}

		moved := false
		movedJavID := ""
		switch movieType {
		case moviedomain.SehuatangThreadMovieTypeJAV:
			javID := strings.TrimSpace(itemRow.MovieJavId)
			if javID == "" {
				break
			}
			result, moveErr := h.mediaSvc.MoveWMediaToRemovedDeferRefresh(c.Request.Context(), javID)
			if moveErr == nil && result != nil && result.Ok {
				moved = true
				movedJavID = javID
			}
		case moviedomain.SehuatangThreadMovieTypeFC2PPV:
			fc2Key := strings.TrimSpace(itemRow.FC2Key)
			if fc2Key == "" {
				break
			}
			result, moveErr := h.mediaSvc.MoveFC2MediaToRemoved(c.Request.Context(), fc2Key)
			if moveErr == nil && result != nil && result.Ok {
				moved = true
			}
		}
		if !moved {
			failedIDs = append(failedIDs, itemID)
			continue
		}

		if _, removeErr := h.movieSvc.RemoveMovieAlbumItemByID(c.Request.Context(), albumID, itemID); removeErr != nil {
			failedIDs = append(failedIDs, itemID)
			continue
		}

		if movedJavID != "" {
			successJavIDs = append(successJavIDs, movedJavID)
			removedJAVCount++
		} else {
			removedFC2Count++
		}
		removedCount++
	}

	if len(successJavIDs) > 0 {
		if err := h.mediaSvc.FinalizeMoveRemovedBatch(c.Request.Context(), successJavIDs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"ok":            false,
				"error":         err.Error(),
				"removed_count": removedCount,
				"failed_ids":    failedIDs,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":                true,
		"removed_count":     removedCount,
		"removed_jav_count": removedJAVCount,
		"removed_fc2_count": removedFC2Count,
		"failed_count":      len(failedIDs),
		"failed_ids":        failedIDs,
		"message":           "统一删除完成",
	})
}

func (h *ListPagesAPI) PreviewRemoveDownloadedMovieAlbumItems(c *gin.Context) {
	albumID, err := parseMovieAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	result, err := h.movieSvc.PreviewRemoveDownloadedMovieAlbumItems(c.Request.Context(), albumID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumInvalidID), errors.Is(err, movie.ErrMovieAlbumUnsupportedRemoveDownloaded):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"count": result.Count,
		"items": result.Items,
	})
}

func (h *ListPagesAPI) RemoveDownloadedMovieAlbumItems(c *gin.Context) {
	albumID, err := parseMovieAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	result, err := h.movieSvc.RemoveDownloadedMovieAlbumItems(c.Request.Context(), albumID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumInvalidID), errors.Is(err, movie.ErrMovieAlbumUnsupportedRemoveDownloaded):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"removed_count": result.RemovedCount,
		"skipped_count": result.SkippedCount,
		"message":       fmt.Sprintf("已移除 %d 条已下载条目", result.RemovedCount),
	})
}

func normalizeMovieAlbumItemsSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "movie_jav_id", "sort_no", "releasing_date", "created_on":
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
		return movieAlbumItemsSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &movieAlbumItemsSortQuery{
		ByMovieName:     makeHref("movie_name"),
		ByJavID:         makeHref("movie_jav_id"),
		BySortNo:        makeHref("sort_no"),
		ByReleasingDate: makeHref("releasing_date"),
		ByCreated:       makeHref("created_on"),
	}
}
