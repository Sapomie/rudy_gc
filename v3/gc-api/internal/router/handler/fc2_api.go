package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	fc2domain "rudy_gc/internal/domain/fc2"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/media"
)

type FC2API struct {
	deps     *dep.Dep
	mediaSvc *media.Service
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
	VideoURL        string
	PlayButtonClass string
	PlayButtonText  string
	ShowPlayButton  bool
	DownloadTime    int64
	ScTimes         int64
	ComeTimes       int64
	LastScTime      int64
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
	return &FC2API{deps: deps, mediaSvc: media.NewService(deps)}
}

func (h *FC2API) List(c *gin.Context) {
	var query fc2ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		jsonBadRequest(c, "FC2列表参数无效", err)
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
		jsonServerError(c, "FC2影片数量加载失败", err)
		return
	}
	rows, err := h.deps.FC2MovieModel.ListPage(c.Request.Context(), (query.Page-1)*query.PageSize, query.PageSize, query.Keyword, query.Status, query.Sort, query.Order)
	if err != nil {
		jsonServerError(c, "FC2影片列表加载失败", err)
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
		jsonServerError(c, "FC2媒体列表加载失败", err)
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
		jsonServerError(c, "FC2影片 SC 统计加载失败", err)
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
		jsonServerError(c, "FC2影片来源帖子加载失败", err)
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
			item.LastScTime = statRow.LastScTime
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
		"PageInfo":  BuildPageInfo(c, total, query.Page, query.PageSize, pageWindow),
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
		jsonServerError(c, "FC2影片详情加载失败", err)
		return
	}
	mediaRow, err := h.deps.FC2MediaModel.FindOneByMovieID(c.Request.Context(), movie.ID)
	if err != nil && err != moviex.ErrNotFound {
		jsonServerError(c, "FC2媒体资料加载失败", err)
		return
	}
	stat, err := h.deps.FC2ScStatModel.FindOneByMovieID(c.Request.Context(), movie.ID)
	if err != nil && err != moviex.ErrNotFound {
		jsonServerError(c, "FC2影片 SC 统计加载失败", err)
		return
	}
	lastScEventName, err := h.findLastFC2ScEventName(c.Request.Context(), movie.ID)
	if err != nil {
		jsonServerError(c, "FC2影片最近 SC 事件加载失败", err)
		return
	}
	mediaImages := make([]*fc2MediaImageView, 0)
	effectiveCoverImageID := int64(0)
	if mediaRow != nil {
		imageRows, imageErr := h.deps.FC2MediaImageModel.ListByMediaID(c.Request.Context(), mediaRow.ID)
		if imageErr != nil {
			jsonServerError(c, "FC2媒体图片加载失败", imageErr)
			return
		}
		if coverImage := media.ResolveFC2CoverImage(mediaRow.ID, movie.CoverImageID, imageRows); coverImage != nil {
			effectiveCoverImageID = coverImage.ID
		}
		mediaImages = make([]*fc2MediaImageView, 0, len(imageRows))
		for _, imageRow := range imageRows {
			if imageView := buildFC2MediaImageView(imageRow, effectiveCoverImageID); imageView != nil {
				mediaImages = append(mediaImages, imageView)
			}
		}
	}
	threads, err := h.deps.ShSehuatangThreadModel.ListFetchedByFC2MovieID(c.Request.Context(), movie.ID)
	if err != nil {
		jsonServerError(c, "FC2影片资源加载失败", err)
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
	c.JSON(http.StatusOK, gin.H{
		"Title":                 fc2Key,
		"PageTitle":             fc2Key,
		"PageNote":              "FC2影片资料及其全部 SH 帖子资源版本。",
		"Movie":                 movie,
		"Media":                 buildFC2MediaView(mediaRow),
		"MediaImages":           mediaImages,
		"EffectiveCoverImageID": effectiveCoverImageID,
		"ScStat":                stat,
		"LastScEventName":       lastScEventName,
		"LastScEventHref":       buildScEventDetailHrefByName(lastScEventName),
		"Resources":             resources,
		"Total":                 len(resources),
		"ResourceCountText":     strconv.Itoa(len(resources)),
	})
}

func (h *FC2API) findLastFC2ScEventName(ctx context.Context, movieID int64) (string, error) {
	rows, err := h.deps.GListModel.ListByFC2MovieID(ctx, movieID)
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.IsSc == consts.GListIsSc && strings.TrimSpace(row.ScName) != "" {
			names = append(names, row.ScName)
		}
	}
	events, err := h.deps.ScModel.ListByNames(ctx, names)
	if err != nil {
		return "", err
	}
	lastName := ""
	lastTime := int64(0)
	for _, event := range events {
		if event != nil && event.ScTime >= lastTime {
			lastTime = event.ScTime
			lastName = event.Name
		}
	}
	return lastName, nil
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
			jsonBadRequest(c, "设置FC2封面失败", err)
			return
		}
		jsonServerError(c, "设置FC2封面失败", err)
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
		jsonServerError(c, "FC2媒体图片删除失败", err)
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
		jsonServerError(c, "FC2影片加载失败", err)
		return nil, false
	}
	mediaRow, err := h.deps.FC2MediaModel.FindOneByMovieID(c.Request.Context(), movieRow.ID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FC2媒体不存在"})
			return nil, false
		}
		jsonServerError(c, "FC2媒体加载失败", err)
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
	case "fc2_key", "title", "performer_text", "post_time", "play", "birth_time":
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
