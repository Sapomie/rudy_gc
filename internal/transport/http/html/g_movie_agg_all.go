package html

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"

	"github.com/gin-gonic/gin"
)

/* ---------- Handler ---------- */
type MovieAggAllHTMLHandler struct {
	deps     *svc.Deps
	movieSvc *movie.MovieService
}

func NewMovieAggAllHTMLHandler(deps *svc.Deps) *MovieAggAllHTMLHandler {
	return &MovieAggAllHTMLHandler{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}

/* ---------- 路由入口（release 全量） ---------- */
func (h *MovieAggAllHTMLHandler) MovieAggAllReleaseYears(c *gin.Context)   { h.aggAllRelease(c) }
func (h *MovieAggAllHTMLHandler) MovieAggAllReleaseMonths(c *gin.Context)  { h.aggAllRelease(c) }
func (h *MovieAggAllHTMLHandler) MovieAggAllReleaseQuarter(c *gin.Context) { h.aggAllRelease(c) }
func (h *MovieAggAllHTMLHandler) MovieAggAllReleaseMonth(c *gin.Context)   { h.aggAllRelease(c) }

/* ---------- 面包屑（全量版） ---------- */
// 仅 release 全量；直接复用已有 Breadcrumb 类型（在 owned 文件里已定义）
func buildAllReleaseBreadcrumbs(year, quarter, month int) []Breadcrumb {
	rootTitle, rootHref := "上映日（全量）", "/movie-agg-all/release"
	bcs := []Breadcrumb{{Title: rootTitle, Href: rootHref}}

	if year == 0 {
		bcs[0].Href = ""
		return bcs
	}
	yearHref := fmt.Sprintf("%s/%d", rootHref, year)

	if quarter == 0 && month == 0 {
		bcs = append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: ""})
		return bcs
	}
	if month == 0 && quarter > 0 {
		bcs = append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: ""},
		)
		return bcs
	}
	if quarter == 0 && month > 0 {
		quarter = ((month - 1) / 3) + 1
	}
	qHref := fmt.Sprintf("%s/%d/q/%d", rootHref, year, quarter)
	bcs = append(bcs,
		Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
		Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: qHref},
		Breadcrumb{Title: fmt.Sprintf("%02d 月", month), Href: ""},
	)
	return bcs
}

/* ---------- 聚合桶（全量：总的/拥有的） ---------- */
// 复用包内已有 minYear / maxYear / monthToQuarter / parseYMD

type AllBucket struct {
	Label      string
	Href       string
	CountAll   int
	CountOwned int
	SizeAllGB  float64
}

func isOwned(m *types.MovieType) bool {
	if m == nil {
		return false
	}
	switch m.Owned {
	case consts.OwnedAllNotRemoved, consts.OwnedHasSubNotRemoved, consts.OwnedNoSubNotRemoved:
		return true
	default:
		return false
	}
}

func aggYearsAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, root string) []AllBucket {
	const span = maxYear - minYear + 1
	var cntAll = make([]int, span)
	var cntOwned = make([]int, span)
	var szAll = make([]int64, span)

	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok {
			y := t.Year()
			if y >= minYear && y <= maxYear {
				i := y - minYear
				cntAll[i]++
				if isOwned(m) {
					cntOwned[i]++
				}
				if m.VFilm != nil {
					szAll[i] += m.VFilm.Size
				}
			}
		}
	}
	out := make([]AllBucket, 0, 60)
	for i := len(cntAll) - 1; i >= 0; i-- {
		if cntAll[i] == 0 {
			continue
		}
		y := minYear + i
		gb := float64(szAll[i]) / (1024.0 * 1024.0 * 1024.0)
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%d 年", y),
			Href:       fmt.Sprintf("%s/%d", root, y),
			CountAll:   cntAll[i],
			CountOwned: cntOwned[i],
			SizeAllGB:  gb,
		})
	}
	return out
}

func aggQuartersAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, root string) []AllBucket {
	var cntAll = make([]int, 5)
	var cntOwned = make([]int, 5)
	var szAll = make([]int64, 5)

	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok && t.Year() == year {
			q := monthToQuarter(int(t.Month()))
			cntAll[q]++
			if isOwned(m) {
				cntOwned[q]++
			}
			if m.VFilm != nil {
				szAll[q] += m.VFilm.Size
			}
		}
	}
	out := make([]AllBucket, 0, 4)
	for q := 1; q <= 4; q++ {
		if cntAll[q] == 0 {
			continue
		}
		gb := float64(szAll[q]) / (1024.0 * 1024.0 * 1024.0)
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("Q%d 季", q),
			Href:       fmt.Sprintf("%s/%d/q/%d", root, year, q),
			CountAll:   cntAll[q],
			CountOwned: cntOwned[q],
			SizeAllGB:  gb,
		})
	}
	return out
}

func aggMonthsAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year, quarter int, root string) []AllBucket {
	startMonth := 3*(quarter-1) + 1
	var cntAll = make([]int, 3)
	var cntOwned = make([]int, 3)
	var szAll = make([]int64, 3)

	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok && t.Year() == year {
			mm := int(t.Month())
			if mm >= startMonth && mm <= startMonth+2 {
				i := mm - startMonth
				cntAll[i]++
				if isOwned(m) {
					cntOwned[i]++
				}
				if m.VFilm != nil {
					szAll[i] += m.VFilm.Size
				}
			}
		}
	}
	out := make([]AllBucket, 0, 3)
	for i := 0; i < 3; i++ {
		if cntAll[i] == 0 {
			continue
		}
		mm := startMonth + i
		gb := float64(szAll[i]) / (1024.0 * 1024.0 * 1024.0)
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%02d 月", mm),
			Href:       fmt.Sprintf("%s/%d/%02d", root, year, mm),
			CountAll:   cntAll[i],
			CountOwned: cntOwned[i],
			SizeAllGB:  gb,
		})
	}
	return out
}

// 全年 12 个月聚合（避免与 owned 版重名）
func aggMonthsYearAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, root string) []AllBucket {
	var cntAll [13]int
	var cntOwned [13]int
	var szAll [13]int64

	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok && t.Year() == year {
			mm := int(t.Month())
			cntAll[mm]++
			if isOwned(m) {
				cntOwned[mm]++
			}
			if m.VFilm != nil {
				szAll[mm] += m.VFilm.Size
			}
		}
	}
	out := make([]AllBucket, 0, 12)
	for mm := 1; mm <= 12; mm++ {
		if cntAll[mm] == 0 {
			continue
		}
		gb := float64(szAll[mm]) / (1024.0 * 1024.0 * 1024.0)
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%02d 月", mm),
			Href:       fmt.Sprintf("%s/%d/%02d", root, year, mm),
			CountAll:   cntAll[mm],
			CountOwned: cntOwned[mm],
			SizeAllGB:  gb,
		})
	}
	return out
}

/* ---------- Top（全量：总的/拥有的） ---------- */
type TopStatAll struct {
	Name       string
	CountAll   int
	CountOwned int
}

func buildTopCastsAll(movies []*types.MovieType, topN int) []TopStatAll {
	if topN <= 0 {
		topN = 20
	}
	type agg struct{ all, own int }
	mp := make(map[string]*agg, 1024)
	for _, m := range movies {
		own := isOwned(m)
		for _, c := range m.Cast {
			if c == nil || c.Name == "" {
				continue
			}
			a := mp[c.Name]
			if a == nil {
				a = &agg{}
				mp[c.Name] = a
			}
			a.all++
			if own {
				a.own++
			}
		}
	}
	out := make([]TopStatAll, 0, len(mp))
	for name, a := range mp {
		out = append(out, TopStatAll{Name: name, CountAll: a.all, CountOwned: a.own})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CountAll != out[j].CountAll {
			return out[i].CountAll > out[j].CountAll
		}
		if out[i].CountOwned != out[j].CountOwned {
			return out[i].CountOwned > out[j].CountOwned
		}
		return out[i].Name < out[j].Name
	})
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

func buildTopByFieldAll(movies []*types.MovieType, pick func(*types.MovieType) string, topN int) []TopStatAll {
	if topN <= 0 {
		topN = 20
	}
	type agg struct{ all, own int }
	mp := make(map[string]*agg, 1024)
	for _, m := range movies {
		k := pick(m)
		if k == "" || k == "nil" {
			continue
		}
		a := mp[k]
		if a == nil {
			a = &agg{}
			mp[k] = a
		}
		a.all++
		if isOwned(m) {
			a.own++
		}
	}
	out := make([]TopStatAll, 0, len(mp))
	for name, a := range mp {
		out = append(out, TopStatAll{Name: name, CountAll: a.all, CountOwned: a.own})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CountAll != out[j].CountAll {
			return out[i].CountAll > out[j].CountAll
		}
		if out[i].CountOwned != out[j].CountOwned {
			return out[i].CountOwned > out[j].CountOwned
		}
		return out[i].Name < out[j].Name
	})
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

func buildTopDirectorsAll(movies []*types.MovieType, topN int) []TopStatAll {
	return buildTopByFieldAll(movies, func(m *types.MovieType) string { return m.Director }, topN)
}
func buildTopLabelsAll(movies []*types.MovieType, topN int) []TopStatAll {
	return buildTopByFieldAll(movies, func(m *types.MovieType) string { return m.Label }, topN)
}
func buildTopPrefixesAll(movies []*types.MovieType, topN int) []TopStatAll {
	return buildTopByFieldAll(movies, func(m *types.MovieType) string { return m.Prefix }, topN)
}

/* ---------- 核心：全量 release 聚合 ---------- */
// 复用已有 aggLevel/levelRoot/levelYear/levelQuarter/levelMonth（在 owned 文件中）
func (h *MovieAggAllHTMLHandler) aggAllRelease(c *gin.Context) {
	topN := readTopN(c) // 复用已有 readTopN
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultAggPageSize)), defaultAggPageSize))
	curOD := normalizeOrderBy(c.Query("od"), "releasing_date")
	sq := buildSortQuery(c, curOD)

	year, _ := strconv.Atoi(c.Param("year"))
	quarter, _ := strconv.Atoi(c.Param("q"))
	month, _ := strconv.Atoi(c.Param("month"))

	var level aggLevel
	switch {
	case year == 0:
		level = levelRoot
	case quarter == 0 && month == 0:
		level = levelYear
	case month == 0:
		level = levelQuarter
	default:
		level = levelMonth
	}

	req := &types.ListMovieFullRequest{
		Owned:    consts.MovieAll,
		OrderBy:  curOD,
		Page:     int64(page),
		PageSize: int64(size),
	}
	var start, end string
	switch level {
	case levelRoot:
		// 根页：只要“年份入口卡片”（来源：已拥有集合）
		buckets, err := buildYearEntrancesByOwned(h, c)
		if err != nil {
			c.String(500, "加载失败: %v", err)
			return
		}
		data := gin.H{
			"Title":       "电影聚合（全量）· 上映日",
			"Breadcrumbs": buildAllReleaseBreadcrumbs(0, 0, 0),
			"BucketsAll":  buckets, // 只给年份入口
		}
		c.HTML(200, "page.movie_agg_all_time", data)
		return
	case levelYear:
		start = fmt.Sprintf("%04d-01-01", year)
		end = fmt.Sprintf("%04d-12-31", year)
	case levelQuarter:
		start, end = quarterRange(year, quarter) // 复用已有 quarterRange
	case levelMonth:
		start = fmt.Sprintf("%04d-%02d-01", year, month)
		end = time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour).Format("2006-01-02")
	}
	req.ReleasingDateStart, req.ReleasingDateEnd = start, end

	// 全量用于聚合/Top
	fullReq := *req
	fullReq.Page, fullReq.PageSize = 1, 999999
	fullResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), &fullReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	root := "/movie-agg-all/release"
	dateFn := func(m *types.MovieType) string { return m.ReleasingDate }

	var (
		buckets     []AllBucket // 通用：季度页的“该季的 3 个月卡”等
		bucketsQAll []AllBucket // 年份页：季度卡
		bucketsMAll []AllBucket // 年份页：月份卡（1..12）
	)
	switch level {
	case levelYear:
		bucketsQAll = aggQuartersAll(fullResp.List, dateFn, year, root)
		bucketsMAll = aggMonthsYearAll(fullResp.List, dateFn, year, root)
	case levelQuarter:
		buckets = aggMonthsAll(fullResp.List, dateFn, year, quarter, root)
	}

	// Top
	topCasts := buildTopCastsAll(fullResp.List, topN)
	topDirectors := buildTopDirectorsAll(fullResp.List, topN)
	topLabels := buildTopLabelsAll(fullResp.List, topN)
	topPrefixes := buildTopPrefixesAll(fullResp.List, topN)

	// 列表（保留，模板可判空）
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAllReleaseBreadcrumbs(year, quarter, month)

	data := gin.H{
		"Mode":        "release_all",
		"Year":        year,
		"Quarter":     quarter,
		"Month":       month,
		"Breadcrumbs": bcs,

		// 通用桶（非年份页使用；年份页用下面两个字段）
		"BucketsAll": buckets,

		// 年份页 Tabs：季度/月份
		"BucketsQAll": bucketsQAll,
		"BucketsMAll": bucketsMAll,

		// Top
		"TopCastsAll":     topCasts,
		"TopDirectorsAll": topDirectors,
		"TopLabelsAll":    topLabels,
		"TopPrefixesAll":  topPrefixes,

		// 列表/分页/排序
		"Movies":      listResp.List,
		"Total":       listResp.Total,
		"PageInfo":    pi,
		"SortQuery":   sq,
		"CurrentSort": curOD,
	}

	if level != levelRoot {
		data["RangeStart"] = start
		data["RangeEnd"] = end
	}

	switch level {
	case levelRoot:
		data["Title"] = "电影聚合（全量）· 上映日"
	case levelYear:
		data["Title"] = fmt.Sprintf("上映日（全量）· %d 年", year)
	case levelQuarter:
		data["Title"] = fmt.Sprintf("上映日（全量）· %d 年 Q%d", year, quarter)
	case levelMonth:
		data["Title"] = fmt.Sprintf("上映日（全量）· %d 年 %02d 月", year, month)
	}

	c.HTML(200, "page.movie_agg_all_time", data)
}

// 根页年份入口：仅用“已拥有”集合判断有哪些年份，避免拉全量
func buildYearEntrancesByOwned(h *MovieAggAllHTMLHandler, c *gin.Context) ([]AllBucket, error) {
	alt := types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  "releasing_date",
		Page:     1,
		PageSize: 999999,
	}
	altResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), &alt)
	if err != nil {
		return nil, err
	}
	root := "/movie-agg-all/release"
	dateFn := func(m *types.MovieType) string { return m.ReleasingDate }
	buckets := aggYearsAll(altResp.List, dateFn, root)

	// 仅作“入口”，不展示全量统计
	for i := range buckets {
		buckets[i].CountAll = 0
		buckets[i].SizeAllGB = 0
	}
	return buckets, nil
}
