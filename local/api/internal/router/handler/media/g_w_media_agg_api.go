package media

import (
	"net/http"
	"rudy_gc/local/api/internal/router/handler/common"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/wmediaagg"
)

type WMediaAggAPI struct {
	wMediaAggSvc *wmediaagg.Service
}

func NewWMediaAggAPI(deps *dep.Dep) *WMediaAggAPI {
	return &WMediaAggAPI{
		wMediaAggSvc: wmediaagg.NewService(deps),
	}
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
	Added       wMediaBirthBucketSortLink
	AddedSize   wMediaBirthBucketSortLink
	RemovedSize wMediaBirthBucketSortLink
	NetSize     wMediaBirthBucketSortLink
	Subtitle    wMediaBirthBucketSortLink
	LatestBirth wMediaBirthBucketSortLink
}

func (h *WMediaAggAPI) BucketList(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/bucketsbirth"
	defer func() { c.Request.URL.Path = originalPath }()

	page := common.AtoiDefault(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := common.ClampPageSize(common.AtoiDefault(c.DefaultQuery("ps", strconv.Itoa(common.DefaultAggPageSize)), common.DefaultAggPageSize))

	level := normalizeBucketListLevel(c.Query("level"))
	scopeKey := strings.TrimSpace(c.Query("scope_key"))
	year, yearRaw := common.BucketListOptionalInt64(c, "year")
	quarter, quarterRaw := common.BucketListOptionalInt64(c, "quarter")
	month, monthRaw := common.BucketListOptionalInt64(c, "month")
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":       "WMedia 时间桶列表",
		"ViewType":    "w-media-bucket-list",
		"PageTitle":   "WMedia 时间桶落库列表",
		"PageNote":    "新增按下载时间统计，删除按删除时间统计；净增加大小 = 新增大小 − 删除大小。",
		"Rows":        out.Rows,
		"Total":       out.Total,
		"PageInfo":    common.BuildPageInfo(c, out.Total, int64(page), int64(size), common.PageWindow),
		"SortLinks":   buildBucketListSortLinks(c, curSort, curDir),
		"HeaderLinks": buildBucketListHeaderLinks(c, curSort, curDir),
		"LevelLinks":  buildBucketListLevelLinks(c, level),
		"ClearHref":   "/bucketsbirth",
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
	})
}

func normalizeBucketListSort(v string) string {
	switch strings.TrimSpace(v) {
	case "":
		return "scope"
	case "updated":
		return "updated"
	case "added", "added_size", "removed", "removed_size", "net_size", "media", "size", "subtitle", "latest_birth", "scope":
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
		{Label: "按新增数目", Href: "added"},
		{Label: "按新增大小", Href: "added_size"},
		{Label: "按删除数目", Href: "removed"},
		{Label: "按删除大小", Href: "removed_size"},
		{Label: "按净增加大小", Href: "net_size"},
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
		Removed:     makeLink("删除数目", "removed"),
		Size:        makeLink("总大小", "size"),
		Added:       makeLink("新增数目", "added"),
		AddedSize:   makeLink("新增大小", "added_size"),
		RemovedSize: makeLink("删除大小", "removed_size"),
		NetSize:     makeLink("净增加大小", "net_size"),
		Subtitle:    makeLink("字幕数", "subtitle"),
		LatestBirth: makeLink("最新下载时间", "latest_birth"),
	}
}

func buildBucketListSortHref(c *gin.Context, sortValue string, currentSort string, currentDir string) string {
	q := common.CloneValues(c)
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
		q := common.CloneValues(c)
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
