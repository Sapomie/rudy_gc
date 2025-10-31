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

/************ 基础工具 ************/

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

/************ TopN 读取（从外部传入） ************/

func clampTopN(n int) int {
	if n < 1 {
		return 1
	}
	if n > 200 {
		return 200
	}
	return n
}

// 支持 tn、top、topn、tc 这几个 query 键，默认 30
func readTopN(c *gin.Context) int {
	keys := []string{"tn", "top", "topn", "tc"}
	for _, k := range keys {
		if v := c.Query(k); v != "" {
			return clampTopN(atoiDef(v, 30))
		}
	}
	return 30
}

/************ 分桶结构（含总大小） ************/

type YearBucket struct {
	Year   int
	Count  int
	SizeGB float64
	Href   string
	Label  string
}

type QuarterBucket struct {
	Quarter int
	Count   int
	SizeGB  float64
	Href    string
	Label   string
}

type MonthBucket struct {
	Month  int
	Count  int
	SizeGB float64
	Href   string
	Label  string
}

/************ 演员 Top 结构 ************/

type CastStat struct {
	Name  string
	Count int   // 出现次数（影片数）
	ScSum int64 // 该演员出现的影片的 ScTimes 求和
}

func buildTopCasts(movies []*types.MovieType, topN int) []CastStat {
	if topN <= 0 {
		topN = 20
	}
	type agg struct {
		n     int
		scsum int64
	}
	mp := make(map[string]*agg, 1024)

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
		out = append(out, CastStat{
			Name:  name,
			Count: a.n,
			ScSum: a.scsum,
		})
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

/************ 工具：季与时间范围 ************/

func monthToQuarter(m int) int {
	return ((m - 1) / 3) + 1 // 1..4
}

func quarterRange(year int, q int) (start, end string) {
	if q < 1 {
		q = 1
	}
	if q > 4 {
		q = 4
	}
	startMonth := 3*(q-1) + 1
	start = fmt.Sprintf("%04d-%02d-01", year, startMonth)
	// quarter end month = startMonth+2
	endMonth := startMonth + 2
	// end day = next month first day - 1d
	e := time.Date(year, time.Month(endMonth)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)
	end = e.Format("2006-01-02")
	return
}

/************ 面包屑 ************/

type Breadcrumb struct {
	Title string
	Href  string // 为空表示当前节点
}

// 年→季→月
func buildAggBreadcrumbs(mode string, year, quarter, month int) []Breadcrumb {
	var rootTitle, rootHref string
	if mode == "birth" {
		rootTitle = "下载日"
		rootHref = "/movie-agg/birth"
	} else {
		rootTitle = "上映日"
		rootHref = "/movie-agg/release"
	}
	bcs := []Breadcrumb{{Title: rootTitle, Href: rootHref}}

	if year > 0 && quarter == 0 && month == 0 {
		bcs = append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: ""})
		return bcs
	}
	if year > 0 && quarter > 0 && month == 0 {
		yearHref := fmt.Sprintf("%s/%d", rootHref, year)
		bcs = append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: ""},
		)
		return bcs
	}
	if year > 0 && quarter > 0 && month > 0 {
		yearHref := fmt.Sprintf("%s/%d", rootHref, year)
		qHref := fmt.Sprintf("%s/%d/q/%d", rootHref, year, quarter)
		bcs = append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: qHref},
			Breadcrumb{Title: fmt.Sprintf("%02d 月", month), Href: ""},
		)
		return bcs
	}
	// 根层
	bcs[0].Href = ""
	return bcs
}

/************ Handler ************/

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

/* ------------------------- 上映日（release） ------------------------- */

// /movie-agg/release
func (h *MovieAggHTMLHandler) MovieAggReleaseYears(c *gin.Context) {
	topN := readTopN(c)

	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	// 全量一次性拉取（既用于年桶，也用于顶层 TopCasts）
	allReq := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		Page:     1,
		PageSize: 999999,
	}
	allResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// 年桶
	type aggY struct {
		n     int
		bytes int64
	}
	yearAgg := map[int]*aggY{}
	for _, m := range allResp.List {
		t, ok := parseYMD(m.ReleasingDate)
		if !ok {
			continue
		}
		y := t.Year()
		if yearAgg[y] == nil {
			yearAgg[y] = &aggY{}
		}
		yearAgg[y].n++
		if m.VFilm != nil {
			yearAgg[y].bytes += m.VFilm.Size
		}
	}
	years := make([]YearBucket, 0, len(yearAgg))
	for y, a := range yearAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		years = append(years, YearBucket{
			Year:   y,
			Count:  a.n,
			SizeGB: gb,
			Href:   fmt.Sprintf("/movie-agg/release/%d", y),
			Label:  fmt.Sprintf("%d 年", y),
		})
	}
	sort.Slice(years, func(i, j int) bool { return years[i].Year > years[j].Year })

	// 顶层 Top 演员：全量
	topCasts := buildTopCasts(allResp.List, topN)

	// 顶层卡片（当前排序分页）
	listReq := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  curOD,
		Page:     int64(page),
		PageSize: int64(size),
	}
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("release", 0, 0, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": "电影聚合 · 上映日",
		"Mode":  "release",
		"Year":  0, "Quarter": 0, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    years, "BucketsQ": nil, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		// 顶层不限定日期 → 不传 RangeStart/RangeEnd
	})
}

// /movie-agg/release/:year   （现在展示“季”桶）
func (h *MovieAggHTMLHandler) MovieAggReleaseMonths(c *gin.Context) {
	// 保留函数名以兼容你现有 router，但语义变为展示 Quarter 桶
	topN := readTopN(c)

	year, _ := strconv.Atoi(c.Param("year"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	// 该年全量（用于季度桶 + TopCasts）
	rReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: fmt.Sprintf("%04d-01-01", year),
		ReleasingDateEnd:   fmt.Sprintf("%04d-12-31", year),
		Page:               1, PageSize: 999999,
	}
	rResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), rReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// 聚合到季度
	type aggQ struct {
		n     int
		bytes int64
	}
	qAgg := map[int]*aggQ{} // key: 1..4
	for _, m := range rResp.List {
		t, ok := parseYMD(m.ReleasingDate)
		if !ok {
			continue
		}
		q := monthToQuarter(int(t.Month()))
		if qAgg[q] == nil {
			qAgg[q] = &aggQ{}
		}
		qAgg[q].n++
		if m.VFilm != nil {
			qAgg[q].bytes += m.VFilm.Size
		}
	}
	quarters := make([]QuarterBucket, 0, 4)
	for q, a := range qAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		quarters = append(quarters, QuarterBucket{
			Quarter: q,
			Count:   a.n,
			SizeGB:  gb,
			Href:    fmt.Sprintf("/movie-agg/release/%d/q/%d", year, q),
			Label:   fmt.Sprintf("Q%d 季", q),
		})
	}
	sort.Slice(quarters, func(i, j int) bool { return quarters[i].Quarter < quarters[j].Quarter })

	topCasts := buildTopCasts(rResp.List, topN)

	// 年列表（保留原逻辑：展示整年的分页卡片）
	listReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: fmt.Sprintf("%04d-01-01", year),
		ReleasingDateEnd:   fmt.Sprintf("%04d-12-31", year),
		OrderBy:            curOD, Page: int64(page), PageSize: int64(size),
	}
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("release", year, 0, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("上映日 · %d 年", year),
		"Mode":  "release", "Year": year, "Quarter": 0, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsQ": quarters, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		"RangeStart": fmt.Sprintf("%04d-01-01", year),
		"RangeEnd":   fmt.Sprintf("%04d-12-31", year),
	})
}

// /movie-agg/release/:year/q/:q   （季页：展示该季的月份桶 + 季内卡片）
func (h *MovieAggHTMLHandler) MovieAggReleaseQuarter(c *gin.Context) {
	topN := readTopN(c)

	year, _ := strconv.Atoi(c.Param("year"))
	q, _ := strconv.Atoi(c.Param("q"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	start, end := quarterRange(year, q)

	// 季内全量（用于月份桶 + TopCasts）
	allReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: start, ReleasingDateEnd: end,
		Page: 1, PageSize: 999999,
	}
	allResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// 季内按月份聚合
	type aggM struct {
		n     int
		bytes int64
	}
	monAgg := map[int]*aggM{}
	for _, m := range allResp.List {
		t, ok := parseYMD(m.ReleasingDate)
		if !ok {
			continue
		}
		mm := int(t.Month())
		if monAgg[mm] == nil {
			monAgg[mm] = &aggM{}
		}
		monAgg[mm].n++
		if m.VFilm != nil {
			monAgg[mm].bytes += m.VFilm.Size
		}
	}
	months := make([]MonthBucket, 0, 3)
	for m, a := range monAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		months = append(months, MonthBucket{
			Month: m, Count: a.n, SizeGB: gb,
			Href:  fmt.Sprintf("/movie-agg/release/%d/%02d", year, m),
			Label: fmt.Sprintf("%02d 月", m),
		})
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Month < months[j].Month })

	topCasts := buildTopCasts(allResp.List, topN)

	// 季内列表（分页 + 排序）
	req := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: start, ReleasingDateEnd: end,
		OrderBy: curOD, Page: int64(page), PageSize: int64(size),
	}
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, resp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("release", year, q, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("上映日 · %d 年 Q%d", year, q),
		"Mode":  "release", "Year": year, "Quarter": q, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsQ": nil, "BucketsM": months,
		"TopCasts": topCasts,
		"Movies":   resp.List, "Total": resp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		"RangeStart": start, "RangeEnd": end,
	})
}

// /movie-agg/release/:year/:month  （保持不变：月页）
func (h *MovieAggHTMLHandler) MovieAggReleaseMonth(c *gin.Context) {
	topN := readTopN(c)

	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour).Format("2006-01-02")

	// 推导季号用于面包屑
	q := monthToQuarter(month)

	allReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: start, ReleasingDateEnd: end,
		Page: 1, PageSize: 999999,
	}
	allResp, _ := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	topCasts := buildTopCasts(allResp.List, topN)

	req := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: start, ReleasingDateEnd: end,
		OrderBy: curOD, Page: int64(page), PageSize: int64(size),
	}
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, resp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("release", year, q, month)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("上映日 · %d 年 %02d 月", year, month),
		"Mode":  "release", "Year": year, "Quarter": q, "Month": month,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsQ": nil, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   resp.List, "Total": resp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		"RangeStart": start, "RangeEnd": end,
	})
}

/* ------------------------- 下载日（birth）------------------------- */

// /movie-agg/birth
func (h *MovieAggHTMLHandler) MovieAggBirthYears(c *gin.Context) {
	topN := readTopN(c)

	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	// 全量一次性拉取（既用于年桶，也用于顶层 TopCasts）
	allReq := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		Page:     1,
		PageSize: 999999,
	}
	allResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// 年桶（按下载日）
	type aggY struct {
		n     int
		bytes int64
	}
	yearAgg := map[int]*aggY{}
	for _, m := range allResp.List {
		t, ok := parseYMD(m.FilmBirthDate)
		if !ok {
			continue
		}
		y := t.Year()
		if yearAgg[y] == nil {
			yearAgg[y] = &aggY{}
		}
		yearAgg[y].n++
		if m.VFilm != nil {
			yearAgg[y].bytes += m.VFilm.Size
		}
	}
	years := make([]YearBucket, 0, len(yearAgg))
	for y, a := range yearAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		years = append(years, YearBucket{
			Year:   y,
			Count:  a.n,
			SizeGB: gb,
			Href:   fmt.Sprintf("/movie-agg/birth/%d", y),
			Label:  fmt.Sprintf("%d 年", y),
		})
	}
	sort.Slice(years, func(i, j int) bool { return years[i].Year > years[j].Year })

	// 顶层 Top 演员：全量
	topCasts := buildTopCasts(allResp.List, topN)

	// 顶层卡片（当前排序分页）
	listReq := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  curOD,
		Page:     int64(page),
		PageSize: int64(size),
	}
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("birth", 0, 0, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": "电影聚合 · 下载日",
		"Mode":  "birth",
		"Year":  0, "Quarter": 0, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    years, "BucketsQ": nil, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		// 顶层不限定日期 → 不传 RangeStart/RangeEnd
	})
}

// /movie-agg/birth/:year   （现在展示“季”桶）
func (h *MovieAggHTMLHandler) MovieAggBirthMonths(c *gin.Context) {
	// 保留函数名以兼容你现有 router，但语义变为展示 Quarter 桶
	topN := readTopN(c)

	year, _ := strconv.Atoi(c.Param("year"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	// 该年全量（用于季度桶 + TopCasts）
	rReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: fmt.Sprintf("%04d-01-01", year),
		FilmBirthTimeEnd:   fmt.Sprintf("%04d-12-31", year),
		Page:               1, PageSize: 999999,
	}
	rResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), rReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// 聚合到季度
	type aggQ struct {
		n     int
		bytes int64
	}
	qAgg := map[int]*aggQ{} // 1..4
	for _, m := range rResp.List {
		t, ok := parseYMD(m.FilmBirthDate)
		if !ok {
			continue
		}
		q := monthToQuarter(int(t.Month()))
		if qAgg[q] == nil {
			qAgg[q] = &aggQ{}
		}
		qAgg[q].n++
		if m.VFilm != nil {
			qAgg[q].bytes += m.VFilm.Size
		}
	}
	quarters := make([]QuarterBucket, 0, 4)
	for q, a := range qAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		quarters = append(quarters, QuarterBucket{
			Quarter: q,
			Count:   a.n,
			SizeGB:  gb,
			Href:    fmt.Sprintf("/movie-agg/birth/%d/q/%d", year, q),
			Label:   fmt.Sprintf("Q%d 季", q),
		})
	}
	sort.Slice(quarters, func(i, j int) bool { return quarters[i].Quarter < quarters[j].Quarter })

	topCasts := buildTopCasts(rResp.List, topN)

	// 年列表（分页 + 排序）
	listReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: fmt.Sprintf("%04d-01-01", year),
		FilmBirthTimeEnd:   fmt.Sprintf("%04d-12-31", year),
		OrderBy:            curOD, Page: int64(page), PageSize: int64(size),
	}
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("birth", year, 0, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("下载日 · %d 年", year),
		"Mode":  "birth", "Year": year, "Quarter": 0, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsQ": quarters, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		"RangeStart": fmt.Sprintf("%04d-01-01", year),
		"RangeEnd":   fmt.Sprintf("%04d-12-31", year),
	})
}

// /movie-agg/birth/:year/q/:q  （季页）
func (h *MovieAggHTMLHandler) MovieAggBirthQuarter(c *gin.Context) {
	topN := readTopN(c)

	year, _ := strconv.Atoi(c.Param("year"))
	q, _ := strconv.Atoi(c.Param("q"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	start, end := quarterRange(year, q)

	// 季内全量（用于月份桶 + TopCasts）
	allReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: start, FilmBirthTimeEnd: end,
		Page: 1, PageSize: 999999,
	}
	allResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	// 季内按月份聚合
	type aggM struct {
		n     int
		bytes int64
	}
	monAgg := map[int]*aggM{}
	for _, m := range allResp.List {
		t, ok := parseYMD(m.FilmBirthDate)
		if !ok {
			continue
		}
		mm := int(t.Month())
		if monAgg[mm] == nil {
			monAgg[mm] = &aggM{}
		}
		monAgg[mm].n++
		if m.VFilm != nil {
			monAgg[mm].bytes += m.VFilm.Size
		}
	}
	months := make([]MonthBucket, 0, 3)
	for m, a := range monAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		months = append(months, MonthBucket{
			Month: m, Count: a.n, SizeGB: gb,
			Href:  fmt.Sprintf("/movie-agg/birth/%d/%02d", year, m),
			Label: fmt.Sprintf("%02d 月", m),
		})
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Month < months[j].Month })

	topCasts := buildTopCasts(allResp.List, topN)

	// 季内列表（分页 + 排序）
	req := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: start, FilmBirthTimeEnd: end,
		OrderBy: curOD, Page: int64(page), PageSize: int64(size),
	}
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, resp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("birth", year, q, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("下载日 · %d 年 Q%d", year, q),
		"Mode":  "birth", "Year": year, "Quarter": q, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsQ": nil, "BucketsM": months,
		"TopCasts": topCasts,
		"Movies":   resp.List, "Total": resp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		"RangeStart": start, "RangeEnd": end,
	})
}

// /movie-agg/birth/:year/:month  （保持不变：月页）
func (h *MovieAggHTMLHandler) MovieAggBirthMonth(c *gin.Context) {
	topN := readTopN(c)

	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour).Format("2006-01-02")

	// 推导季号用于面包屑
	q := monthToQuarter(month)

	allReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: start, FilmBirthTimeEnd: end,
		Page: 1, PageSize: 999999,
	}
	allResp, _ := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	topCasts := buildTopCasts(allResp.List, topN)

	req := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: start, FilmBirthTimeEnd: end,
		OrderBy: curOD, Page: int64(page), PageSize: int64(size),
	}
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, resp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs("birth", year, q, month)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("下载日 · %d 年 %02d 月", year, month),
		"Mode":  "birth", "Year": year, "Quarter": q, "Month": month,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsQ": nil, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   resp.List, "Total": resp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
		"RangeStart": start, "RangeEnd": end,
	})
}
