package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/media"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/spider"
	"rudy_gc/internal/svc"

	"github.com/gin-gonic/gin"
)

type MovieAPI struct {
	movieSvc   *movie.Service
	mediaSvc   *media.Service
	crawlLogic *spider.CrawlLogic
	deps       *svc.Deps
}

func NewMovieAPI(deps *svc.Deps) *MovieAPI {
	return &MovieAPI{
		movieSvc:   movie.NewService(deps),
		mediaSvc:   media.NewService(deps),
		crawlLogic: spider.NewCrawlLogic(deps),
		deps:       deps,
	}
}

func (h *MovieAPI) AddToDownloadLater(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.movieSvc.AddToDownloadLater(c.Request.Context(), javId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}

func (h *MovieAPI) RemoveFromDownloadLater(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.movieSvc.RemoveFromDownloadLater(c.Request.Context(), javId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}

// ===== 新增：下载当前电影封面 =====
// POST /api/movie/:movie/download-cover
func (h *MovieAPI) DownloadCoverNow(c *gin.Context) {
	javId := c.Param("movie")
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	if err := h.crawlLogic.DownloadPictureOfMovieByJavId(c.Request.Context(), javId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/movie/:movie/move-wmedia-removed
func (h *MovieAPI) MoveWMediaToRemoved(c *gin.Context) {
	javId := strings.TrimSpace(c.Param("movie"))
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	created, albumName, err := h.movieSvc.AddMovieToAlbum(c.Request.Context(), consts.MovieDeleteAlbumName, javId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if created {
		c.JSON(http.StatusOK, gin.H{"ok": true, "created": true, "album_name": albumName, "message": "已加入待删除相册：" + albumName})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "created": false, "album_name": albumName, "message": "该电影已在待删除相册：" + albumName})
}

type addMovieCastReq struct {
	Name string `json:"name"`
}

type addMovieAlbumItemReq struct {
	SourceType  string `json:"source_type"`
	SourceRowID int64  `json:"source_row_id"`
	MovieJavID  string `json:"movie_jav_id"`
	AlbumName   string `json:"album_name"`
}

type removeAlbumItemReq struct {
	ItemID int64 `json:"item_id"`
}

type batchRemoveAlbumItemReq struct {
	ItemIDs []int64 `json:"item_ids"`
}

type batchMoveAlbumItemReq struct {
	ItemIDs       []int64 `json:"item_ids"`
	TargetAlbumID int64   `json:"target_album_id"`
}

type createAlbumReq struct {
	Name string `json:"name"`
}

type addMovieToMovieAlbumReq struct {
	MovieJavID string `json:"movie_jav_id"`
	AlbumName  string `json:"album_name"`
}

// POST /api/movie/:movie/add-cast
func (h *MovieAPI) AddCast(c *gin.Context) {
	javId := strings.TrimSpace(c.Param("movie"))
	if javId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	mv, err := h.deps.MovieModel.FindOneByJavId(c.Request.Context(), javId)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "影片不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if mv == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "影片不存在"})
		return
	}

	var req addMovieCastReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "演员名不能为空"})
		return
	}

	castRow, err := h.deps.CastModel.FindOneByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "演员不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if castRow == nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "演员不存在"})
		return
	}

	now := time.Now().Unix()
	if _, err := h.deps.MovieCastModel.FindOneByMovieJavIdCastId(c.Request.Context(), javId, castRow.Id); err != nil {
		if !errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
		if _, err := h.deps.MovieCastModel.Insert(c.Request.Context(), &moviex.AmrMovieCast{
			MovieJavId: javId,
			CastId:     castRow.Id,
			CreatedOn:  now,
			UpdatedOn:  now,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
	}

	movieNumber, ownedMovieNumber, ownedWMediaNumber, err := h.deps.CastModel.GetMovieNumbersWithWMediaByID(c.Request.Context(), castRow.Id, consts.OwnedAllNotRemoved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	castRow.MovieNumber = movieNumber
	castRow.OwnedMovieNumber = ownedMovieNumber
	castRow.OwnedWMediaNumber = ownedWMediaNumber
	castRow.UpdatedOn = now
	if err := h.deps.CastModel.Update(c.Request.Context(), castRow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if err := h.deps.SyncPersonStatsByIDs(c.Request.Context(), []int64{castRow.PersonId}, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	h.movieSvc.InvalidateMovieType(c.Request.Context(), javId)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/movie/:movie/album-item
func (h *MovieAPI) AddFetchResourceToAlbum(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	var req addMovieAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	if req.MovieJavID != "" && !strings.EqualFold(strings.TrimSpace(req.MovieJavID), javID) {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "影片参数不匹配"})
		return
	}

	created, albumName, err := h.movieSvc.AddFetchResourceToAlbum(c.Request.Context(), req.AlbumName, javID, req.SourceType, req.SourceRowID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrInvalidSourceType), errors.Is(err, movie.ErrInvalidSourceRow), errors.Is(err, movie.ErrSourceMovieMiss):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		case errors.Is(err, movie.ErrSourceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	if created {
		c.JSON(http.StatusOK, gin.H{"ok": true, "created": true, "favorited": true, "album_name": albumName, "message": "已加入相册：" + albumName})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "created": false, "favorited": true, "album_name": albumName, "message": "该资源已在相册：" + albumName})
}

// DELETE /api/movie/:movie/album-item
func (h *MovieAPI) RemoveFetchResourceFromAlbum(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	var req addMovieAlbumItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	if req.MovieJavID != "" && !strings.EqualFold(strings.TrimSpace(req.MovieJavID), javID) {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "影片参数不匹配"})
		return
	}

	removed, albumName, err := h.movieSvc.RemoveFetchResourceFromAlbum(c.Request.Context(), req.AlbumName, javID, req.SourceType, req.SourceRowID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrInvalidSourceType), errors.Is(err, movie.ErrInvalidSourceRow), errors.Is(err, movie.ErrSourceMovieMiss):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		case errors.Is(err, movie.ErrSourceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}

	if removed {
		c.JSON(http.StatusOK, gin.H{"ok": true, "removed": true, "favorited": false, "album_name": albumName, "message": "已从相册移除：" + albumName})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "removed": false, "favorited": false, "album_name": albumName, "message": "该资源不在相册：" + albumName})
}

// POST /api/albums/:albumID/items/remove
func (h *MovieAPI) RemoveAlbumItem(c *gin.Context) {
	albumID, err := parseAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req removeAlbumItemReq
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

// POST /api/albums/:albumID/items/batch-remove
func (h *MovieAPI) BatchRemoveAlbumItems(c *gin.Context) {
	albumID, err := parseAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req batchRemoveAlbumItemReq
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

// POST /api/albums/:albumID/items/batch-move
func (h *MovieAPI) BatchMoveAlbumItems(c *gin.Context) {
	albumID, err := parseAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req batchMoveAlbumItemReq
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

// POST /api/albums
func (h *MovieAPI) CreateAlbum(c *gin.Context) {
	var req createAlbumReq
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

// POST /api/movie/:movie/movie-album-item
func (h *MovieAPI) AddMovieToMovieAlbum(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	var req addMovieToMovieAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}
	if req.MovieJavID != "" && !strings.EqualFold(strings.TrimSpace(req.MovieJavID), javID) {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "影片参数不匹配"})
		return
	}

	created, albumName, err := h.movieSvc.AddMovieToAlbum(c.Request.Context(), req.AlbumName, javID)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumNameEmpty), errors.Is(err, movie.ErrMovieAlbumNameTooLong):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}
	if created {
		c.JSON(http.StatusOK, gin.H{"ok": true, "created": true, "album_name": albumName, "message": "已加入电影相册：" + albumName})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "created": false, "album_name": albumName, "message": "该电影已在相册：" + albumName})
}

// POST /api/movie-albums
func (h *MovieAPI) CreateMovieAlbum(c *gin.Context) {
	var req createAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	album, err := h.movieSvc.CreateMovieAlbum(c.Request.Context(), req.Name)
	if err != nil {
		switch {
		case errors.Is(err, movie.ErrMovieAlbumNameEmpty), errors.Is(err, movie.ErrMovieAlbumNameTooLong), errors.Is(err, movie.ErrMovieAlbumNameExists):
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
		"message":    "电影相册创建成功",
	})
}

// POST /api/movie-albums/:albumID/items/remove
func (h *MovieAPI) RemoveMovieAlbumItem(c *gin.Context) {
	albumID, err := parseAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req removeAlbumItemReq
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

// POST /api/movie-albums/:albumID/execute-remove
func (h *MovieAPI) ExecuteMovieAlbumRemove(c *gin.Context) {
	albumID, err := parseAlbumID(c.Param("albumID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	var req batchRemoveAlbumItemReq
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

		result, moveErr := h.mediaSvc.MoveWMediaToRemovedDeferRefresh(c.Request.Context(), strings.TrimSpace(itemRow.MovieJavId))
		if moveErr != nil || result == nil || !result.Ok {
			failedIDs = append(failedIDs, itemID)
			continue
		}

		if _, removeErr := h.movieSvc.RemoveMovieAlbumItemByID(c.Request.Context(), albumID, itemID); removeErr != nil {
			failedIDs = append(failedIDs, itemID)
			continue
		}

		successJavIDs = append(successJavIDs, strings.TrimSpace(itemRow.MovieJavId))
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
		"ok":            true,
		"removed_count": removedCount,
		"failed_ids":    failedIDs,
		"message":       "统一删除完成",
	})
}

// POST /api/movie-albums/:albumID/remove-downloaded-items
func (h *MovieAPI) RemoveDownloadedMovieAlbumItems(c *gin.Context) {
	albumID, err := parseAlbumID(c.Param("albumID"))
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

// POST /api/movie-albums/:albumID/remove-downloaded-items-preview
func (h *MovieAPI) PreviewRemoveDownloadedMovieAlbumItems(c *gin.Context) {
	albumID, err := parseAlbumID(c.Param("albumID"))
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

func parseAlbumID(raw string) (int64, error) {
	albumID, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || albumID <= 0 {
		return 0, movie.ErrInvalidAlbumID
	}
	return albumID, nil
}
