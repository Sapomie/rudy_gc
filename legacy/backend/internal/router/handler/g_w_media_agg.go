package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/wmediaagg"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type WMediaAggHTMLHandler struct {
	deps         *svc.Deps
	movieSvc     *movie.Service
	wMediaAggSvc *wmediaagg.Service
}

func NewWMediaAggHTMLHandler(deps *svc.Deps) *WMediaAggHTMLHandler {
	return &WMediaAggHTMLHandler{
		deps:         deps,
		movieSvc:     movie.NewService(deps),
		wMediaAggSvc: wmediaagg.NewService(deps),
	}
}

func (h *WMediaAggHTMLHandler) BirthYears(c *gin.Context)   { h.birthCommon(c) }
func (h *WMediaAggHTMLHandler) BirthYear(c *gin.Context)    { h.birthCommon(c) }
func (h *WMediaAggHTMLHandler) BirthQuarter(c *gin.Context) { h.birthCommon(c) }
func (h *WMediaAggHTMLHandler) BirthMonth(c *gin.Context)   { h.birthCommon(c) }
func (h *WMediaAggHTMLHandler) BirthDay(c *gin.Context)     { h.birthCommon(c) }
func (h *WMediaAggHTMLHandler) BucketList(c *gin.Context)   { h.bucketList(c) }

func (h *WMediaAggHTMLHandler) birthCommon(c *gin.Context) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultAggPageSize)), defaultAggPageSize))
	curOD := normalizeOrderBy(c.Query("od"), consts.OrderByMediaBirthTime)
	sq := buildSortQuery(c, curOD)

	year := atoiDef(c.Query("year"), 0)
	quarter := atoiDef(c.Query("quarter"), 0)
	month := atoiDef(c.Query("month"), 0)
	day := atoiDef(c.Query("day"), 0)

	topN := 30
	if v := c.Query("tn"); v != "" {
		topN = atoiDef(v, 30)
	} else if v := c.Query("top"); v != "" {
		topN = atoiDef(v, 30)
	} else if v := c.Query("topn"); v != "" {
		topN = atoiDef(v, 30)
	} else if v := c.Query("tc"); v != "" {
		topN = atoiDef(v, 30)
	}
	if topN < 1 {
		topN = 1
	}
	if topN > 200 {
		topN = 200
	}

	vm, err := h.wMediaAggSvc.BuildBirthView(c.Request.Context(), wmediaagg.AggParams{
		Year:    year,
		Quarter: quarter,
		Month:   month,
		Day:     day,
		TopN:    topN,
	})
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	baseReq := types.ListMovieFullRequest{
		MediaOwned: consts.OwnedAllNotRemoved,
		OrderBy:    curOD,
		Page:       int64(page),
		PageSize:   int64(size),
	}
	listReq, err := parseMovieCardRequest(c, baseReq)
	if err != nil {
		c.String(400, "参数解析错误: %v", err)
		return
	}
	listReq.MediaBirthTimeStart = vm.RangeStart
	listReq.MediaBirthTimeEnd = vm.RangeEnd

	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), &listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pageInfo := BuildPageInfo(c, listResp.Total, listReq.Page, listReq.PageSize, pageWindow)

	data := map[string]any{
		"Title":        vm.Title,
		"Breadcrumbs":  vm.Breadcrumbs,
		"Level":        vm.Level,
		"RangeStart":   vm.RangeStart,
		"RangeEnd":     vm.RangeEnd,
		"BucketsY":     vm.BucketsY,
		"BucketsQ":     vm.BucketsQ,
		"BucketsM":     vm.BucketsM,
		"BucketsD":     vm.BucketsD,
		"TopCasts":     vm.TopCasts,
		"TopDirectors": vm.TopDirectors,
		"TopLabels":    vm.TopLabels,
		"TopPrefixes":  vm.TopPrefixes,
		"Movies":       listResp.List,
		"movies":       listResp.List,
		"Total":        listResp.Total,
		"total":        listResp.Total,
		"PageInfo":     pageInfo,
		"pageInfo":     pageInfo,
		"SortQuery":    sq,
		"sortQuery":    sq,
		"CurrentSort":  curOD,
		"ownedQuery":   buildOwnedFilterInfoWithDefaults(c, "3"),
	}

	c.HTML(200, "page.w_media_agg_birth_time", data)
}

type wMediaBirthBucketListQuery struct {
	Level    string
	ScopeKey string
	Year     string
	Quarter  string
	Month    string
	Sort     string
	Dir      string
	Page     int64
	PageSize int64
}

type wMediaBirthBucketSortLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

type wMediaBirthBucketHeaderLinks struct {
	Bucket      wMediaBirthBucketSortLink
	Media       wMediaBirthBucketSortLink
	Removed     wMediaBirthBucketSortLink
	Size        wMediaBirthBucketSortLink
	Subtitle    wMediaBirthBucketSortLink
	LatestBirth wMediaBirthBucketSortLink
}

func (h *WMediaAggHTMLHandler) bucketList(c *gin.Context) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultAggPageSize)), defaultAggPageSize))

	level := normalizeBucketListLevel(c.Query("level"))
	scopeKey := strings.TrimSpace(c.Query("scope_key"))
	year, yearRaw := bucketListOptionalInt64(c, "year")
	quarter, quarterRaw := bucketListOptionalInt64(c, "quarter")
	month, monthRaw := bucketListOptionalInt64(c, "month")
	curSort := normalizeBucketListSort(c.Query("sort"))
	curDir := normalizeBucketListDir(c.Query("dir"))

	out, err := h.wMediaAggSvc.BuildBucketList(c.Request.Context(), wmediaagg.BucketListParams{
		Level:        level,
		ScopeKeyLike: scopeKey,
		Year:         year,
		Quarter:      quarter,
		Month:        month,
		Sort:         curSort,
		Dir:          curDir,
		Page:         int64(page),
		PageSize:     int64(size),
	})
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	data := map[string]any{
		"Title":       "WMedia 时间桶列表",
		"PageTitle":   "WMedia 时间桶落库列表",
		"PageNote":    "直接查看 w_media_birth_bucket_stat 原始落库结果，不做二次聚合。",
		"Rows":        out.Rows,
		"Total":       out.Total,
		"PageInfo":    BuildPageInfo(c, out.Total, int64(page), int64(size), pageWindow),
		"SortLinks":   buildBucketListSortLinks(c, curSort, curDir),
		"HeaderLinks": buildBucketListHeaderLinks(c, curSort, curDir),
		"LevelLinks":  buildBucketListLevelLinks(c, level),
		"ClearHref":   "/w-media-agg/bucket-list",
		"AggPageHref": "/w-media-agg/birth",
		"Query": wMediaBirthBucketListQuery{
			Level:    level,
			ScopeKey: scopeKey,
			Year:     yearRaw,
			Quarter:  quarterRaw,
			Month:    monthRaw,
			Sort:     curSort,
			Dir:      curDir,
			Page:     int64(page),
			PageSize: int64(size),
		},
	}

	c.HTML(200, "page.w_media_birth_bucket_list", data)
}

func bucketListOptionalInt64(c *gin.Context, key string) (*int64, string) {
	raw, ok := c.GetQuery(key)
	if !ok {
		return nil, ""
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ""
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, raw
	}
	return &v, raw
}

func normalizeBucketListSort(v string) string {
	switch strings.TrimSpace(v) {
	case "":
		return "scope"
	case "updated":
		return "updated"
	case "media", "removed", "size", "subtitle", "latest_birth", "scope":
		return strings.TrimSpace(v)
	default:
		return "scope"
	}
}

func normalizeBucketListDir(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}

func normalizeBucketListLevel(v string) string {
	switch strings.TrimSpace(v) {
	case "year", "quarter", "month", "day":
		return strings.TrimSpace(v)
	default:
		return "month"
	}
}

func buildBucketListSortLinks(c *gin.Context, currentSort string, currentDir string) []wMediaBirthBucketSortLink {
	items := []wMediaBirthBucketSortLink{
		{Label: "按更新时间", Href: "updated"},
		{Label: "按媒体数", Href: "media"},
		{Label: "按已删除数", Href: "removed"},
		{Label: "按总大小", Href: "size"},
		{Label: "按字幕数", Href: "subtitle"},
		{Label: "按最新下载时间", Href: "latest_birth"},
		{Label: "按时间桶", Href: "scope"},
	}
	out := make([]wMediaBirthBucketSortLink, 0, len(items))
	for _, item := range items {
		out = append(out, wMediaBirthBucketSortLink{
			Label:  item.Label,
			Href:   buildBucketListSortHref(c, item.Href, currentSort, currentDir),
			Active: currentSort == item.Href,
			Desc:   currentSort == item.Href && currentDir == "desc",
		})
	}
	return out
}

func buildBucketListHeaderLinks(c *gin.Context, currentSort string, currentDir string) wMediaBirthBucketHeaderLinks {
	makeLink := func(label, sortValue string) wMediaBirthBucketSortLink {
		return wMediaBirthBucketSortLink{
			Label:  label,
			Href:   buildBucketListSortHref(c, sortValue, currentSort, currentDir),
			Active: currentSort == sortValue,
			Desc:   currentSort == sortValue && currentDir == "desc",
		}
	}

	return wMediaBirthBucketHeaderLinks{
		Bucket:      makeLink("时间桶", "scope"),
		Media:       makeLink("媒体数", "media"),
		Removed:     makeLink("已删除数", "removed"),
		Size:        makeLink("总大小", "size"),
		Subtitle:    makeLink("字幕数", "subtitle"),
		LatestBirth: makeLink("最新下载时间", "latest_birth"),
	}
}

func buildBucketListSortHref(c *gin.Context, sortValue string, currentSort string, currentDir string) string {
	q := cloneValues(c)
	q.Set("sort", sortValue)
	q.Set("dir", nextBucketListDir(sortValue, currentSort, currentDir))
	q.Set("p", "1")
	href := c.Request.URL.Path
	if enc := q.Encode(); enc != "" {
		href += "?" + enc
	}
	return href
}

func nextBucketListDir(sortValue string, currentSort string, currentDir string) string {
	if currentSort == sortValue && currentDir == "desc" {
		return "asc"
	}
	return "desc"
}

func buildBucketListLevelLinks(c *gin.Context, current string) []wMediaBirthBucketSortLink {
	items := []struct {
		Label string
		Value string
	}{
		{Label: "Year", Value: "year"},
		{Label: "Quarter", Value: "quarter"},
		{Label: "Month", Value: "month"},
		{Label: "Day", Value: "day"},
	}

	out := make([]wMediaBirthBucketSortLink, 0, len(items))
	for _, item := range items {
		q := cloneValues(c)
		q.Set("level", item.Value)
		q.Set("p", "1")
		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		out = append(out, wMediaBirthBucketSortLink{
			Label:  item.Label,
			Href:   href,
			Active: current == item.Value,
		})
	}
	return out
}
