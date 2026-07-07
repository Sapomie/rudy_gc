package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/spider"
)

type MovieDetailAPI struct {
	movieSvc   *movie.Service
	crawlLogic *spider.CrawlLogic
	runtime    *loop.FetchLoopService
	deps       *dep.Dep
}

func NewMovieDetailAPI(deps *dep.Dep) *MovieDetailAPI {
	return &MovieDetailAPI{
		movieSvc:   movie.NewService(deps),
		crawlLogic: spider.NewCrawlLogic(deps),
		runtime:    newCrawlerRuntime(deps),
		deps:       deps,
	}
}

func (h *MovieDetailAPI) Detail(c *gin.Context) {
	movieName := strings.TrimSpace(c.Param("movie"))
	if movieName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: movie"})
		return
	}

	md, err := h.movieSvc.GetMovieDetailByName(c.Request.Context(), movieName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}
	if md == nil || md.MovieType == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到影片: " + movieName})
		return
	}

	data := gin.H{
		"Title":         md.MovieType.Title,
		"movieType":     md.MovieType,
		"media":         md.MediaInfo,
		"rankInfos":     md.RankInfos,
		"sc":            md.SC,
		"javbusFetch":   md.JavbusFetch,
		"javbusMagnets": md.JavbusMagnets,
		"sukebeiFetch":  md.SukebeiFetch,
		"sukebeiRows":   md.SukebeiTorrents,
		"sehuatangRows": md.SehuatangMagnets,
		"magUpdateDate": md.MovieType.UpdateDate,
	}
	if movieAlbums, err := h.movieSvc.ListMovieAlbumsByMovieJavID(c.Request.Context(), md.MovieType.JavId); err == nil {
		data["MovieAlbums"] = movieAlbums
	}
	if markedDelete, err := h.movieSvc.IsMovieMarkedDelete(c.Request.Context(), md.MovieType.JavId); err == nil {
		data["MovieDeleteAlbumMarked"] = markedDelete
	}

	c.JSON(http.StatusOK, data)
}

func (h *MovieDetailAPI) AddToDownloadLater(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.movieSvc.AddToDownloadLater(c.Request.Context(), javID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}

func (h *MovieDetailAPI) RemoveFromDownloadLater(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	newStatus, err := h.movieSvc.RemoveFromDownloadLater(c.Request.Context(), javID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "needDownload": newStatus})
}

func (h *MovieDetailAPI) DownloadCoverNow(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	if err := h.crawlLogic.DownloadPictureOfMovieByJavId(c.Request.Context(), javID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MovieDetailAPI) MoveWMediaToRemoved(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	created, albumName, err := h.movieSvc.AddMovieToAlbum(c.Request.Context(), consts.MovieDeleteAlbumName, javID)
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

type updateMovieTorrentHashReq struct {
	TorrentHash string `json:"torrent_hash"`
}

type movieAlbumItemReq struct {
	SourceType  string `json:"source_type"`
	SourceRowID int64  `json:"source_row_id"`
	MovieJavID  string `json:"movie_jav_id"`
	AlbumName   string `json:"album_name"`
}

type createMovieAlbumReq struct {
	Name string `json:"name"`
}

type addMovieToMovieAlbumReq struct {
	MovieJavID string `json:"movie_jav_id"`
	AlbumName  string `json:"album_name"`
}

func (h *MovieDetailAPI) UpdateTorrentHash(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	var req updateMovieTorrentHashReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}
	torrentHash := strings.ToUpper(strings.TrimSpace(req.TorrentHash))
	if torrentHash != "" && !isTorrentHash(torrentHash) {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "TorrentHash 必须是 40 位十六进制字符串"})
		return
	}

	mv, err := h.deps.MovieModel.FindOneByJavId(c.Request.Context(), javID)
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

	mediaRow, err := h.deps.WMediaModel.FindOneByMovieJavId(c.Request.Context(), mv.JavId)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "视频信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if mediaRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "视频信息不存在"})
		return
	}

	mediaRow.SourceTorrentHash = torrentHash
	mediaRow.UpdatedOn = time.Now().Unix()
	if err := h.deps.WMediaModel.Update(c.Request.Context(), mediaRow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	h.movieSvc.InvalidateMovieType(c.Request.Context(), mv.JavId)
	c.JSON(http.StatusOK, gin.H{"ok": true, "torrent_hash": torrentHash})
}

func isTorrentHash(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, ch := range value {
		if (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f') {
			continue
		}
		return false
	}
	return true
}

func (h *MovieDetailAPI) AddCast(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	mv, err := h.deps.MovieModel.FindOneByJavId(c.Request.Context(), javID)
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
	if _, err := h.deps.MovieCastModel.FindOneByMovieJavIdCastId(c.Request.Context(), javID, castRow.Id); err != nil {
		if !errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
		if _, err := h.deps.MovieCastModel.Insert(c.Request.Context(), &moviex.AmrMovieCast{
			MovieJavId: javID,
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

	h.movieSvc.InvalidateMovieType(c.Request.Context(), javID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MovieDetailAPI) AddFetchResourceToAlbum(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	var req movieAlbumItemReq
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
		writeFetchResourceAlbumError(c, err)
		return
	}
	if created {
		c.JSON(http.StatusOK, gin.H{"ok": true, "created": true, "favorited": true, "album_name": albumName, "message": "已加入相册：" + albumName})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "created": false, "favorited": true, "album_name": albumName, "message": "该资源已在相册：" + albumName})
}

func (h *MovieDetailAPI) RemoveFetchResourceFromAlbum(c *gin.Context) {
	javID := strings.TrimSpace(c.Param("movie"))
	if javID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 movie 参数"})
		return
	}

	var req movieAlbumItemReq
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
		writeFetchResourceAlbumError(c, err)
		return
	}
	if removed {
		c.JSON(http.StatusOK, gin.H{"ok": true, "removed": true, "favorited": false, "album_name": albumName, "message": "已从相册移除：" + albumName})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "removed": false, "favorited": false, "album_name": albumName, "message": "该资源不在相册：" + albumName})
}

func writeFetchResourceAlbumError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, movie.ErrInvalidSourceType), errors.Is(err, movie.ErrInvalidSourceRow), errors.Is(err, movie.ErrSourceMovieMiss):
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
	case errors.Is(err, movie.ErrSourceNotFound):
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
	}
}

func (h *MovieDetailAPI) AddMovieToMovieAlbum(c *gin.Context) {
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

func (h *MovieDetailAPI) CreateMovieAlbum(c *gin.Context) {
	var req createMovieAlbumReq
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

func (h *MovieDetailAPI) StartCrawlerJob(c *gin.Context) {
	var req loop.StartTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCrawlerError(c, http.StatusBadRequest, "invalid request")
		return
	}

	jobID, err := h.runtime.StartTask(req)
	if err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeCrawlerJobStarted(c, jobID, strings.TrimSpace(req.TaskType))
}
