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

/************ 分桶结构（含总大小） ************/

type YearBucket struct {
	Year   int
	Count  int
	SizeGB float64
	Href   string
	Label  string
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

/************ 面包屑 ************/

type Breadcrumb struct {
	Title string
	Href  string // 为空表示当前节点
}

func buildAggBreadcrumbs(mode string, year, month int) []Breadcrumb {
	var rootTitle, rootHref string
	if mode == "birth" {
		rootTitle = "拍摄日"
		rootHref = "/movie-agg/birth"
	} else {
		rootTitle = "上映日"
		rootHref = "/movie-agg/release"
	}
	bcs := []Breadcrumb{{Title: rootTitle, Href: rootHref}}

	if year > 0 && month <= 0 {
		bcs = append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: ""})
		return bcs
	}
	if year > 0 && month > 0 {
		yearHref := fmt.Sprintf("%s/%d", rootHref, year)
		bcs = append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("%02d 月", month), Href: ""},
		)
		return bcs
	}
	bcs[0].Href = "" // 根层当前页
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

	// 顶层 Top 演员：✅ 改为基于“全量 allResp.List”
	topCasts := buildTopCasts(allResp.List, 20)

	// 顶层卡片（保持现状：展示按当前排序的分页列表）
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
	bcs := buildAggBreadcrumbs("release", 0, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": "电影聚合 · 上映日",
		"Mode":  "release",
		"Year":  0, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    years, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
	})
}

// /movie-agg/release/:year
func (h *MovieAggHTMLHandler) MovieAggReleaseMonths(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	// 该年全量（用于月份桶 + TopCasts）
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

	type aggM struct {
		n     int
		bytes int64
	}
	monAgg := map[int]*aggM{}
	for _, m := range rResp.List {
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
	months := make([]MonthBucket, 0, len(monAgg))
	for m, a := range monAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		months = append(months, MonthBucket{
			Month: m, Count: a.n, SizeGB: gb,
			Href:  fmt.Sprintf("/movie-agg/release/%d/%02d", year, m),
			Label: fmt.Sprintf("%02d 月", m),
		})
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Month < months[j].Month })

	topCasts := buildTopCasts(rResp.List, 20)

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
	bcs := buildAggBreadcrumbs("release", year, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("上映日 · %d 年", year),
		"Mode":  "release", "Year": year, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsM": months,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
	})
}

// /movie-agg/release/:year/:month
func (h *MovieAggHTMLHandler) MovieAggReleaseMonth(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour).Format("2006-01-02")

	allReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: start, ReleasingDateEnd: end,
		Page: 1, PageSize: 999999,
	}
	allResp, _ := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	topCasts := buildTopCasts(allResp.List, 20)

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
	bcs := buildAggBreadcrumbs("release", year, month)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("上映日 · %d 年 %02d 月", year, month),
		"Mode":  "release", "Year": year, "Month": month,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   resp.List, "Total": resp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
	})
}

/* ------------------------- 拍摄日（birth）------------------------- */

// /movie-agg/birth
func (h *MovieAggHTMLHandler) MovieAggBirthYears(c *gin.Context) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	// 全量一次性拉取（既用于年桶，也用于顶层 TopCasts）✅ 修正：TopCasts 用全量
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

	// 年桶（按拍摄日）
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

	// 顶层 Top 演员：✅ 改为基于“全量 allResp.List”（修复你指出的问题）
	topCasts := buildTopCasts(allResp.List, 20)

	// 顶层卡片（保持现状：展示按当前排序的分页列表）
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
	bcs := buildAggBreadcrumbs("birth", 0, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": "电影聚合 · 拍摄日",
		"Mode":  "birth",
		"Year":  0, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    years, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
	})
}

// /movie-agg/birth/:year
func (h *MovieAggHTMLHandler) MovieAggBirthMonths(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	// 该年全量（用于月份桶 + TopCasts）
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
	type aggM struct {
		n     int
		bytes int64
	}
	monAgg := map[int]*aggM{}
	for _, m := range rResp.List {
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
	months := make([]MonthBucket, 0, len(monAgg))
	for m, a := range monAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		months = append(months, MonthBucket{
			Month: m, Count: a.n, SizeGB: gb,
			Href:  fmt.Sprintf("/movie-agg/birth/%d/%02d", year, m),
			Label: fmt.Sprintf("%02d 月", m),
		})
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Month < months[j].Month })

	topCasts := buildTopCasts(rResp.List, 20)

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
	bcs := buildAggBreadcrumbs("birth", year, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("拍摄日 · %d 年", year),
		"Mode":  "birth", "Year": year, "Month": 0,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsM": months,
		"TopCasts": topCasts,
		"Movies":   listResp.List, "Total": listResp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
	})
}

// /movie-agg/birth/:year/:month
func (h *MovieAggHTMLHandler) MovieAggBirthMonth(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPageSize)), defaultPageSize))
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour).Format("2006-01-02")

	allReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: start, FilmBirthTimeEnd: end,
		Page: 1, PageSize: 999999,
	}
	allResp, _ := h.movieSvc.ListMovieFull(c.Request.Context(), allReq)
	topCasts := buildTopCasts(allResp.List, 20)

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
	bcs := buildAggBreadcrumbs("birth", year, month)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title": fmt.Sprintf("拍摄日 · %d 年 %02d 月", year, month),
		"Mode":  "birth", "Year": year, "Month": month,
		"Breadcrumbs": bcs,
		"BucketsY":    nil, "BucketsM": nil,
		"TopCasts": topCasts,
		"Movies":   resp.List, "Total": resp.Total,
		"PageInfo": pi, "sortQuery": sq, "SortQuery": sq, "CurrentSort": curOD,
	})
}
