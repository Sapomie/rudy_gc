package handler

import (
	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/moviereleaseagg"
	"rudy_gc/internal/types"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ------- All/Release 路由 -------

func (h *MovieAggHTMLHandler) MovieAggAllReleaseYears(c *gin.Context)   { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseMonths(c *gin.Context)  { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseQuarter(c *gin.Context) { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseMonth(c *gin.Context)   { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseDay(c *gin.Context)     { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseBucketList(c *gin.Context) {
	h.releaseBucketList(c)
}

func (h *MovieAggHTMLHandler) aggAllRelease(c *gin.Context) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultAggPageSize)), defaultAggPageSize))
	curOD := normalizeOrderBy(c.Query("od"), orderByRelease)
	sq := buildSortQuery(c, curOD)

	year, _ := strconv.Atoi(c.Param("year"))
	quarter, _ := strconv.Atoi(c.Param("q"))
	month, _ := strconv.Atoi(c.Param("month"))
	day, _ := strconv.Atoi(c.Param("day"))

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

	vm, err := h.releaseAggSvc.BuildReleaseView(c.Request.Context(), moviereleaseagg.AggParams{
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

	data := map[string]any{
		"Title":           vm.Title,
		"Breadcrumbs":     vm.Breadcrumbs,
		"Level":           vm.Level,
		"RangeStart":      vm.RangeStart,
		"RangeEnd":        vm.RangeEnd,
		"BucketsAll":      vm.BucketsAll,
		"BucketsQAll":     vm.BucketsQAll,
		"BucketsMAll":     vm.BucketsMAll,
		"TopCastsAll":     vm.TopCastsAll,
		"TopDirectorsAll": vm.TopDirectorsAll,
		"TopLabelsAll":    vm.TopLabelsAll,
		"TopPrefixesAll":  vm.TopPrefixesAll,
		"Mode":            "release_all",
		"CurrentSort":     curOD,
		"SortQuery":       sq,
	}

	if vm.Level != "root" {
		listReq := &types.ListMovieFullRequest{
			Owned:              consts.MovieAll,
			OrderBy:            curOD,
			Page:               int64(page),
			PageSize:           int64(size),
			ReleasingDateStart: vm.RangeStart,
			ReleasingDateEnd:   vm.RangeEnd,
		}
		listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
		if err != nil {
			c.String(500, "加载失败: %v", err)
			return
		}
		data["Movies"] = listResp.List
		data["Total"] = listResp.Total
		data["PageInfo"] = BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	}

	c.HTML(200, "page.movie_agg_all_time", data)
}

type movieReleaseBucketListQuery struct {
	Level    string
	ScopeKey string
	Year     string
	Quarter  string
	Month    string
	Day      string
	Sort     string
	Dir      string
	Page     int64
	PageSize int64
}

type movieReleaseBucketSortLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

type movieReleaseBucketHeaderLinks struct {
	Bucket        movieReleaseBucketSortLink
	CountAll      movieReleaseBucketSortLink
	CountOwned    movieReleaseBucketSortLink
	Size          movieReleaseBucketSortLink
	LatestRelease movieReleaseBucketSortLink
}

func (h *MovieAggHTMLHandler) releaseBucketList(c *gin.Context) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultAggPageSize)), defaultAggPageSize))

	level := normalizeMovieReleaseBucketListLevel(c.Query("level"))
	scopeKey := strings.TrimSpace(c.Query("scope_key"))
	year, yearRaw := bucketListOptionalInt64(c, "year")
	quarter, quarterRaw := bucketListOptionalInt64(c, "quarter")
	month, monthRaw := bucketListOptionalInt64(c, "month")
	day, dayRaw := bucketListOptionalInt64(c, "day")
	curSort := normalizeMovieReleaseBucketListSort(c.Query("sort"))
	curDir := normalizeMovieReleaseBucketListDir(c.Query("dir"))

	out, err := h.releaseAggSvc.BuildBucketList(c.Request.Context(), moviereleaseagg.BucketListParams{
		Level:        level,
		ScopeKeyLike: scopeKey,
		Year:         year,
		Quarter:      quarter,
		Month:        month,
		Day:          day,
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
		"Title":       "上映日时间桶列表",
		"PageTitle":   "上映日时间桶落库列表",
		"PageNote":    "直接查看 movie_release_bucket_stat 原始落库结果，不做二次聚合。",
		"Rows":        out.Rows,
		"Total":       out.Total,
		"PageInfo":    BuildPageInfo(c, out.Total, int64(page), int64(size), pageWindow),
		"SortLinks":   buildMovieReleaseBucketListSortLinks(c, curSort, curDir),
		"HeaderLinks": buildMovieReleaseBucketListHeaderLinks(c, curSort, curDir),
		"LevelLinks":  buildMovieReleaseBucketListLevelLinks(c, level),
		"ClearHref":   "/movie-agg-all/release-bucket-list",
		"AggPageHref": "/movie-agg-all/release",
		"Query": movieReleaseBucketListQuery{
			Level:    level,
			ScopeKey: scopeKey,
			Year:     yearRaw,
			Quarter:  quarterRaw,
			Month:    monthRaw,
			Day:      dayRaw,
			Sort:     curSort,
			Dir:      curDir,
			Page:     int64(page),
			PageSize: int64(size),
		},
	}

	c.HTML(200, "page.movie_release_bucket_list", data)
}

func normalizeMovieReleaseBucketListSort(v string) string {
	switch strings.TrimSpace(v) {
	case "", "updated":
		return "updated"
	case "all", "owned", "size", "latest_release", "scope":
		return strings.TrimSpace(v)
	default:
		return "updated"
	}
}

func normalizeMovieReleaseBucketListDir(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}

func normalizeMovieReleaseBucketListLevel(v string) string {
	switch strings.TrimSpace(v) {
	case "year", "quarter", "month", "day":
		return strings.TrimSpace(v)
	default:
		return "month"
	}
}

func buildMovieReleaseBucketListSortLinks(c *gin.Context, currentSort string, currentDir string) []movieReleaseBucketSortLink {
	items := []movieReleaseBucketSortLink{
		{Label: "按更新时间", Href: "updated"},
		{Label: "按总数", Href: "all"},
		{Label: "按已拥有", Href: "owned"},
		{Label: "按总大小", Href: "size"},
		{Label: "按最新上映时间", Href: "latest_release"},
		{Label: "按时间桶", Href: "scope"},
	}
	out := make([]movieReleaseBucketSortLink, 0, len(items))
	for _, item := range items {
		out = append(out, movieReleaseBucketSortLink{
			Label:  item.Label,
			Href:   buildMovieReleaseBucketListSortHref(c, item.Href, currentSort, currentDir),
			Active: currentSort == item.Href,
			Desc:   currentSort == item.Href && currentDir == "desc",
		})
	}
	return out
}

func buildMovieReleaseBucketListHeaderLinks(c *gin.Context, currentSort string, currentDir string) movieReleaseBucketHeaderLinks {
	makeLink := func(label, sortValue string) movieReleaseBucketSortLink {
		return movieReleaseBucketSortLink{
			Label:  label,
			Href:   buildMovieReleaseBucketListSortHref(c, sortValue, currentSort, currentDir),
			Active: currentSort == sortValue,
			Desc:   currentSort == sortValue && currentDir == "desc",
		}
	}

	return movieReleaseBucketHeaderLinks{
		Bucket:        makeLink("时间桶", "scope"),
		CountAll:      makeLink("总数", "all"),
		CountOwned:    makeLink("已拥有", "owned"),
		Size:          makeLink("总大小", "size"),
		LatestRelease: makeLink("最新上映时间", "latest_release"),
	}
}

func buildMovieReleaseBucketListSortHref(c *gin.Context, sortValue string, currentSort string, currentDir string) string {
	q := cloneValues(c)
	q.Set("sort", sortValue)
	q.Set("dir", nextMovieReleaseBucketListDir(sortValue, currentSort, currentDir))
	q.Set("p", "1")
	href := c.Request.URL.Path
	if enc := q.Encode(); enc != "" {
		href += "?" + enc
	}
	return href
}

func nextMovieReleaseBucketListDir(sortValue string, currentSort string, currentDir string) string {
	if currentSort == sortValue && currentDir == "desc" {
		return "asc"
	}
	return "desc"
}

func buildMovieReleaseBucketListLevelLinks(c *gin.Context, current string) []movieReleaseBucketSortLink {
	items := []struct {
		Label string
		Value string
	}{
		{Label: "Year", Value: "year"},
		{Label: "Quarter", Value: "quarter"},
		{Label: "Month", Value: "month"},
		{Label: "Day", Value: "day"},
	}
	out := make([]movieReleaseBucketSortLink, 0, len(items))
	for _, item := range items {
		q := cloneValues(c)
		q.Set("level", item.Value)
		q.Set("p", "1")
		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		out = append(out, movieReleaseBucketSortLink{
			Label:  item.Label,
			Href:   href,
			Active: current == item.Value,
		})
	}
	return out
}
