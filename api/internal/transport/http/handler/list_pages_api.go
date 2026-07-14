package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/service/media"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/types"
)

type ListPagesAPI struct {
	deps     *dep.Dep
	mediaSvc *media.Service
	movieSvc *movie.Service
}

func NewListPagesAPI(deps *dep.Dep) *ListPagesAPI {
	return &ListPagesAPI{
		deps:     deps,
		mediaSvc: media.NewService(deps),
		movieSvc: movie.NewService(deps),
	}
}

type mediaListQuery struct {
	Page            int64  `form:"p"`
	PageSize        int64  `form:"ps"`
	Sort            string `form:"sort"`
	Order           string `form:"order"`
	Keyword         string `form:"keyword"`
	SizeMin         string `form:"size_min"`
	SizeMax         string `form:"size_max"`
	HeightMin       string `form:"height_min"`
	HeightMax       string `form:"height_max"`
	DurationMin     string `form:"duration_min"`
	DurationMax     string `form:"duration_max"`
	BitRateMin      string `form:"bit_rate_min"`
	BitRateMax      string `form:"bit_rate_max"`
	FrameAverageMin string `form:"frame_average_min"`
	FrameAverageMax string `form:"frame_average_max"`
	SelfMake        string `form:"self_make"`
	HasMask         string `form:"has_mask"`
	ScTimesMin      string `form:"sc_times_min"`
	ScTimesMax      string `form:"sc_times_max"`
	LastScFrom      string `form:"last_sc_from"`
	LastScTo        string `form:"last_sc_to"`
	BirthFrom       string `form:"birth_from"`
	BirthTo         string `form:"birth_to"`
	ReleasingFrom   string `form:"releasing_from"`
	ReleasingTo     string `form:"releasing_to"`
}

func (h *ListPagesAPI) MediaList(c *gin.Context) {
	defer withFrontendPath(c, "/medialist")()

	var q mediaListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 24
	}
	q.Sort = normalizeFilmSortFieldForPage(q.Sort)
	q.Order = normalizeFilmSortOrderForPage(q.Order)

	filter, err := buildMediaListFilter(q)
	if err != nil {
		jsonBadRequest(c, "筛选参数错误", err)
		return
	}

	items, total, err := h.mediaSvc.ListMediaPageWithFilter(
		c.Request.Context(),
		q.Page,
		q.PageSize,
		buildMediaOrderByForPage(q.Sort, q.Order),
		filter,
	)
	if err != nil {
		jsonServerError(c, "媒体列表加载失败", err)
		return
	}

	pi := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":         "Medias",
		"ViewType":      "medias",
		"Items":         items,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"FilmSortQuery": buildFilmSortQuery(c, q.Sort, q.Order),
		"Query":         q,
	})
}

func (h *ListPagesAPI) ProbeMediaMeta(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "media id 无效"})
		return
	}

	result, err := h.mediaSvc.ProbeMediaMetaByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) || errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "media 不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":     true,
		"result": result,
	})
}

func buildMediaListFilter(q mediaListQuery) (types.MediaListFilter, error) {
	filter := types.MediaListFilter{
		MovieNameKeyword: strings.TrimSpace(q.Keyword),
	}

	if v, ok, err := parseOptionalNonNegativeFloat64(q.SizeMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SizeMin = gbToBytes(v)
		filter.HasSizeMin = true
	}
	if v, ok, err := parseOptionalNonNegativeFloat64(q.SizeMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SizeMax = gbToBytes(v)
		filter.HasSizeMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.HeightMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HeightMin = v
		filter.HasHeightMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.HeightMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HeightMax = v
		filter.HasHeightMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.DurationMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.DurationMin = minutesToSeconds(v)
		filter.HasDurationMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.DurationMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.DurationMax = minutesToSeconds(v)
		filter.HasDurationMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.BitRateMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BitRateMin = v
		filter.HasBitRateMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.BitRateMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BitRateMax = v
		filter.HasBitRateMax = true
	}

	if v, ok, err := parseOptionalNonNegativeFloat64(q.FrameAverageMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.FrameAverageMin = v
		filter.HasFrameAverageMin = true
	}
	if v, ok, err := parseOptionalNonNegativeFloat64(q.FrameAverageMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.FrameAverageMax = v
		filter.HasFrameAverageMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.SelfMake); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SelfMake = v
		filter.HasSelfMake = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.HasMask); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HasMask = v
		filter.HasHasMask = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ScTimesMin = v
		filter.HasScTimesMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ScTimesMax = v
		filter.HasScTimesMax = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.LastScFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.LastScFrom = v
		filter.HasLastScFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.LastScTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.LastScTo = v
		filter.HasLastScTo = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.BirthFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BirthTimeFrom = v
		filter.HasBirthTimeFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.BirthTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BirthTimeTo = v
		filter.HasBirthTimeTo = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.ReleasingFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ReleasingDateFrom = v
		filter.HasReleasingDateFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.ReleasingTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ReleasingDateTo = v
		filter.HasReleasingDateTo = true
	}

	return filter, nil
}

func buildMediaOrderByForPage(sortField, sortOrder string) string {
	field := normalizeFilmSortFieldForPage(sortField)
	order := "DESC"
	if normalizeFilmSortOrderForPage(sortOrder) == "asc" {
		order = "ASC"
	}

	switch field {
	case "movie_name":
		return "wm.movie_name " + order + ", wm.birth_time DESC, wm.id DESC"
	case "size":
		return "wm.size " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "height":
		return "wm.height " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "duration":
		return "wm.duration " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "bit_rate":
		return "wm.bit_rate " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "frame_average":
		return "wm.frame_average " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "self_make":
		return "wm.self_make " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "has_mask":
		return "wm.has_mask " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "sc_times":
		return "COALESCE(gss.sc_times, 0) " + order + ", COALESCE(gss.last_sc_time, 0) DESC, wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "last_sc_time":
		return "COALESCE(gss.last_sc_time, 0) " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "releasing_date":
		return "wm.releasing_date " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	default:
		return "wm.birth_time " + order + ", wm.movie_name DESC, wm.id DESC"
	}
}

type filmSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type filmSortQuery struct {
	ByMovieName     filmSortLink
	BySize          filmSortLink
	ByHeight        filmSortLink
	ByDuration      filmSortLink
	ByBitRate       filmSortLink
	ByFrameAverage  filmSortLink
	BySelfMake      filmSortLink
	ByHasMask       filmSortLink
	ByScTimes       filmSortLink
	ByLastSc        filmSortLink
	ByBirthTime     filmSortLink
	ByReleasingDate filmSortLink
}

func buildFilmSortQuery(c *gin.Context, currentField, currentOrder string) *filmSortQuery {
	makeHref := func(field string) filmSortLink {
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

		return filmSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &filmSortQuery{
		ByMovieName:     makeHref("movie_name"),
		BySize:          makeHref("size"),
		ByHeight:        makeHref("height"),
		ByDuration:      makeHref("duration"),
		ByBitRate:       makeHref("bit_rate"),
		ByFrameAverage:  makeHref("frame_average"),
		BySelfMake:      makeHref("self_make"),
		ByHasMask:       makeHref("has_mask"),
		ByScTimes:       makeHref("sc_times"),
		ByLastSc:        makeHref("last_sc_time"),
		ByBirthTime:     makeHref("birth_time"),
		ByReleasingDate: makeHref("releasing_date"),
	}
}

func normalizeFilmSortFieldForPage(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "movie_name", "size", "height", "duration", "bit_rate", "frame_average", "self_make", "has_mask", "sc_times", "last_sc_time", "birth_time", "releasing_date":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "birth_time"
	}
}

func normalizeFilmSortOrderForPage(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}

type wAggEventListQuery struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
	AggKey   string `form:"agg_key"`
	FlowKey  string `form:"flow_key"`
	Status   string `form:"status"`
}

type wAggEventListRow struct {
	StartedTime         int64
	AggKey              string
	FlowKey             string
	ScopeCount          int64
	BucketCount         int64
	TopCount            int64
	FinishedTime        int64
	DurationSecondsText string
}

type wAggEventSortLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

type wAggEventHeaderLinks struct {
	StartedTime  wAggEventSortLink
	AggKey       wAggEventSortLink
	FlowKey      wAggEventSortLink
	ScopeCount   wAggEventSortLink
	BucketCount  wAggEventSortLink
	TopCount     wAggEventSortLink
	FinishedTime wAggEventSortLink
	Duration     wAggEventSortLink
}

func (h *ListPagesAPI) WAggEventList(c *gin.Context) {
	defer withFrontendPath(c, "/aggeventlist")()

	var q wAggEventListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	q.Sort = normalizeWAggEventSort(q.Sort)
	q.Order = normalizeWAggEventOrder(q.Order)
	q.AggKey = strings.TrimSpace(q.AggKey)
	q.FlowKey = strings.TrimSpace(q.FlowKey)
	q.Status = strings.TrimSpace(q.Status)

	rows, total, err := h.deps.WAggEventModel.ListPage(c.Request.Context(), moviex.WAggEventListFilter{
		AggKey:   q.AggKey,
		FlowKey:  q.FlowKey,
		Status:   q.Status,
		Sort:     q.Sort,
		Order:    q.Order,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		jsonServerError(c, "w_agg_event 列表加载失败", err)
		return
	}

	pageInfo := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":       "WAggEvents",
		"ViewType":    "w-agg-events",
		"PageTitle":   "WAggEvent 列表",
		"PageNote":    "直接查看 w_agg_event 落库记录。",
		"Rows":        buildWAggEventListRows(rows),
		"Total":       total,
		"PageInfo":    pageInfo,
		"pageInfo":    pageInfo,
		"Query":       q,
		"HeaderLinks": buildWAggEventHeaderLinks(c, q.Sort, q.Order),
		"ClearHref":   "/aggeventlist",
		"MenuListID":  "w_agg_events",
	})
}

func buildWAggEventListRows(rows []*moviex.WAggEvent) []*wAggEventListRow {
	if len(rows) == 0 {
		return []*wAggEventListRow{}
	}

	out := make([]*wAggEventListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &wAggEventListRow{
			StartedTime:         row.StartedTime,
			AggKey:              row.AggKey,
			FlowKey:             row.FlowKey,
			ScopeCount:          row.ScopeCount,
			BucketCount:         row.BucketCount,
			TopCount:            row.TopCount,
			FinishedTime:        row.FinishedTime,
			DurationSecondsText: formatWAggEventDurationSeconds(row.DurationMs),
		})
	}
	return out
}

func formatWAggEventDurationSeconds(durationMs int64) string {
	if durationMs <= 0 {
		return "-"
	}
	if durationMs%1000 == 0 {
		return fmt.Sprintf("%d", durationMs/1000)
	}
	seconds := float64(durationMs) / 1000.0
	text := fmt.Sprintf("%.1f", seconds)
	text = strings.TrimRight(text, "0")
	return strings.TrimRight(text, ".")
}

func normalizeWAggEventSort(sortField string) string {
	switch strings.TrimSpace(sortField) {
	case "scope_count":
		return "scope_count"
	case "bucket_count":
		return "bucket_count"
	case "top_count":
		return "top_count"
	case "finished_time":
		return "finished_time"
	case "duration_ms":
		return "duration_ms"
	case "agg_key":
		return "agg_key"
	case "flow_key":
		return "flow_key"
	default:
		return "started_time"
	}
}

func normalizeWAggEventOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "asc"
	}
	return "desc"
}

func buildWAggEventHeaderLinks(c *gin.Context, currentSort string, currentOrder string) wAggEventHeaderLinks {
	return wAggEventHeaderLinks{
		StartedTime:  buildWAggEventSortLink(c, currentSort, currentOrder, "started_time", "StartedTime"),
		AggKey:       buildWAggEventSortLink(c, currentSort, currentOrder, "agg_key", "AggKey"),
		FlowKey:      buildWAggEventSortLink(c, currentSort, currentOrder, "flow_key", "FlowKey"),
		ScopeCount:   buildWAggEventSortLink(c, currentSort, currentOrder, "scope_count", "ScopeCount"),
		BucketCount:  buildWAggEventSortLink(c, currentSort, currentOrder, "bucket_count", "BucketCount"),
		TopCount:     buildWAggEventSortLink(c, currentSort, currentOrder, "top_count", "TopCount"),
		FinishedTime: buildWAggEventSortLink(c, currentSort, currentOrder, "finished_time", "FinishedTime"),
		Duration:     buildWAggEventSortLink(c, currentSort, currentOrder, "duration_ms", "Duration(秒)"),
	}
}

func buildWAggEventSortLink(c *gin.Context, currentSort string, currentOrder string, sortField string, label string) wAggEventSortLink {
	nextOrder := "desc"
	active := currentSort == sortField
	if active && currentOrder == "desc" {
		nextOrder = "asc"
	}

	q := c.Request.URL.Query()
	q.Set("p", "1")
	q.Set("sort", sortField)
	q.Set("order", nextOrder)
	href := c.Request.URL.Path
	if encoded := q.Encode(); encoded != "" {
		href += "?" + encoded
	}

	return wAggEventSortLink{
		Label:  label,
		Href:   href,
		Active: active,
		Desc:   active && currentOrder == "desc",
	}
}

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
	defer withFrontendPath(c, "/torrentalbumlist")()

	var q torrentAlbumsListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
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

	if ts, ok, err := parseOptionalDateStart(q.PublishFrom); err != nil {
		jsonBadRequest(c, "发布日期开始参数错误", err)
		return
	} else if ok {
		request.PublishTimeFrom = ts
		request.HasPublishTimeFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.PublishTo); err != nil {
		jsonBadRequest(c, "发布日期结束参数错误", err)
		return
	} else if ok {
		request.PublishTimeTo = ts
		request.HasPublishTimeTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.CreatedFrom); err != nil {
		jsonBadRequest(c, "下载中时间开始参数错误", err)
		return
	} else if ok {
		request.CreatedOnFrom = ts
		request.HasCreatedOnFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.CreatedTo); err != nil {
		jsonBadRequest(c, "下载中时间结束参数错误", err)
		return
	} else if ok {
		request.CreatedOnTo = ts
		request.HasCreatedOnTo = true
	}

	result, err := h.movieSvc.BuildTorrentAlbumsPage(c.Request.Context(), request)
	if err != nil {
		jsonServerError(c, "相册列表加载失败", err)
		return
	}

	pageInfo := BuildPageInfo(c, result.Total, result.Page, result.PageSize, pageWindow)
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
	defer withFrontendPath(c, "/moviealbumlist")()

	var q movieAlbumItemsListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
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
	if ts, ok, err := parseOptionalDateStart(q.ReleasingFrom); err != nil {
		jsonBadRequest(c, "上映开始日期错误", err)
		return
	} else if ok {
		pageQuery.ReleasingDateFrom = ts
		pageQuery.HasReleasingDateFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.ReleasingTo); err != nil {
		jsonBadRequest(c, "上映结束日期错误", err)
		return
	} else if ok {
		pageQuery.ReleasingDateTo = ts
		pageQuery.HasReleasingDateTo = true
	}

	result, err := h.movieSvc.BuildMovieAlbumItemsPage(c.Request.Context(), pageQuery)
	if err != nil {
		jsonServerError(c, "电影相册加载失败", err)
		return
	}

	pageInfo := BuildPageInfo(c, result.Total, result.Page, result.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":             "电影相册",
		"ViewType":          "movie-albums",
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
		ByMovieName:     makeHref("movie_name"),
		ByJavID:         makeHref("movie_jav_id"),
		BySortNo:        makeHref("sort_no"),
		ByReleasingDate: makeHref("releasing_date"),
		ByCreated:       makeHref("created_on"),
	}
}

type castListQuery struct {
	Page         int64  `form:"p"`
	PageSize     int64  `form:"ps"`
	Sort         string `form:"sort"`
	Order        string `form:"order"`
	Keyword      string `form:"keyword"`
	OwnedMin     string `form:"wowned_min"`
	OwnedMax     string `form:"wowned_max"`
	LegacyMin    string `form:"owned_min"`
	LegacyMax    string `form:"owned_max"`
	ScTimesMin   string `form:"sc_times_min"`
	ScTimesMax   string `form:"sc_times_max"`
	ComeTimesMin string `form:"come_times_min"`
	ComeTimesMax string `form:"come_times_max"`
	LastScFrom   string `form:"last_sc_from"`
	LastScTo     string `form:"last_sc_to"`
}

func (h *ListPagesAPI) CastList(c *gin.Context) {
	defer withFrontendPath(c, "/castlist")()

	var q castListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 24
	}
	if strings.TrimSpace(q.OwnedMin) == "" {
		q.OwnedMin = strings.TrimSpace(q.LegacyMin)
	}
	if strings.TrimSpace(q.OwnedMax) == "" {
		q.OwnedMax = strings.TrimSpace(q.LegacyMax)
	}
	q.Sort = normalizeCastSortFieldForPage(q.Sort)
	q.Order = normalizeCastSortOrderForPage(q.Order)

	filter, err := buildCastListFilter(q)
	if err != nil {
		jsonBadRequest(c, "筛选参数错误", err)
		return
	}
	total, err := h.deps.PersonModel.CountAll(c.Request.Context(), filter)
	if err != nil {
		jsonServerError(c, "演员列表加载失败", err)
		return
	}

	offset := (q.Page - 1) * q.PageSize
	items, err := h.deps.PersonModel.ListPage(
		c.Request.Context(),
		offset,
		q.PageSize,
		buildCastOrderByForPage(q.Sort, q.Order),
		filter,
	)
	if err != nil {
		jsonServerError(c, "演员列表加载失败", err)
		return
	}

	pi := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":         "Casts",
		"ViewType":      "casts",
		"Items":         items,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"CastSortQuery": buildCastSortQuery(c, q.Sort, q.Order),
		"Query":         q,
	})
}

func buildCastListFilter(q castListQuery) (types.PersonListFilter, error) {
	filter := types.PersonListFilter{
		Keyword: strings.TrimSpace(q.Keyword),
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.OwnedMin); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.OwnedMin = v
		filter.HasOwnedMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.OwnedMax); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.OwnedMax = v
		filter.HasOwnedMax = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMin); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ScTimesMin = v
		filter.HasScTimesMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMax); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ScTimesMax = v
		filter.HasScTimesMax = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ComeTimesMin); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ComeTimesMin = v
		filter.HasComeTimesMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ComeTimesMax); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ComeTimesMax = v
		filter.HasComeTimesMax = true
	}

	if strings.TrimSpace(q.LastScFrom) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(q.LastScFrom), time.Local)
		if err != nil {
			return types.PersonListFilter{}, err
		}
		filter.LastScFrom = t.Unix()
		filter.HasLastScFrom = true
	}

	if strings.TrimSpace(q.LastScTo) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(q.LastScTo), time.Local)
		if err != nil {
			return types.PersonListFilter{}, err
		}
		filter.LastScTo = t.Add(24*time.Hour - time.Second).Unix()
		filter.HasLastScTo = true
	}

	return filter, nil
}

func buildCastOrderByForPage(sortField, sortOrder string) string {
	field := normalizeCastSortFieldForPage(sortField)
	order := "DESC"
	if normalizeCastSortOrderForPage(sortOrder) == "asc" {
		order = "ASC"
	}

	column := "p.owned_w_media_number"
	switch field {
	case "name":
		column = "p.name"
	case "chinese":
		column = "p.chinese"
	case "age":
		if order == "ASC" {
			return "CASE WHEN p.birth_day > 0 THEN 0 ELSE 1 END ASC, p.birth_day DESC, p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
		}
		return "CASE WHEN p.birth_day > 0 THEN 0 ELSE 1 END ASC, p.birth_day ASC, p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
	case "height":
		return "CASE WHEN p.height > 0 THEN 0 ELSE 1 END ASC, p.height " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
	case "movie_number":
		column = "p.movie_number"
	case "owned_w_media_number":
		column = "p.owned_w_media_number"
	case "sc_times":
		column = "p.sc_times"
	case "come_times":
		column = "p.come_times"
	case "last_sc_time":
		column = "p.last_sc_time"
	case "highest_rank":
		column = "p.highest_rank"
	}

	if column == "p.owned_w_media_number" {
		return column + " " + order + ", p.movie_number DESC, p.name ASC, p.id DESC"
	}
	if column == "p.name" {
		return column + " " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.id DESC"
	}
	if column == "p.chinese" {
		return "CASE WHEN p.chinese <> '' THEN 0 ELSE 1 END ASC, " + column + " " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
	}
	return column + " " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
}

type castSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type castSortQuery struct {
	ByName        castSortLink
	ByChinese     castSortLink
	ByAge         castSortLink
	ByHeight      castSortLink
	ByMovieNumber castSortLink
	ByWOwned      castSortLink
	ByScTimes     castSortLink
	ByComeTimes   castSortLink
	ByLastSc      castSortLink
	ByHighestRank castSortLink
}

func buildCastSortQuery(c *gin.Context, currentField, currentOrder string) *castSortQuery {
	makeHref := func(field string) castSortLink {
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

		return castSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &castSortQuery{
		ByName:        makeHref("name"),
		ByChinese:     makeHref("chinese"),
		ByAge:         makeHref("age"),
		ByHeight:      makeHref("height"),
		ByMovieNumber: makeHref("movie_number"),
		ByWOwned:      makeHref("owned_w_media_number"),
		ByScTimes:     makeHref("sc_times"),
		ByComeTimes:   makeHref("come_times"),
		ByLastSc:      makeHref("last_sc_time"),
		ByHighestRank: makeHref("highest_rank"),
	}
}

func normalizeCastSortFieldForPage(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "owned_movie_number":
		return "owned_w_media_number"
	case "name", "chinese", "age", "height", "movie_number", "owned_w_media_number", "sc_times", "come_times", "last_sc_time", "highest_rank":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "owned_w_media_number"
	}
}

func normalizeCastSortOrderForPage(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}

type itemListQuery struct {
	Page                    int64  `form:"p"`
	PageSize                int64  `form:"ps"`
	Sort                    string `form:"sort"`
	Order                   string `form:"order"`
	JavID                   string `form:"jav_id"`
	Name                    string `form:"name"`
	HasDetail               string `form:"has_detail"`
	HasDownloadCover        string `form:"has_download_cover"`
	HasChinese              string `form:"has_chinese"`
	DetailNeedScan          string `form:"detail_need_scan"`
	DetailBirthTimeFrom     string `form:"detail_birth_time_from"`
	DetailBirthTimeTo       string `form:"detail_birth_time_to"`
	LastQueryDetailTimeFrom string `form:"last_query_detail_time_from"`
	LastQueryDetailTimeTo   string `form:"last_query_detail_time_to"`
}

func (h *ListPagesAPI) EItemList(c *gin.Context) {
	defer withFrontendPath(c, "/eitemlist")()

	var q itemListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	q.Sort = normalizeItemSortField(q.Sort)
	q.Order = normalizeItemSortOrder(q.Order)

	filter, err := buildItemListFilter(q)
	if err != nil {
		jsonBadRequest(c, "筛选参数错误", err)
		return
	}

	total, err := h.deps.ItemModel.CountAll(c.Request.Context(), filter)
	if err != nil {
		jsonServerError(c, "e_item 列表加载失败", err)
		return
	}

	offset := (q.Page - 1) * q.PageSize
	items, err := h.deps.ItemModel.ListPage(c.Request.Context(), offset, q.PageSize, buildItemOrderByForPage(q.Sort, q.Order), filter)
	if err != nil {
		jsonServerError(c, "e_item 列表加载失败", err)
		return
	}
	rows := h.buildItemListRows(items)

	pi := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":         "E Items",
		"ViewType":      "e-items",
		"Items":         rows,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"Query":         q,
		"ItemSortQuery": buildItemSortQuery(c, q.Sort, q.Order),
	})
}

func (h *ListPagesAPI) buildItemListRows(items []*types.Item) []*types.ItemListRow {
	if len(items) == 0 {
		return []*types.ItemListRow{}
	}

	rows := make([]*types.ItemListRow, 0, len(items))
	for _, item := range items {
		row := buildItemListRow(item, h.deps.Config.Fetcher.JavAddress)
		if row != nil {
			rows = append(rows, row)
		}
	}
	return rows
}

func buildItemListRow(item *types.Item, javAddress string) *types.ItemListRow {
	if item == nil {
		return nil
	}

	hasDetailValue, hasDetailText := itemDetailStatusMeta(item.HasDetail)
	hasDownloadCoverValue, hasDownloadCoverText := itemCoverStatusMeta(item.HasDownloadCover)
	hasChineseValue, hasChineseText := itemChineseStatusMeta(item.HasChinese)
	detailNeedScanValue, detailNeedScanText := itemNeedScanStatusMeta(item.DetailNeedScan)

	return &types.ItemListRow{
		Id:                    item.Id,
		JavId:                 item.JavId,
		JavUrl:                buildItemJavURL(javAddress, item.JavId),
		Name:                  item.Name,
		HasDetail:             item.HasDetail,
		HasDetailValue:        hasDetailValue,
		HasDetailText:         hasDetailText,
		HasDownloadCover:      item.HasDownloadCover,
		HasDownloadCoverValue: hasDownloadCoverValue,
		HasDownloadCoverText:  hasDownloadCoverText,
		HasChinese:            item.HasChinese,
		HasChineseValue:       hasChineseValue,
		HasChineseText:        hasChineseText,
		DetailNeedScan:        item.DetailNeedScan,
		DetailNeedScanValue:   detailNeedScanValue,
		DetailNeedScanText:    detailNeedScanText,
		DetailBirthTime:       item.DetailBirthTime,
		LastQueryDetailTime:   item.LastQueryDetailTime,
	}
}

func buildItemJavURL(javAddress, javID string) string {
	if strings.TrimSpace(javAddress) == "" || strings.TrimSpace(javID) == "" {
		return ""
	}
	return "https://" + javAddress + "/cn/?v=" + javID
}

func itemDetailStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemDetailNone:
		return "none", "无详情"
	case consts.ItemDetailOK:
		return "ok", "已有详情"
	default:
		return "", "-"
	}
}

func itemCoverStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemCoverNone:
		return "none", "无本地封面"
	case consts.ItemCoverOK:
		return "ok", "已有本地封面"
	case consts.ItemCoverWrong:
		return "wrong", "错误封面链接"
	default:
		return "", "-"
	}
}

func itemChineseStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemChineseNone:
		return "none", "无中文字幕"
	case consts.ItemChineseOK:
		return "ok", "有中文字幕"
	case consts.ItemChineseError:
		return "error", "翻译错误"
	case consts.ItemChineseSensitive:
		return "sensitive", "含敏感词"
	default:
		return "", "-"
	}
}

func itemNeedScanStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemDetailStatusNeedScan:
		return "need", "需要扫描"
	case consts.ItemDetailStatusNoNeedScan:
		return "no_need", "无需扫描"
	case consts.ItemDetailStatusWrongContent:
		return "wrong", "内容错误"
	default:
		return "", "-"
	}
}

func buildItemListFilter(q itemListQuery) (types.ItemListFilter, error) {
	filter := types.ItemListFilter{
		JavID: strings.TrimSpace(q.JavID),
		Name:  strings.TrimSpace(q.Name),
	}

	if v, ok, err := parseItemDetailFilter(q.HasDetail); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.HasDetail = v
		filter.HasDetailSet = true
	}
	if v, ok, err := parseItemCoverFilter(q.HasDownloadCover); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.HasDownloadCover = v
		filter.HasDownloadCoverSet = true
	}
	if v, ok, err := parseItemChineseFilter(q.HasChinese); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.HasChinese = v
		filter.HasChineseSet = true
	}
	if v, ok, err := parseItemNeedScanFilter(q.DetailNeedScan); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.DetailNeedScan = v
		filter.DetailNeedScanSet = true
	}

	if ts, ok, err := parseOptionalDateStart(q.DetailBirthTimeFrom); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.DetailBirthTimeFrom = ts
		filter.HasDetailBirthTimeFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.DetailBirthTimeTo); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.DetailBirthTimeTo = ts
		filter.HasDetailBirthTimeTo = true
	}
	if ts, ok, err := parseOptionalDateStart(q.LastQueryDetailTimeFrom); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.LastQueryDetailTimeFrom = ts
		filter.HasLastQueryDetailTimeFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(q.LastQueryDetailTimeTo); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.LastQueryDetailTimeTo = ts
		filter.HasLastQueryDetailTimeTo = true
	}

	return filter, nil
}

func parseItemDetailFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "none":
		return consts.ItemDetailNone, true, nil
	case "ok":
		return consts.ItemDetailOK, true, nil
	default:
		return 0, false, fmt.Errorf("详情状态无效")
	}
}

func parseItemCoverFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "none":
		return consts.ItemCoverNone, true, nil
	case "ok":
		return consts.ItemCoverOK, true, nil
	case "wrong":
		return consts.ItemCoverWrong, true, nil
	default:
		return 0, false, fmt.Errorf("封面状态无效")
	}
}

func parseItemChineseFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "none":
		return consts.ItemChineseNone, true, nil
	case "ok":
		return consts.ItemChineseOK, true, nil
	case "error":
		return consts.ItemChineseError, true, nil
	case "sensitive":
		return consts.ItemChineseSensitive, true, nil
	default:
		return 0, false, fmt.Errorf("中文字幕状态无效")
	}
}

func parseItemNeedScanFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "need":
		return consts.ItemDetailStatusNeedScan, true, nil
	case "no_need":
		return consts.ItemDetailStatusNoNeedScan, true, nil
	case "wrong":
		return consts.ItemDetailStatusWrongContent, true, nil
	default:
		return 0, false, fmt.Errorf("详情扫描状态无效")
	}
}

type itemSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type itemSortQuery struct {
	ByJavID               itemSortLink
	ByName                itemSortLink
	ByHasDetail           itemSortLink
	ByHasDownloadCover    itemSortLink
	ByHasChinese          itemSortLink
	ByDetailNeedScan      itemSortLink
	ByDetailBirthTime     itemSortLink
	ByLastQueryDetailTime itemSortLink
}

func buildItemSortQuery(c *gin.Context, currentField, currentOrder string) *itemSortQuery {
	makeHref := func(field string) itemSortLink {
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

		return itemSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &itemSortQuery{
		ByJavID:               makeHref("jav_id"),
		ByName:                makeHref("name"),
		ByHasDetail:           makeHref("has_detail"),
		ByHasDownloadCover:    makeHref("has_download_cover"),
		ByHasChinese:          makeHref("has_chinese"),
		ByDetailNeedScan:      makeHref("detail_need_scan"),
		ByDetailBirthTime:     makeHref("detail_birth_time"),
		ByLastQueryDetailTime: makeHref("last_query_detail_time"),
	}
}

func normalizeItemSortField(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "jav_id", "name", "has_detail", "has_download_cover", "has_chinese", "detail_need_scan", "detail_birth_time", "last_query_detail_time":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "last_query_detail_time"
	}
}

func normalizeItemSortOrder(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}

func buildItemOrderByForPage(sortField, sortOrder string) string {
	field := normalizeItemSortField(sortField)
	order := "DESC"
	if normalizeItemSortOrder(sortOrder) == "asc" {
		order = "ASC"
	}

	column := "`last_query_detail_time`"
	switch field {
	case "jav_id":
		column = "`jav_id`"
	case "name":
		column = "`name`"
	case "has_detail":
		column = "`has_detail`"
	case "has_download_cover":
		column = "`has_download_cover`"
	case "has_chinese":
		column = "`has_chinese`"
	case "detail_need_scan":
		column = "`detail_need_scan`"
	case "detail_birth_time":
		column = "`detail_birth_time`"
	case "last_query_detail_time":
		column = "`last_query_detail_time`"
	}

	if column == "`jav_id`" || column == "`name`" {
		return column + " " + order + ", `id` DESC"
	}
	return column + " " + order + ", `id` DESC"
}

type crawlRecordListQuery struct {
	Type     string `form:"type"`
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

type crawlRecordListRow struct {
	Name            string
	Type            string
	DetailNumber    int64
	DurationMinText string
	StartedTime     int64
}

type crawlRecordTypeLink struct {
	Label  string
	Href   string
	Active bool
}

type crawlRecordSortLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

type crawlRecordHeaderLinks struct {
	StartedTime  crawlRecordSortLink
	Type         crawlRecordSortLink
	DetailNumber crawlRecordSortLink
	DurationMin  crawlRecordSortLink
	Name         crawlRecordSortLink
}

func (h *ListPagesAPI) CrawlRecordList(c *gin.Context) {
	defer withFrontendPath(c, "/crawlrecordlist")()

	var q crawlRecordListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	if q.Type == "" {
		q.Type = consts.RecordTypeSeedsActive
	}
	q.Sort = normalizeCrawlRecordSort(q.Sort)
	q.Order = normalizeCrawlRecordOrder(q.Order)

	rows, total, err := h.deps.RecordModel.ListPage(c.Request.Context(), moviex.ERecordListFilter{
		Type:     q.Type,
		Sort:     q.Sort,
		Order:    q.Order,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		jsonServerError(c, "query crawl records failed", err)
		return
	}

	pageInfo := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":       "CrawlRecords",
		"ViewType":    "crawl-records",
		"PageTitle":   "CrawlRecords 列表",
		"PageNote":    "直接查看抓取流程写入的 e_record 记录。",
		"Rows":        buildCrawlRecordListRows(rows),
		"Total":       total,
		"PageInfo":    pageInfo,
		"pageInfo":    pageInfo,
		"Query":       q,
		"TypeLinks":   buildCrawlRecordTypeLinks(c, q.Type),
		"HeaderLinks": buildCrawlRecordHeaderLinks(c, q.Sort, q.Order),
		"ClearHref":   "/crawlrecordlist",
	})
}

func buildCrawlRecordListRows(rows []*moviex.ERecord) []*crawlRecordListRow {
	if len(rows) == 0 {
		return []*crawlRecordListRow{}
	}

	out := make([]*crawlRecordListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &crawlRecordListRow{
			Name:            row.Name,
			Type:            row.Type,
			DetailNumber:    row.DetailNumber,
			DurationMinText: formatCrawlRecordDurationMinutes(row.StartTime, row.EndTime),
			StartedTime:     row.StartTime,
		})
	}
	return out
}

func formatCrawlRecordDurationMinutes(startTime int64, endTime int64) string {
	duration := endTime - startTime
	if duration <= 0 {
		return "-"
	}
	minutes := (duration + 59) / 60
	return fmt.Sprintf("%d", minutes)
}

func normalizeCrawlRecordSort(sortField string) string {
	switch sortField {
	case "type":
		return "type"
	case "detail_number":
		return "detail_number"
	case "duration":
		return "duration"
	case "name":
		return "name"
	default:
		return "start_time"
	}
}

func normalizeCrawlRecordOrder(sortOrder string) string {
	if sortOrder == "asc" {
		return "asc"
	}
	return "desc"
}

func buildCrawlRecordTypeLinks(c *gin.Context, currentType string) []crawlRecordTypeLink {
	options := []struct {
		Label string
		Value string
	}{
		{Label: "All", Value: ""},
		{Label: consts.RecordTypeDailyBest, Value: consts.RecordTypeDailyBest},
		{Label: consts.RecordTypeSeedsActive, Value: consts.RecordTypeSeedsActive},
		{Label: consts.RecordTypeSeedName, Value: consts.RecordTypeSeedName},
	}

	out := make([]crawlRecordTypeLink, 0, len(options))
	for _, option := range options {
		q := c.Request.URL.Query()
		q.Set("p", "1")
		if option.Value == "" {
			q.Del("type")
		} else {
			q.Set("type", option.Value)
		}

		href := c.Request.URL.Path
		if encoded := q.Encode(); encoded != "" {
			href += "?" + encoded
		}
		out = append(out, crawlRecordTypeLink{
			Label:  option.Label,
			Href:   href,
			Active: currentType == option.Value || (currentType == "" && option.Value == consts.RecordTypeSeedsActive),
		})
	}
	return out
}

func buildCrawlRecordHeaderLinks(c *gin.Context, currentSort string, currentOrder string) crawlRecordHeaderLinks {
	return crawlRecordHeaderLinks{
		StartedTime:  buildCrawlRecordSortLink(c, currentSort, currentOrder, "start_time", "StartedTime"),
		Type:         buildCrawlRecordSortLink(c, currentSort, currentOrder, "type", "Type"),
		DetailNumber: buildCrawlRecordSortLink(c, currentSort, currentOrder, "detail_number", "DetailNumber"),
		DurationMin:  buildCrawlRecordSortLink(c, currentSort, currentOrder, "duration", "Duration(min)"),
		Name:         buildCrawlRecordSortLink(c, currentSort, currentOrder, "name", "Name"),
	}
}

func buildCrawlRecordSortLink(c *gin.Context, currentSort string, currentOrder string, sortField string, label string) crawlRecordSortLink {
	nextOrder := "desc"
	active := currentSort == sortField
	if active && currentOrder == "desc" {
		nextOrder = "asc"
	}

	q := c.Request.URL.Query()
	q.Set("p", "1")
	q.Set("sort", sortField)
	q.Set("order", nextOrder)
	href := c.Request.URL.Path
	if encoded := q.Encode(); encoded != "" {
		href += "?" + encoded
	}

	return crawlRecordSortLink{
		Label:  label,
		Href:   href,
		Active: active,
		Desc:   active && currentOrder == "desc",
	}
}
