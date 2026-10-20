package list

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/router/handler/common"
	"rudy_gc/local/api/internal/types"
)

type mediaListQuery struct {
	Page               int64  `form:"p"`
	PageSize           int64  `form:"ps"`
	Sort               string `form:"sort"`
	Order              string `form:"order"`
	Keyword            string `form:"keyword"`
	SizeMin            string `form:"size_min"`
	SizeMax            string `form:"size_max"`
	HeightMin          string `form:"height_min"`
	HeightMax          string `form:"height_max"`
	DurationMin        string `form:"duration_min"`
	DurationMax        string `form:"duration_max"`
	BitRateMin         string `form:"bit_rate_min"`
	BitRateMax         string `form:"bit_rate_max"`
	FrameAverageMin    string `form:"frame_average_min"`
	FrameAverageMax    string `form:"frame_average_max"`
	SelfMake           string `form:"self_make"`
	HasMask            string `form:"has_mask"`
	ScTimesMin         string `form:"sc_times_min"`
	ScTimesMax         string `form:"sc_times_max"`
	LastSelectedScFrom string `form:"last_selected_sc_from"`
	LastSelectedScTo   string `form:"last_selected_sc_to"`
	BirthFrom          string `form:"birth_from"`
	BirthTo            string `form:"birth_to"`
	ReleasingFrom      string `form:"releasing_from"`
	ReleasingTo        string `form:"releasing_to"`
}

func (h *ListPagesAPI) MediaList(c *gin.Context) {
	defer common.WithFrontendPath(c, "/medialist")()

	var q mediaListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
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
		common.JSONBadRequest(c, "筛选参数错误", err)
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
		common.JSONServerError(c, "媒体列表加载失败", err)
		return
	}

	pi := common.BuildPageInfo(c, total, q.Page, q.PageSize, common.PageWindow)
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

	if v, ok, err := common.ParseOptionalNonNegativeFloat64(q.SizeMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SizeMin = common.GBToBytes(v)
		filter.HasSizeMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeFloat64(q.SizeMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SizeMax = common.GBToBytes(v)
		filter.HasSizeMax = true
	}

	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.HeightMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HeightMin = v
		filter.HasHeightMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.HeightMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HeightMax = v
		filter.HasHeightMax = true
	}

	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.DurationMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.DurationMin = common.MinutesToSeconds(v)
		filter.HasDurationMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.DurationMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.DurationMax = common.MinutesToSeconds(v)
		filter.HasDurationMax = true
	}

	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.BitRateMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BitRateMin = v
		filter.HasBitRateMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.BitRateMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BitRateMax = v
		filter.HasBitRateMax = true
	}

	if v, ok, err := common.ParseOptionalNonNegativeFloat64(q.FrameAverageMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.FrameAverageMin = v
		filter.HasFrameAverageMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeFloat64(q.FrameAverageMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.FrameAverageMax = v
		filter.HasFrameAverageMax = true
	}

	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.SelfMake); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SelfMake = v
		filter.HasSelfMake = true
	}

	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.HasMask); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HasMask = v
		filter.HasHasMask = true
	}

	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.ScTimesMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ScTimesMin = v
		filter.HasScTimesMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.ScTimesMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ScTimesMax = v
		filter.HasScTimesMax = true
	}

	if v, ok, err := common.ParseOptionalDateUnixStart(q.LastSelectedScFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.LastSelectedScFrom = v
		filter.HasLastSelectedScFrom = true
	}
	if v, ok, err := common.ParseOptionalDateUnixEnd(q.LastSelectedScTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.LastSelectedScTo = v
		filter.HasLastSelectedScTo = true
	}

	if v, ok, err := common.ParseOptionalDateUnixStart(q.BirthFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BirthTimeFrom = v
		filter.HasBirthTimeFrom = true
	}
	if v, ok, err := common.ParseOptionalDateUnixEnd(q.BirthTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BirthTimeTo = v
		filter.HasBirthTimeTo = true
	}

	if v, ok, err := common.ParseOptionalDateUnixStart(q.ReleasingFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ReleasingDateFrom = v
		filter.HasReleasingDateFrom = true
	}
	if v, ok, err := common.ParseOptionalDateUnixEnd(q.ReleasingTo); err != nil {
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
		return "COALESCE(gss.sc_times, 0) " + order + ", COALESCE(gss.last_selected_sc_time, 0) DESC, wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "last_selected_sc_time":
		return "COALESCE(gss.last_selected_sc_time, 0) " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
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
	ByMovieName      filmSortLink
	BySize           filmSortLink
	ByHeight         filmSortLink
	ByDuration       filmSortLink
	ByBitRate        filmSortLink
	ByFrameAverage   filmSortLink
	BySelfMake       filmSortLink
	ByHasMask        filmSortLink
	ByScTimes        filmSortLink
	ByLastSelectedSc filmSortLink
	ByBirthTime      filmSortLink
	ByReleasingDate  filmSortLink
}

func buildFilmSortQuery(c *gin.Context, currentField, currentOrder string) *filmSortQuery {
	makeHref := func(field string) filmSortLink {
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

		return filmSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &filmSortQuery{
		ByMovieName:      makeHref("movie_name"),
		BySize:           makeHref("size"),
		ByHeight:         makeHref("height"),
		ByDuration:       makeHref("duration"),
		ByBitRate:        makeHref("bit_rate"),
		ByFrameAverage:   makeHref("frame_average"),
		BySelfMake:       makeHref("self_make"),
		ByHasMask:        makeHref("has_mask"),
		ByScTimes:        makeHref("sc_times"),
		ByLastSelectedSc: makeHref("last_selected_sc_time"),
		ByBirthTime:      makeHref("birth_time"),
		ByReleasingDate:  makeHref("releasing_date"),
	}
}

func normalizeFilmSortFieldForPage(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "movie_name", "size", "height", "duration", "bit_rate", "frame_average", "self_make", "has_mask", "sc_times", "last_selected_sc_time", "birth_time", "releasing_date":
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
	defer common.WithFrontendPath(c, "/aggeventlist")()

	var q wAggEventListQuery
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
		common.JSONServerError(c, "w_agg_event 列表加载失败", err)
		return
	}

	pageInfo := common.BuildPageInfo(c, total, q.Page, q.PageSize, common.PageWindow)
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
