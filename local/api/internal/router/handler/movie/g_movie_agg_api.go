package movie

import (
	"net/http"
	"rudy_gc/local/api/internal/router/handler/common"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/moviereleaseagg"
)

type MovieAggAPI struct {
	releaseAggSvc *moviereleaseagg.Service
}

func NewMovieAggAPI(deps *dep.Dep) *MovieAggAPI {
	return &MovieAggAPI{
		releaseAggSvc: moviereleaseagg.NewService(deps),
	}
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
	Removed       movieReleaseBucketSortLink
	Size          movieReleaseBucketSortLink
	LatestRelease movieReleaseBucketSortLink
}

func (h *MovieAggAPI) MovieAggAllReleaseBucketList(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/bucketsrelease"
	defer func() { c.Request.URL.Path = originalPath }()

	aggMode := moviereleaseagg.NormalizeAggMode(c.Query("agg_mode"))
	page := common.AtoiDefault(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := common.ClampPageSize(common.AtoiDefault(c.DefaultQuery("ps", strconv.Itoa(common.DefaultAggPageSize)), common.DefaultAggPageSize))

	level := normalizeMovieReleaseBucketListLevel(c.Query("level"))
	scopeKey := strings.TrimSpace(c.Query("scope_key"))
	year, yearRaw := common.BucketListOptionalInt64(c, "year")
	quarter, quarterRaw := common.BucketListOptionalInt64(c, "quarter")
	month, monthRaw := common.BucketListOptionalInt64(c, "month")
	day, dayRaw := common.BucketListOptionalInt64(c, "day")
	curSort := normalizeMovieReleaseBucketListSort(c.Query("sort"))
	curDir := normalizeMovieReleaseBucketListDir(c.Query("dir"))

	out, err := h.releaseAggSvc.BuildBucketList(c.Request.Context(), moviereleaseagg.BucketListParams{
		AggMode:      aggMode,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载失败: " + err.Error()})
		return
	}

	pageTitle := "上映日时间桶落库列表"
	pageNote := "直接查看 movie_release_bucket_stat 原始落库结果，不做二次聚合。"
	clearHref := "/bucketsrelease"
	if aggMode == moviereleaseagg.AggModeOwned {
		pageTitle = "上映日时间桶落库列表（已拥有）"
		pageNote = "直接查看 movie_release_bucket_stat 的 owned 口径原始落库结果，只统计 native w_media。"
		clearHref = "/bucketsrelease?agg_mode=owned"
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":        pageTitle,
		"ViewType":     "movie-release-bucket-list",
		"PageTitle":    pageTitle,
		"PageNote":     pageNote,
		"AggMode":      aggMode,
		"AggModeLabel": map[string]string{moviereleaseagg.AggModeAll: "全量", moviereleaseagg.AggModeOwned: "已拥有"}[aggMode],
		"Rows":         out.Rows,
		"Total":        out.Total,
		"PageInfo":     common.BuildPageInfo(c, out.Total, int64(page), int64(size), common.PageWindow),
		"ModeSwitches": buildReleaseModeSwitches(c, aggMode),
		"SortLinks":    buildMovieReleaseBucketListSortLinks(c, curSort, curDir),
		"HeaderLinks":  buildMovieReleaseBucketListHeaderLinks(c, curSort, curDir),
		"LevelLinks":   buildMovieReleaseBucketListLevelLinks(c, level),
		"ClearHref":    clearHref,
		"ListPath":     clearHref,
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
	})
}

func buildReleaseModeSwitches(c *gin.Context, currentMode string) []movieReleaseBucketSortLink {
	items := []struct {
		Label string
		Mode  string
	}{
		{Label: "全量", Mode: moviereleaseagg.AggModeAll},
		{Label: "已拥有", Mode: moviereleaseagg.AggModeOwned},
	}
	out := make([]movieReleaseBucketSortLink, 0, len(items))
	for _, item := range items {
		q := common.CloneValues(c)
		q.Set("agg_mode", item.Mode)
		q.Set("p", "1")
		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		out = append(out, movieReleaseBucketSortLink{
			Label:  item.Label,
			Href:   href,
			Active: currentMode == item.Mode,
		})
	}
	return out
}

func normalizeMovieReleaseBucketListSort(v string) string {
	switch strings.TrimSpace(v) {
	case "":
		return "scope"
	case "updated":
		return "updated"
	case "all", "owned", "removed", "size", "latest_release", "scope":
		return strings.TrimSpace(v)
	default:
		return "scope"
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
		{Label: "按已删除", Href: "removed"},
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
		Removed:       makeLink("已删除", "removed"),
		Size:          makeLink("总大小", "size"),
		LatestRelease: makeLink("最新上映时间", "latest_release"),
	}
}

func buildMovieReleaseBucketListSortHref(c *gin.Context, sortValue string, currentSort string, currentDir string) string {
	q := common.CloneValues(c)
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
		q := common.CloneValues(c)
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
