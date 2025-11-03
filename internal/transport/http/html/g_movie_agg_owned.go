package html

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type MovieAggHTMLHandler struct {
	deps     *svc.Deps
	movieSvc *movie.MovieService
}

func NewMovieAggHTMLHandler(deps *svc.Deps) *MovieAggHTMLHandler {
	return &MovieAggHTMLHandler{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}

const (
	defaultAggPageSize = 24
	orderByRelease     = "releasing_date"
	orderByBirth       = "birth_time"
)

// ------- Owned 路由 -------

func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseYears(c *gin.Context) {
	h.aggOwnedCommon(c, "release", orderByRelease)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseMonths(c *gin.Context) {
	h.aggOwnedCommon(c, "release", orderByRelease)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseQuarter(c *gin.Context) {
	h.aggOwnedCommon(c, "release", orderByRelease)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseMonth(c *gin.Context) {
	h.aggOwnedCommon(c, "release", orderByRelease)
}

func (h *MovieAggHTMLHandler) MovieAggOwnedBirthYears(c *gin.Context) {
	h.aggOwnedCommon(c, "birth", orderByBirth)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedBirthMonths(c *gin.Context) {
	h.aggOwnedCommon(c, "birth", orderByBirth)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedBirthQuarter(c *gin.Context) {
	h.aggOwnedCommon(c, "birth", orderByBirth)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedBirthMonth(c *gin.Context) {
	h.aggOwnedCommon(c, "birth", orderByBirth)
}

func (h *MovieAggHTMLHandler) aggOwnedCommon(c *gin.Context, mode, defaultOD string) {
	// --- query / path ---
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultAggPageSize)), defaultAggPageSize))
	curOD := normalizeOrderBy(c.Query("od"), defaultOD)
	sq := buildSortQuery(c, curOD)

	year, _ := strconv.Atoi(c.Param("year"))
	quarter, _ := strconv.Atoi(c.Param("q"))
	month, _ := strconv.Atoi(c.Param("month"))

	// TopN
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
	vm, err := h.movieSvc.BuildOwnedAggView(c.Request.Context(), movie.AggParams{
		Mode:    mode,
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
		// 直接铺 domain 返回
		"Title":        vm.Title,
		"Breadcrumbs":  vm.Breadcrumbs,
		"Level":        vm.Level,
		"RangeStart":   vm.RangeStart,
		"RangeEnd":     vm.RangeEnd,
		"Movies":       vm.Movies,
		"Total":        vm.Total,
		"BucketsY":     vm.BucketsY,
		"BucketsQ":     vm.BucketsQ,
		"BucketsM":     vm.BucketsM,
		"TopCasts":     vm.TopCasts,
		"TopDirectors": vm.TopDirectors,
		"TopLabels":    vm.TopLabels,
		"TopPrefixes":  vm.TopPrefixes,
		"Mode":         mode,
	}

	// PageInfo/SortQuery
	pi := BuildPageInfo(c, vm.Total, int64(page), int64(size), pageWindow)
	data["PageInfo"] = pi
	data["SortQuery"] = sq
	data["CurrentSort"] = curOD

	c.HTML(200, "page.movie_agg_owned_time", data)
}

func atoiDef(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
func clampPageSize(ps int) int {
	if ps <= 0 {
		return defaultPageSize
	}
	if ps > maxPageSize {
		return maxPageSize
	}
	return ps
}
