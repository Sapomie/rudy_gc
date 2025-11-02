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

/* ---------- 常量 & 基础工具 ---------- */
const (
	minYear = 1900 // 聚合用数组下界
	maxYear = 2100 // 聚合用数组上界

	// 排序字段（不再依赖 consts.OrderBy*）
	orderByRelease = "releasing_date"
	orderByBirth   = "birth_time"
)

/* ---------- 对外保留的 8 个函数（名称不变） ---------- */
func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseYears(c *gin.Context) {
	h.aggCommon(c, "release", orderByRelease)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseMonths(c *gin.Context) {
	h.aggCommon(c, "release", orderByRelease)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseQuarter(c *gin.Context) {
	h.aggCommon(c, "release", orderByRelease)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedReleaseMonth(c *gin.Context) {
	h.aggCommon(c, "release", orderByRelease)
}

func (h *MovieAggHTMLHandler) MovieAggOwnedBirthYears(c *gin.Context) {
	h.aggCommon(c, "birth", orderByBirth)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedBirthMonths(c *gin.Context) {
	h.aggCommon(c, "birth", orderByBirth)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedBirthQuarter(c *gin.Context) {
	h.aggCommon(c, "birth", orderByBirth)
}
func (h *MovieAggHTMLHandler) MovieAggOwnedBirthMonth(c *gin.Context) {
	h.aggCommon(c, "birth", orderByBirth)
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

func parseYMD(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil || t.IsZero() {
		return time.Time{}, false
	}
	return t, true
}

/* ---------- TopN ---------- */
func clampTopN(n int) int {
	if n < 1 {
		return 1
	}
	if n > 200 {
		return 200
	}
	return n
}

func readTopN(c *gin.Context) int {
	keys := []string{"tn", "top", "topn", "tc"}
	for _, k := range keys {
		if v := c.Query(k); v != "" {
			return clampTopN(atoiDef(v, 30))
		}
	}
	return 30
}

/* ---------- 通用桶结构 ---------- */
type Bucket struct {
	Label  string
	Href   string
	Count  int
	SizeGB float64
}

/* ---------- 演员 Top ---------- */
type CastStat struct {
	Name  string
	Count int
	ScSum int64
}

func buildTopCasts(movies []*types.MovieType, topN int) []CastStat {
	if topN <= 0 {
		topN = 20
	}
	type agg struct {
		n     int
		scsum int64
	}
	mp := make(map[string]*agg, len(movies)/10)

	for _, m := range movies {
		sc := m.ScTimes
		for _, c := range m.Cast {
			if c == nil || c.Name == "" {
				continue
			}
			a := mp[c.Name]
			if a == nil {
				a = &agg{}
				mp[c.Name] = a
			}
			a.n++
			a.scsum += sc
		}
	}
	out := make([]CastStat, 0, len(mp))
	for name, a := range mp {
		out = append(out, CastStat{Name: name, Count: a.n, ScSum: a.scsum})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		if out[i].ScSum != out[j].ScSum {
			return out[i].ScSum > out[j].ScSum
		}
		return out[i].Name < out[j].Name
	})
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

/* ---------- 时间工具 ---------- */
func monthToQuarter(m int) int { return ((m - 1) / 3) + 1 }

func quarterRange(year, q int) (start, end string) {
	if q < 1 {
		q = 1
	}
	if q > 4 {
		q = 4
	}
	startMonth := 3*(q-1) + 1
	start = fmt.Sprintf("%04d-%02d-01", year, startMonth)
	endTime := time.Date(year, time.Month(startMonth+3), 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	end = endTime.Format("2006-01-02")
	return
}

/* ---------- 面包屑 ---------- */
type Breadcrumb struct {
	Title string
	Href  string
}

// 年→季→月（当 quarter 未传但有 month 时，自动推导季）
func buildAggBreadcrumbs(mode string, year, quarter, month int) []Breadcrumb {
	var rootTitle, rootHref string
	if mode == "birth" {
		rootTitle, rootHref = "下载日", "/movie-agg-owned/birth"
	} else {
		rootTitle, rootHref = "上映日", "/movie-agg-owned/release"
	}
	bcs := []Breadcrumb{{Title: rootTitle, Href: rootHref}}

	// 根层
	if year == 0 {
		bcs[0].Href = ""
		return bcs
	}
	yearHref := fmt.Sprintf("%s/%d", rootHref, year)

	// 只有年
	if quarter == 0 && month == 0 {
		bcs = append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: ""})
		return bcs
	}

	// 只有季（标准季页）
	if month == 0 && quarter > 0 {
		bcs = append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: ""},
		)
		return bcs
	}

	// 月页：如果 quarter 未传，则用 month 推导
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

/* ---------- Handler ---------- */
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

/* ---------- 统一处理函数 ---------- */
type aggLevel int

const (
	levelRoot aggLevel = iota
	levelYear
	levelQuarter
	levelMonth
)

func (h *MovieAggHTMLHandler) aggCommon(c *gin.Context,
	mode string, // "release" or "birth"
	defaultOD string, // 排序字段字符串
) {
	topN := readTopN(c)

	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	if page < 1 {
		page = 1
	}
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.Query("od"), defaultOD)
	sq := buildSortQuery(c, curOD)

	// 解析路径参数
	year, _ := strconv.Atoi(c.Param("year"))
	quarter, _ := strconv.Atoi(c.Param("q"))
	month, _ := strconv.Atoi(c.Param("month"))

	// 确定层级
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

	// 构建请求
	req := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  curOD,
		Page:     int64(page),
		PageSize: int64(size),
	}
	var start, end string
	switch level {
	case levelRoot:
		// 全量
	case levelYear:
		start = fmt.Sprintf("%04d-01-01", year)
		end = fmt.Sprintf("%04d-12-31", year)
	case levelQuarter:
		start, end = quarterRange(year, quarter)
	case levelMonth:
		start = fmt.Sprintf("%04d-%02d-01", year, month)
		end = time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour).Format("2006-01-02")
	}
	if mode == "birth" {
		req.FilmBirthTimeStart, req.FilmBirthTimeEnd = start, end
	} else {
		req.ReleasingDateStart, req.ReleasingDateEnd = start, end
	}

	// 全量请求（聚合 + TopCasts）
	fullReq := *req
	fullReq.Page, fullReq.PageSize = 1, 999999
	fullResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), &fullReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// 聚合桶
	var buckets []Bucket
	dateFn := func(m *types.MovieType) string {
		if mode == "birth" {
			return m.FilmBirthDate
		}
		return m.ReleasingDate
	}
	switch level {
	case levelRoot:
		buckets = aggYears(fullResp.List, dateFn, mode)
	case levelYear:
		buckets = aggQuarters(fullResp.List, dateFn, year, mode)
	case levelQuarter:
		buckets = aggMonths(fullResp.List, dateFn, year, quarter, mode)
	}

	//topCasts := buildTopCasts(fullResp.List, topN)
	// 计算 Top 列表（基于 fullResp.List）
	topCasts := buildTopCasts(fullResp.List, topN)
	topDirectors := buildTopDirectors(fullResp.List, topN)
	topLabels := buildTopLabels(fullResp.List, topN)
	topPrefixes := buildTopPrefixes(fullResp.List, topN)

	// 分页列表
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs(mode, year, quarter, month)

	data := gin.H{
		"Mode":         mode,
		"Year":         year,
		"Quarter":      quarter,
		"Month":        month,
		"Breadcrumbs":  bcs,
		"TopCasts":     topCasts,
		"TopDirectors": topDirectors,
		"TopLabels":    topLabels,
		"TopPrefixes":  topPrefixes,
		"Movies":       listResp.List,
		"Total":        listResp.Total,
		"PageInfo":     pi,
		"sortQuery":    sq,
		"SortQuery":    sq,
		"CurrentSort":  curOD,
	}
	if level != levelRoot {
		data["RangeStart"] = start
		data["RangeEnd"] = end
	}
	switch level {
	case levelRoot:
		data["Title"] = "电影聚合 · " + map[string]string{"release": "上映日", "birth": "下载日"}[mode]
		data["BucketsY"] = buckets
	case levelYear:
		data["Title"] = fmt.Sprintf("%s · %d 年", map[string]string{"release": "上映日", "birth": "下载日"}[mode], year)
		data["BucketsQ"] = buckets
	case levelQuarter:
		data["Title"] = fmt.Sprintf("%s · %d 年 Q%d", map[string]string{"release": "上映日", "birth": "下载日"}[mode], year, quarter)
		data["BucketsM"] = buckets
	case levelMonth:
		q := monthToQuarter(month)
		data["Title"] = fmt.Sprintf("%s · %d 年 %02d 月", map[string]string{"release": "上映日", "birth": "下载日"}[mode], year, month)
		data["Quarter"] = q
	}

	c.HTML(200, "page.movie_agg_owned_time", data)
}

/* ---------- 聚合函数（数组版，性能高） ---------- */
func aggYears(movies []*types.MovieType, dateFn func(*types.MovieType) string, mode string) []Bucket {
	const span = maxYear - minYear + 1
	cnt := make([]int, span)
	sz := make([]int64, span)

	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok {
			y := t.Year()
			if y >= minYear && y <= maxYear {
				idx := y - minYear
				cnt[idx]++
				if m.VFilm != nil {
					sz[idx] += m.VFilm.Size
				}
			}
		}
	}
	out := make([]Bucket, 0, 50)
	root := "/movie-agg-owned/release"
	if mode == "birth" {
		root = "/movie-agg-owned/birth"
	}
	for i := len(cnt) - 1; i >= 0; i-- {
		if cnt[i] == 0 {
			continue
		}
		y := minYear + i
		gb := float64(sz[i]) / (1024.0 * 1024.0 * 1024.0)
		out = append(out, Bucket{
			Label:  fmt.Sprintf("%d 年", y),
			Href:   fmt.Sprintf("%s/%d", root, y),
			Count:  cnt[i],
			SizeGB: gb,
		})
	}
	return out
}

func aggQuarters(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, mode string) []Bucket {
	cnt := make([]int, 5) // 1..4
	sz := make([]int64, 5)

	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok && t.Year() == year {
			q := monthToQuarter(int(t.Month()))
			cnt[q]++
			if m.VFilm != nil {
				sz[q] += m.VFilm.Size
			}
		}
	}
	out := make([]Bucket, 0, 4)
	root := "/movie-agg-owned/release"
	if mode == "birth" {
		root = "/movie-agg-owned/birth"
	}
	for q := 1; q <= 4; q++ {
		if cnt[q] == 0 {
			continue
		}
		gb := float64(sz[q]) / (1024.0 * 1024.0 * 1024.0)
		out = append(out, Bucket{
			Label:  fmt.Sprintf("Q%d 季", q),
			Href:   fmt.Sprintf("%s/%d/q/%d", root, year, q),
			Count:  cnt[q],
			SizeGB: gb,
		})
	}
	return out
}

func aggMonths(movies []*types.MovieType, dateFn func(*types.MovieType) string, year, quarter int, mode string) []Bucket {
	startMonth := 3*(quarter-1) + 1
	cnt := make([]int, 3)
	sz := make([]int64, 3)

	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok {
			mm := int(t.Month())
			if mm >= startMonth && mm <= startMonth+2 {
				idx := mm - startMonth
				cnt[idx]++
				if m.VFilm != nil {
					sz[idx] += m.VFilm.Size
				}
			}
		}
	}
	out := make([]Bucket, 0, 3)
	root := "/movie-agg-owned/release"
	if mode == "birth" {
		root = "/movie-agg-owned/birth"
	}
	for i := 0; i < 3; i++ {
		if cnt[i] == 0 {
			continue
		}
		m := startMonth + i
		gb := float64(sz[i]) / (1024.0 * 1024.0 * 1024.0)
		out = append(out, Bucket{
			Label:  fmt.Sprintf("%02d 月", m),
			Href:   fmt.Sprintf("%s/%d/%02d", root, year, m),
			Count:  cnt[i],
			SizeGB: gb,
		})
	}
	return out
}

// --- 通用 Top 聚合（字符串字段） ---
type TopStat struct {
	Name  string
	Count int
	ScSum int64
}

// 复用 Cast 的排序逻辑：Count desc -> ScSum desc -> Name asc
func sortTopStats(a []TopStat) {
	sort.Slice(a, func(i, j int) bool {
		if a[i].Count != a[j].Count {
			return a[i].Count > a[j].Count
		}
		if a[i].ScSum != a[j].ScSum {
			return a[i].ScSum > a[j].ScSum
		}
		return a[i].Name < a[j].Name
	})
}

func buildTopByStringField(movies []*types.MovieType, pick func(*types.MovieType) string, topN int) []TopStat {
	if topN <= 0 {
		topN = 20
	}
	type agg struct {
		n     int
		scsum int64
	}
	mp := make(map[string]*agg, 1024)

	for _, m := range movies {
		name := pick(m)
		if name == "" || name == "nil" {
			continue
		}
		a := mp[name]
		if a == nil {
			a = &agg{}
			mp[name] = a
		}
		a.n++
		a.scsum += m.ScTimes
	}
	out := make([]TopStat, 0, len(mp))
	for name, a := range mp {
		out = append(out, TopStat{Name: name, Count: a.n, ScSum: a.scsum})
	}
	sortTopStats(out)
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

// --- 三个具体 Top（导演 / 厂牌Label / 前缀Prefix） ---
func buildTopDirectors(movies []*types.MovieType, topN int) []TopStat {
	return buildTopByStringField(movies, func(m *types.MovieType) string { return m.Director }, topN)
}
func buildTopLabels(movies []*types.MovieType, topN int) []TopStat {
	return buildTopByStringField(movies, func(m *types.MovieType) string { return m.Label }, topN)
}
func buildTopPrefixes(movies []*types.MovieType, topN int) []TopStat {
	return buildTopByStringField(movies, func(m *types.MovieType) string { return m.Prefix }, topN)
}
