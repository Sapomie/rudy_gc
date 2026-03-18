package handler

import (
	"rudy_gc/internal/service/movie"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ------- All/Release 路由 -------

func (h *MovieAggHTMLHandler) MovieAggAllReleaseYears(c *gin.Context)   { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseMonths(c *gin.Context)  { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseQuarter(c *gin.Context) { h.aggAllRelease(c) }
func (h *MovieAggHTMLHandler) MovieAggAllReleaseMonth(c *gin.Context)   { h.aggAllRelease(c) }

func (h *MovieAggHTMLHandler) aggAllRelease(c *gin.Context) {
	// --- query / path ---
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

	// --- 调 domain ---
	vm, err := h.movieSvc.BuildAllReleaseAggView(c.Request.Context(), movie.AggParams{
		Mode:    "release", // 固定
		Year:    year,
		Quarter: quarter,
		Month:   month,
		OrderBy: curOD,
		Page:    page,
		Size:    size,
		TopN:    topN,
	})
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// --- HTTP 层补充 PageInfo / SortQuery ---
	data := map[string]any{
		"Title":           vm.Title,
		"Breadcrumbs":     vm.Breadcrumbs,
		"Level":           vm.Level,
		"RangeStart":      vm.RangeStart,
		"RangeEnd":        vm.RangeEnd,
		"Movies":          vm.Movies,
		"Total":           vm.Total,
		"BucketsAll":      vm.BucketsAll,
		"BucketsQAll":     vm.BucketsQAll,
		"BucketsMAll":     vm.BucketsMAll,
		"TopCastsAll":     vm.TopCastsAll,
		"TopDirectorsAll": vm.TopDirectorsAll,
		"TopLabelsAll":    vm.TopLabelsAll,
		"TopPrefixesAll":  vm.TopPrefixesAll,
		"Mode":            "release_all",
	}

	pi := BuildPageInfo(c, vm.Total, int64(page), int64(size), pageWindow)
	data["PageInfo"] = pi
	data["SortQuery"] = sq
	data["CurrentSort"] = curOD

	c.HTML(200, "page.movie_agg_all_time", data)
}
