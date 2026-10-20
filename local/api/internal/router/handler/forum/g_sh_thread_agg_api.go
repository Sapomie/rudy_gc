package forum

import (
	"fmt"
	"net/http"
	"rudy_gc/local/api/internal/router/handler/common"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/shthreadagg"
)

type SHThreadAggAPI struct {
	service *shthreadagg.Service
}

type shBucketListQuery struct {
	AggScope string
	Level    string
	TimeFrom string
	TimeTo   string
	Sort     string
	Dir      string
	Page     int64
	PageSize int64
}

type shBucketLink struct {
	Label  string
	Href   string
	Active bool
	Desc   bool
}

type shBucketHeaderLinks struct {
	Bucket   shBucketLink
	All      shBucketLink
	Pending  shBucketLink
	Fetched  shBucketLink
	Failed   shBucketLink
	Magnet   shBucketLink
	InfoHash shBucketLink
	Latest   shBucketLink
}

func NewSHThreadAggAPI(deps *dep.Dep) *SHThreadAggAPI {
	return &SHThreadAggAPI{service: shthreadagg.NewService(deps)}
}

func (h *SHThreadAggAPI) BucketList(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/shbucketlist"
	defer func() { c.Request.URL.Path = originalPath }()
	removeLegacySHBucketTimeQuery(c)

	page := common.AtoiDefault(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := common.ClampPageSize(common.AtoiDefault(c.DefaultQuery("ps", strconv.Itoa(common.DefaultAggPageSize)), common.DefaultAggPageSize))
	aggScope := normalizeSHBucketAggScope(c.Query("agg_scope"))
	level := normalizeSHBucketLevel(c.Query("level"))
	timeFrom, timeFromRaw, err := parseSHBucketDate(c.Query("time_from"), "起始时间", false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	timeTo, timeToRaw, err := parseSHBucketDate(c.Query("time_to"), "结束时间", true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if timeFrom != nil && timeTo != nil && *timeFrom > *timeTo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "起始时间不能晚于结束时间"})
		return
	}
	sortValue := normalizeSHBucketSort(c.Query("sort"))
	dir := normalizeSHBucketDir(c.Query("dir"))

	result, err := h.service.BuildBucketList(c.Request.Context(), shthreadagg.BucketListParams{
		AggScope: aggScope, Level: level, TimeFrom: timeFrom, TimeTo: timeTo,
		Sort: sortValue, Dir: dir, Page: int64(page), PageSize: int64(size),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":       "SH 帖子时间桶列表",
		"ViewType":    "sh-thread-bucket-list",
		"PageTitle":   "SH 帖子时间桶落库列表",
		"PageNote":    "直接查看 sh_bucket_stat 原始落库结果，点击时间桶可进入帖子池查看对应帖子。",
		"Rows":        result.Rows,
		"Total":       result.Total,
		"PageInfo":    common.BuildPageInfo(c, result.Total, int64(page), int64(size), common.PageWindow),
		"HeaderLinks": buildSHBucketHeaderLinks(c, sortValue, dir),
		"LevelLinks":  buildSHBucketLevelLinks(c, level),
		"ClearHref":   "/shbucketlist",
		"ListPath":    "/shbucketlist",
		"Query": shBucketListQuery{
			AggScope: aggScope, Level: level, TimeFrom: timeFromRaw, TimeTo: timeToRaw,
			Sort: sortValue, Dir: dir, Page: int64(page), PageSize: int64(size),
		},
	})
}

func removeLegacySHBucketTimeQuery(c *gin.Context) {
	q := c.Request.URL.Query()
	for _, key := range []string{"scope_key", "year", "quarter", "month", "week", "day"} {
		q.Del(key)
	}
	c.Request.URL.RawQuery = q.Encode()
}

func parseSHBucketDate(value, label string, endOfDay bool) (*int64, string, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, "", nil
	}
	parsed, err := time.ParseInLocation(time.DateOnly, raw, time.Local)
	if err != nil {
		return nil, raw, fmt.Errorf("%s格式无效，请使用 YYYY-MM-DD", label)
	}
	if endOfDay {
		parsed = parsed.AddDate(0, 0, 1).Add(-time.Second)
	}
	unix := parsed.Unix()
	return &unix, raw, nil
}

func normalizeSHBucketAggScope(value string) string {
	switch strings.TrimSpace(value) {
	case "forum:36", "forum:103":
		return strings.TrimSpace(value)
	default:
		return shthreadagg.AggScopeAll
	}
}

func normalizeSHBucketLevel(value string) string {
	switch strings.TrimSpace(value) {
	case "year", "quarter", "month", "week", "day":
		return strings.TrimSpace(value)
	default:
		return "month"
	}
}

func normalizeSHBucketSort(value string) string {
	switch strings.TrimSpace(value) {
	case "updated", "all", "pending", "fetched", "failed", "magnet", "info_hash", "latest_post", "scope":
		return strings.TrimSpace(value)
	default:
		return "scope"
	}
}

func normalizeSHBucketDir(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "asc"
	}
	return "desc"
}

func buildSHBucketHeaderLinks(c *gin.Context, currentSort, currentDir string) shBucketHeaderLinks {
	link := func(label, value string) shBucketLink {
		return shBucketLink{
			Label: label, Href: buildSHBucketSortHref(c, value, currentSort, currentDir),
			Active: currentSort == value, Desc: currentSort == value && currentDir == "desc",
		}
	}
	return shBucketHeaderLinks{
		Bucket: link("时间桶", "scope"), All: link("帖子总数", "all"), Pending: link("待抓取", "pending"),
		Fetched: link("已抓取", "fetched"), Failed: link("失败", "failed"), Magnet: link("包含磁力", "magnet"),
		InfoHash: link("包含 InfoHash", "info_hash"), Latest: link("最新发帖", "latest_post"),
	}
}

func buildSHBucketSortHref(c *gin.Context, sortValue, currentSort, currentDir string) string {
	q := common.CloneValues(c)
	q.Set("sort", sortValue)
	q.Set("dir", nextSHBucketDir(sortValue, currentSort, currentDir))
	q.Set("p", "1")
	href := c.Request.URL.Path
	if encoded := q.Encode(); encoded != "" {
		href += "?" + encoded
	}
	return href
}

func nextSHBucketDir(sortValue, currentSort, currentDir string) string {
	if currentSort == sortValue && currentDir == "desc" {
		return "asc"
	}
	return "desc"
}

func buildSHBucketLevelLinks(c *gin.Context, current string) []shBucketLink {
	items := []struct{ label, value string }{
		{"年", "year"}, {"季度", "quarter"}, {"月", "month"}, {"ISO 周", "week"}, {"日", "day"},
	}
	out := make([]shBucketLink, 0, len(items))
	for _, item := range items {
		q := common.CloneValues(c)
		q.Set("level", item.value)
		q.Set("p", "1")
		href := c.Request.URL.Path
		if encoded := q.Encode(); encoded != "" {
			href += "?" + encoded
		}
		out = append(out, shBucketLink{Label: item.label, Href: href, Active: current == item.value})
	}
	return out
}
