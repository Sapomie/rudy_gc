package fc2

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"rudy_gc/local/api/internal/router/handler/common"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/service/media"
	movieservice "rudy_gc/local/api/internal/service/movie"
	"rudy_gc/local/api/internal/types"
	fc2domain "rudy_gc/local/api/internal/types/fc2"
)

type FC2API struct {
	deps     *dep.Dep
	mediaSvc *media.Service
	movieSvc *movieservice.Service
}

type addFC2MovieToAlbumReq struct {
	FC2Key    string `json:"fc2_key"`
	AlbumName string `json:"album_name"`
}

type updateFC2ResourceAlbumReq struct {
	SourceRowID int64  `json:"source_row_id"`
	AlbumName   string `json:"album_name"`
}

type fc2ListQuery struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

type fc2ResourceView struct {
	ID           int64           `json:"ID"`
	ThreadTID    string          `json:"ThreadTID"`
	ThreadURL    string          `json:"ThreadURL"`
	ThreadTitle  string          `json:"ThreadTitle"`
	PostTime     int64           `json:"PostTime"`
	MagnetText   string          `json:"MagnetText"`
	InfoHashText string          `json:"InfoHashText"`
	Meta         json.RawMessage `json:"Meta"`
}

type fc2ListRow struct {
	*moviex.FC2Movie
	VideoURL           string
	PlayButtonClass    string
	PlayButtonText     string
	ShowPlayButton     bool
	DownloadTime       int64
	ScTimes            int64
	ComeTimes          int64
	LastSelectedScTime int64
}

type fc2MediaView struct {
	*moviex.FC2Media
	VideoURL        string
	PlayButtonClass string
	PlayButtonText  string
	ShowPlayButton  bool
}

type fc2MediaImageView struct {
	*moviex.FC2MediaImage
	ImageURL string `json:"ImageURL"`
	IsCover  bool   `json:"IsCover"`
}

func NewFC2API(deps *dep.Dep) *FC2API {
	return &FC2API{deps: deps, mediaSvc: media.NewService(deps), movieSvc: movieservice.NewService(deps)}
}

func (h *FC2API) List(c *gin.Context) {
	var query fc2ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		common.JSONBadRequest(c, "FC2列表参数无效", err)
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Status = normalizeFC2Status(query.Status)
	query.Sort = normalizeFC2Sort(query.Sort)
	query.Order = normalizeFC2Order(query.Order)

	total, err := h.deps.FC2MovieModel.CountAll(c.Request.Context(), query.Keyword, query.Status)
	if err != nil {
		common.JSONServerError(c, "FC2影片数量加载失败", err)
		return
	}
	rows, err := h.deps.FC2MovieModel.ListPage(c.Request.Context(), (query.Page-1)*query.PageSize, query.PageSize, query.Keyword, query.Status, query.Sort, query.Order)
	if err != nil {
		common.JSONServerError(c, "FC2影片列表加载失败", err)
		return
	}
	movieIDs := make([]int64, 0, len(rows))
	threadIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.ID > 0 {
			movieIDs = append(movieIDs, row.ID)
			if row.MetadataThreadID > 0 {
				threadIDs = append(threadIDs, row.MetadataThreadID)
			}
		}
	}
	mediaRows, err := h.deps.FC2MediaModel.ListByMovieIDs(c.Request.Context(), movieIDs)
	if err != nil {
		common.JSONServerError(c, "FC2媒体列表加载失败", err)
		return
	}
	mediaByMovieID := make(map[int64]*moviex.FC2Media, len(mediaRows))
	for _, mediaRow := range mediaRows {
		if mediaRow != nil && mediaRow.FC2MovieID > 0 {
			mediaByMovieID[mediaRow.FC2MovieID] = mediaRow
		}
	}
	statRows, err := h.deps.FC2ScStatModel.ListByMovieIDs(c.Request.Context(), movieIDs)
	if err != nil {
		common.JSONServerError(c, "FC2影片 SC 统计加载失败", err)
		return
	}
	statByMovieID := make(map[int64]*moviex.FC2ScStat, len(statRows))
	for _, statRow := range statRows {
		if statRow != nil && statRow.FC2MovieID > 0 {
			statByMovieID[statRow.FC2MovieID] = statRow
		}
	}
	threadRows, err := h.deps.ShSehuatangThreadModel.ListByIDs(c.Request.Context(), threadIDs)
	if err != nil {
		common.JSONServerError(c, "FC2影片来源帖子加载失败", err)
		return
	}
	threadTIDByID := make(map[int64]string, len(threadRows))
	for _, threadRow := range threadRows {
		if threadRow != nil && threadRow.Id > 0 {
			threadTIDByID[threadRow.Id] = threadRow.ThreadTID
		}
	}
	listRows := make([]*fc2ListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		row.MetadataThreadTID = threadTIDByID[row.MetadataThreadID]
		item := &fc2ListRow{FC2Movie: row}
		applyFC2PlayButton(item, mediaByMovieID[row.ID])
		if statRow := statByMovieID[row.ID]; statRow != nil {
			item.ScTimes = statRow.ScTimes
			item.ComeTimes = statRow.ComeTimes
			item.LastSelectedScTime = statRow.LastSelectedScTime
		}
		listRows = append(listRows, item)
	}
	c.JSON(http.StatusOK, gin.H{
		"Title":     "FC2影片列表",
		"PageTitle": "FC2影片列表",
		"PageNote":  "展示由 SH 帖子正文解析生成的 FC2影片主数据。",
		"Rows":      listRows,
		"Total":     total,
		"Query":     query,
		"PageInfo":  common.BuildPageInfo(c, total, query.Page, query.PageSize, common.PageWindow),
	})
}

func applyFC2PlayButton(item *fc2ListRow, media *moviex.FC2Media) {
	if item == nil || media == nil {
		return
	}
	item.DownloadTime = media.BirthTime
	mediaView := buildFC2MediaView(media)
	item.VideoURL = mediaView.VideoURL
	item.PlayButtonClass = mediaView.PlayButtonClass
	item.PlayButtonText = mediaView.PlayButtonText
	item.ShowPlayButton = mediaView.ShowPlayButton
}

func buildFC2MediaView(media *moviex.FC2Media) *fc2MediaView {
	if media == nil {
		return nil
	}
	item := &fc2MediaView{FC2Media: media}
	if strings.TrimSpace(media.FullDir) == "" || strings.TrimSpace(media.FileName) == "" {
		return item
	}
	item.VideoURL = filepath.Join(media.FullDir, media.FileName)
	item.PlayButtonText = "Play"
	item.ShowPlayButton = true
	switch {
	case media.IsRemoved == int64(consts.FilmIsRemoved):
		item.PlayButtonClass = "btn-secondary"
		item.PlayButtonText = "Removed"
	case media.HasSub == int64(consts.FilmHasSub):
		item.PlayButtonClass = "btn-success"
	default:
		item.PlayButtonClass = "btn-info"
	}
	return item
}

func buildFC2MediaImageView(image *moviex.FC2MediaImage, coverImageIDs ...int64) *fc2MediaImageView {
	if image == nil {
		return nil
	}
	coverImageID := int64(0)
	if len(coverImageIDs) > 0 {
		coverImageID = coverImageIDs[0]
	}
	return &fc2MediaImageView{
		FC2MediaImage: image,
		ImageURL:      filepath.ToSlash(strings.TrimSpace(image.FilePath)),
		IsCover:       image.ID > 0 && image.ID == coverImageID,
	}
}

func buildFC2DetailCoverImageView(mediaID int64, imageRows []*moviex.FC2MediaImage) *fc2MediaImageView {
	return buildFC2MediaImageView(media.ResolveFC2DetailCoverImage(mediaID, imageRows))
}

func (h *FC2API) Detail(c *gin.Context) {
	fc2Key, _, _, _, ok := fc2domain.NormalizeKey(c.Param("fc2_key"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FC2影片标识格式无效"})
		return
	}
	movie, err := h.deps.FC2MovieModel.FindOneByKey(c.Request.Context(), fc2Key)
	if err != nil {
		if err == moviex.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "FC2影片不存在"})
			return
		}
		common.JSONServerError(c, "FC2影片详情加载失败", err)
		return
	}
	mediaRow, err := h.deps.FC2MediaModel.FindOneByMovieID(c.Request.Context(), movie.ID)
	if err != nil && err != moviex.ErrNotFound {
		common.JSONServerError(c, "FC2媒体资料加载失败", err)
		return
	}
	stat, err := h.deps.FC2ScStatModel.FindOneByMovieID(c.Request.Context(), movie.ID)
	if err != nil && err != moviex.ErrNotFound {
		common.JSONServerError(c, "FC2影片 SC 统计加载失败", err)
		return
	}
	scEvents, err := h.findFC2ScEvents(c.Request.Context(), movie.ID)
	if err != nil {
		common.JSONServerError(c, "FC2影片 SC 记录加载失败", err)
		return
	}
	mediaImages := make([]*fc2MediaImageView, 0)
	effectiveCoverImageID := int64(0)
	var detailCoverImage *fc2MediaImageView
	if mediaRow != nil {
		imageRows, imageErr := h.deps.FC2MediaImageModel.ListByMediaID(c.Request.Context(), mediaRow.ID)
		if imageErr != nil {
			common.JSONServerError(c, "FC2媒体图片加载失败", imageErr)
			return
		}
		if coverImage := media.ResolveFC2CoverImage(mediaRow.ID, movie.CoverImageID, imageRows); coverImage != nil {
			effectiveCoverImageID = coverImage.ID
		}
		detailCoverImage = buildFC2DetailCoverImageView(mediaRow.ID, imageRows)
		mediaImages = make([]*fc2MediaImageView, 0, len(imageRows))
		for _, imageRow := range imageRows {
			if imageView := buildFC2MediaImageView(imageRow, effectiveCoverImageID); imageView != nil {
				mediaImages = append(mediaImages, imageView)
			}
		}
	}
	threads, err := h.deps.ShSehuatangThreadModel.ListFetchedByFC2MovieID(c.Request.Context(), movie.ID)
	if err != nil {
		common.JSONServerError(c, "FC2影片资源加载失败", err)
		return
	}
	resources := make([]fc2ResourceView, 0, len(threads))
	for _, thread := range threads {
		if thread == nil {
			continue
		}
		meta := json.RawMessage(`{}`)
		if json.Valid([]byte(thread.FC2MetaJSON)) {
			meta = json.RawMessage(thread.FC2MetaJSON)
		}
		resources = append(resources, fc2ResourceView{
			ID: thread.Id, ThreadTID: thread.ThreadTID, ThreadURL: thread.ThreadURL,
			ThreadTitle: thread.ThreadTitle, PostTime: thread.PostTime,
			MagnetText: thread.MagnetText, InfoHashText: thread.InfoHashText, Meta: meta,
		})
	}
	magnetResources, err := h.movieSvc.ListFC2SHMagnetResources(c.Request.Context(), fc2Key)
	if err != nil {
		common.JSONServerError(c, "FC2影片磁力资源加载失败", err)
		return
	}
	movieAlbums, err := h.movieSvc.ListMovieAlbumsByFC2Key(c.Request.Context(), fc2Key)
	if err != nil {
		common.JSONServerError(c, "FC2影片相册状态加载失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Title":                 fc2Key,
		"PageTitle":             fc2Key,
		"PageNote":              "FC2影片资料及其全部 SH 帖子资源版本。",
		"Movie":                 movie,
		"Media":                 buildFC2MediaView(mediaRow),
		"MediaImages":           mediaImages,
		"EffectiveCoverImageID": effectiveCoverImageID,
		"DetailCoverImage":      detailCoverImage,
		"ScStat":                stat,
		"SC":                    scEvents,
		"MagnetResources":       magnetResources,
		"MagnetResourceCount":   len(magnetResources),
		"Resources":             resources,
		"Total":                 len(resources),
		"ResourceCountText":     strconv.Itoa(len(resources)),
		"MovieAlbums":           movieAlbums,
	})
}

func (h *FC2API) AddResourceToAlbum(c *gin.Context) {
	h.updateResourceAlbum(c, true)
}

func (h *FC2API) RemoveResourceFromAlbum(c *gin.Context) {
	h.updateResourceAlbum(c, false)
}

func (h *FC2API) updateResourceAlbum(c *gin.Context, add bool) {
	fc2Key, _, _, _, ok := fc2domain.NormalizeKey(c.Param("fc2_key"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "FC2影片标识格式无效"})
		return
	}
	var req updateFC2ResourceAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	var (
		changed   bool
		albumName string
		err       error
	)
	if add {
		changed, albumName, err = h.movieSvc.AddFC2SHMagnetToAlbum(c.Request.Context(), fc2Key, req.AlbumName, req.SourceRowID)
	} else {
		changed, albumName, err = h.movieSvc.RemoveFC2SHMagnetFromAlbum(c.Request.Context(), fc2Key, req.AlbumName, req.SourceRowID)
	}
	if err != nil {
		switch {
		case errors.Is(err, movieservice.ErrInvalidFC2Key),
			errors.Is(err, movieservice.ErrInvalidFC2Album),
			errors.Is(err, movieservice.ErrInvalidSourceRow),
			errors.Is(err, movieservice.ErrSourceNotFound),
			errors.Is(err, movieservice.ErrSourceMovieMiss):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			common.JSONServerError(c, "FC2磁力相册状态更新失败", err)
		}
		return
	}

	favorited := add
	message := "已加入" + albumName
	if add && !changed {
		message = "该磁力已在" + albumName
	}
	if !add {
		message = "已移出" + albumName
		if !changed {
			message = "该磁力未在" + albumName
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"changed":       changed,
		"favorited":     favorited,
		"album_name":    albumName,
		"source_row_id": req.SourceRowID,
		"message":       message,
	})
}

func (h *FC2API) AddMovieToAlbum(c *gin.Context) {
	fc2Key, _, _, _, ok := fc2domain.NormalizeKey(c.Param("fc2_key"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "FC2影片标识格式无效"})
		return
	}

	var req addFC2MovieToAlbumReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}
	if strings.TrimSpace(req.FC2Key) != "" {
		reqFC2Key, _, _, _, reqOK := fc2domain.NormalizeKey(req.FC2Key)
		if !reqOK || reqFC2Key != fc2Key {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "FC2影片参数不匹配"})
			return
		}
	}

	created, albumName, err := h.movieSvc.AddFC2MovieToAlbum(c.Request.Context(), req.AlbumName, fc2Key)
	if err != nil {
		switch {
		case errors.Is(err, movieservice.ErrMovieAlbumNameEmpty),
			errors.Is(err, movieservice.ErrMovieAlbumNameTooLong),
			errors.Is(err, movieservice.ErrInvalidSourceRow):
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		}
		return
	}
	message := "该FC2影片已在相册：" + albumName
	if created {
		message = "已加入电影相册：" + albumName
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"created":    created,
		"album_name": albumName,
		"message":    message,
	})
}

func (h *FC2API) findFC2ScEvents(ctx context.Context, movieID int64) ([]*types.MovieScEvent, error) {
	rows, err := h.deps.GListModel.ListByFC2MovieID(ctx, movieID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.IsSc != consts.GListIsSc {
			continue
		}
		name := strings.TrimSpace(row.ScName)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	events, err := h.deps.ScModel.ListByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	return buildFC2ScEvents(rows, events), nil
}

func buildFC2ScEvents(rows []*moviex.GList, events []*moviex.GSc) []*types.MovieScEvent {
	eventByName := make(map[string]*moviex.GSc, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}
		eventByName[strings.TrimSpace(event.Name)] = event
	}

	out := make([]*types.MovieScEvent, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.IsSc != consts.GListIsSc {
			continue
		}
		name := strings.TrimSpace(row.ScName)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		item := &types.MovieScEvent{
			ScName: name,
			IsCome: row.IsCome == consts.GListIsCome,
			Href:   buildScEventDetailHrefByName(name),
		}
		if event := eventByName[name]; event != nil {
			item.ScTime = event.ScTime
			item.Cooldown = event.Cooldown
		}
		out = append(out, item)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ScTime == out[j].ScTime {
			return out[i].ScName > out[j].ScName
		}
		return out[i].ScTime > out[j].ScTime
	})
	return out
}

func buildScEventDetailHrefByName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return "/scevent/" + url.PathEscape(name)
}

func (h *FC2API) SetCoverImage(c *gin.Context) {
	imageID, err := strconv.ParseInt(strings.TrimSpace(c.Param("image_id")), 10, 64)
	if err != nil || imageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FC2媒体图片 ID 无效"})
		return
	}
	mediaRow, ok := h.loadFC2MediaForCapture(c)
	if !ok {
		return
	}
	imageRow, err := h.mediaSvc.SetFC2CoverImage(c.Request.Context(), mediaRow, imageID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FC2媒体图片不存在"})
			return
		}
		if errors.Is(err, media.ErrInvalidFC2CaptureRequest) {
			common.JSONBadRequest(c, "设置FC2封面失败", err)
			return
		}
		common.JSONServerError(c, "设置FC2封面失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "image": buildFC2MediaImageView(imageRow, imageRow.ID), "cover_image_id": imageRow.ID})
}

func (h *FC2API) DeleteMediaImage(c *gin.Context) {
	imageID, err := strconv.ParseInt(strings.TrimSpace(c.Param("image_id")), 10, 64)
	if err != nil || imageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FC2媒体图片 ID 无效"})
		return
	}
	mediaRow, ok := h.loadFC2MediaForCapture(c)
	if !ok {
		return
	}
	coverImageID, err := h.mediaSvc.DeleteFC2MediaImage(c.Request.Context(), mediaRow, imageID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FC2媒体图片不存在"})
			return
		}
		common.JSONServerError(c, "FC2媒体图片删除失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "image_id": imageID, "cover_image_id": coverImageID})
}

func (h *FC2API) loadFC2MediaForCapture(c *gin.Context) (*moviex.FC2Media, bool) {
	fc2Key, _, _, _, ok := fc2domain.NormalizeKey(c.Param("fc2_key"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FC2影片标识格式无效"})
		return nil, false
	}
	movieRow, err := h.deps.FC2MovieModel.FindOneByKey(c.Request.Context(), fc2Key)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FC2影片不存在"})
			return nil, false
		}
		common.JSONServerError(c, "FC2影片加载失败", err)
		return nil, false
	}
	mediaRow, err := h.deps.FC2MediaModel.FindOneByMovieID(c.Request.Context(), movieRow.ID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FC2媒体不存在"})
			return nil, false
		}
		common.JSONServerError(c, "FC2媒体加载失败", err)
		return nil, false
	}
	return mediaRow, true
}

func normalizeFC2Status(value string) string {
	switch strings.TrimSpace(value) {
	case "ready", "conflict":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeFC2Sort(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "fc2_key", "title", "performer_text", "post_time", "play", "birth_time", "sc_times", "come_times", "last_selected_sc_time":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "post_time"
	}
}

func normalizeFC2Order(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "asc"
	}
	return "desc"
}
