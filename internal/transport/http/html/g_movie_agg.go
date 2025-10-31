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

/************ 本文件内使用的默认值 ************/
const defaultPS = 18 // 统一把默认页面大小改为 18

/************ 基础工具 ************/

func atoiDef(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func clampPageSize(ps, def, max int) int {
	if ps <= 0 {
		return def
	}
	if max > 0 && ps > max {
		return max
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

/************ 分桶结构（含总大小 GB） ************/

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

/************ 面包屑 ************/

type Breadcrumb struct {
	Title string
	Href  string // 为空表示当前节点
}

func buildAggBreadcrumbs(mode string, year, month int) []Breadcrumb {
	var rootTitle, rootHref string
	if mode == "birth" {
		rootTitle = "下载日"
		rootHref = "/movie-agg/birth"
	} else {
		rootTitle = "上映日"
		rootHref = "/movie-agg/release"
	}
	bcs := []Breadcrumb{{Title: rootTitle, Href: rootHref}}

	if year > 0 && month <= 0 {
		// 年层
		bcs = append(bcs, Breadcrumb{
			Title: fmt.Sprintf("%d 年", year),
			Href:  "", // 当前
		})
		return bcs
	}
	if year > 0 && month > 0 {
		// 月层
		yearHref := fmt.Sprintf("%s/%d", rootHref, year)
		bcs = append(bcs, Breadcrumb{
			Title: fmt.Sprintf("%d 年", year),
			Href:  yearHref,
		})
		bcs = append(bcs, Breadcrumb{
			Title: fmt.Sprintf("%02d 月", month),
			Href:  "", // 当前
		})
		return bcs
	}
	// 根层：只有根节点，当前页
	bcs[0].Href = "" // 当前页不跳转
	return bcs
}

/************ 日期字段选择：release vs birth ************/

// 返回：默认排序键、提取 MovieType 日期字符串的函数、设置请求日期范围的函数
func pickDateOps(mode string) (
	defaultOD string,
	getDate func(*types.MovieType) string,
	applyRange func(req *types.ListMovieFullRequest, start, end string),
) {
	if mode == "birth" {
		defaultOD = consts.OrderByBirthTime
		getDate = func(m *types.MovieType) string { return m.FilmBirthDate }
		applyRange = func(req *types.ListMovieFullRequest, start, end string) {
			req.FilmBirthTimeStart = start
			req.FilmBirthTimeEnd = end
		}
	} else {
		defaultOD = consts.OrderByReleasingDate
		getDate = func(m *types.MovieType) string { return m.ReleasingDate }
		applyRange = func(req *types.ListMovieFullRequest, start, end string) {
			req.ReleasingDateStart = start
			req.ReleasingDateEnd = end
		}
	}
	return
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

/* ------------------------- 对外路由包装（保持你现有的 6 个路由） ------------------------- */

// 上映日（release）
func (h *MovieAggHTMLHandler) MovieAggReleaseYears(c *gin.Context)  { h.aggYears(c, "release") }
func (h *MovieAggHTMLHandler) MovieAggReleaseMonths(c *gin.Context) { h.aggMonths(c, "release") }
func (h *MovieAggHTMLHandler) MovieAggReleaseMonth(c *gin.Context)  { h.aggMonth(c, "release") }

// 下载日（birth）
func (h *MovieAggHTMLHandler) MovieAggBirthYears(c *gin.Context)  { h.aggYears(c, "birth") }
func (h *MovieAggHTMLHandler) MovieAggBirthMonths(c *gin.Context) { h.aggMonths(c, "birth") }
func (h *MovieAggHTMLHandler) MovieAggBirthMonth(c *gin.Context)  { h.aggMonth(c, "birth") }

/* ------------------------- 公共实现：年层 / 月层 / 某月卡片 ------------------------- */

// 顶层“年聚合”页：展示所有年份桶 + “最新年份”的卡片
func (h *MovieAggHTMLHandler) aggYears(c *gin.Context, mode string) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPS)), defaultPS), defaultPS, maxPageSize)

	defOD, getDate, applyRange := pickDateOps(mode)
	curOD := normalizeOrderBy(c.DefaultQuery("od", defOD), defOD)
	sq := buildSortQuery(c, curOD)

	// 拉全量聚合 “年桶”
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

	type agg struct {
		n     int
		bytes int64
	}
	yearAgg := map[int]*agg{}
	maxYear := 0
	for _, m := range allResp.List {
		if t, ok := parseYMD(getDate(m)); ok {
			y := t.Year()
			if y > maxYear {
				maxYear = y
			}
			if yearAgg[y] == nil {
				yearAgg[y] = &agg{}
			}
			yearAgg[y].n++
			if m.VFilm != nil {
				yearAgg[y].bytes += m.VFilm.Size
			}
		}
	}
	years := make([]YearBucket, 0, len(yearAgg))
	root := "/movie-agg/release"
	if mode == "birth" {
		root = "/movie-agg/birth"
	}
	for y, a := range yearAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		years = append(years, YearBucket{
			Year:   y,
			Count:  a.n,
			SizeGB: gb,
			Href:   fmt.Sprintf("%s/%d", root, y),
			Label:  fmt.Sprintf("%d 年", y),
		})
	}
	sort.Slice(years, func(i, j int) bool { return years[i].Year > years[j].Year })

	// 顶层卡片：展示“最新年份”的片子（避免与聚合语义不一致）
	listReq := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  curOD,
		Page:     int64(page),
		PageSize: int64(size),
	}
	if maxYear > 0 {
		start := fmt.Sprintf("%04d-01-01", maxYear)
		end := fmt.Sprintf("%04d-12-31", maxYear)
		applyRange(listReq, start, end)
	}
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs(mode, 0, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       "电影聚合 · " + map[string]string{"release": "上映日", "birth": "下载日"}[mode],
		"Mode":        mode,
		"Year":        0,
		"Month":       0,
		"Breadcrumbs": bcs,
		"BucketsY":    years,
		"BucketsM":    nil,
		"Movies":      listResp.List,
		"Total":       listResp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq, // 兼容你的 sort_bar
		"SortQuery":   sq, // 有的模板用了 SortQuery，有的用了 sortQuery，两个都给
		"CurrentSort": curOD,
	})
}

// 某“年聚合”页：展示 12 个月桶 + 当年的卡片
func (h *MovieAggHTMLHandler) aggMonths(c *gin.Context, mode string) {
	year, _ := strconv.Atoi(c.Param("year"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPS)), defaultPS), defaultPS, maxPageSize)

	defOD, getDate, applyRange := pickDateOps(mode)
	curOD := normalizeOrderBy(c.DefaultQuery("od", defOD), defOD)
	sq := buildSortQuery(c, curOD)

	// 该年的月桶
	rReq := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		Page:     1,
		PageSize: 999999,
	}
	applyRange(rReq, fmt.Sprintf("%04d-01-01", year), fmt.Sprintf("%04d-12-31", year))

	rResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), rReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}

	type agg struct {
		n     int
		bytes int64
	}
	monAgg := map[int]*agg{}
	for _, m := range rResp.List {
		if t, ok := parseYMD(getDate(m)); ok {
			mm := int(t.Month())
			if monAgg[mm] == nil {
				monAgg[mm] = &agg{}
			}
			monAgg[mm].n++
			if m.VFilm != nil {
				monAgg[mm].bytes += m.VFilm.Size
			}
		}
	}
	months := make([]MonthBucket, 0, len(monAgg))
	root := "/movie-agg/release"
	if mode == "birth" {
		root = "/movie-agg/birth"
	}
	for m, a := range monAgg {
		gb := float64(a.bytes) / (1024.0 * 1024.0 * 1024.0)
		months = append(months, MonthBucket{
			Month:  m,
			Count:  a.n,
			SizeGB: gb,
			Href:   fmt.Sprintf("%s/%d/%02d", root, year, m),
			Label:  fmt.Sprintf("%02d 月", m),
		})
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Month < months[j].Month })

	// 当年卡片（带排序/分页）
	listReq := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  curOD,
		Page:     int64(page),
		PageSize: int64(size),
	}
	applyRange(listReq, fmt.Sprintf("%04d-01-01", year), fmt.Sprintf("%04d-12-31", year))

	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs(mode, year, 0)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       fmt.Sprintf("%s · %d 年", map[string]string{"release": "上映日", "birth": "下载日"}[mode], year),
		"Mode":        mode,
		"Year":        year,
		"Month":       0,
		"Breadcrumbs": bcs,
		"BucketsY":    nil,
		"BucketsM":    months,
		"Movies":      listResp.List,
		"Total":       listResp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"SortQuery":   sq,
		"CurrentSort": curOD,
	})
}

// 某“年-月”卡片页：只展示该月的影片卡片
func (h *MovieAggHTMLHandler) aggMonth(c *gin.Context, mode string) {
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := clampPageSize(atoiDef(c.DefaultQuery("ps", strconv.Itoa(defaultPS)), defaultPS), defaultPS, maxPageSize)

	defOD, _, applyRange := pickDateOps(mode)
	curOD := normalizeOrderBy(c.DefaultQuery("od", defOD), defOD)
	sq := buildSortQuery(c, curOD)

	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := fmt.Sprintf("%04d-%02d-31", year, month)

	req := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  curOD,
		Page:     int64(page),
		PageSize: int64(size),
	}
	applyRange(req, start, end)

	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, resp.Total, int64(page), int64(size), pageWindow)
	bcs := buildAggBreadcrumbs(mode, year, month)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       fmt.Sprintf("%s · %d 年 %02d 月", map[string]string{"release": "上映日", "birth": "下载日"}[mode], year, month),
		"Mode":        mode,
		"Year":        year,
		"Month":       month,
		"Breadcrumbs": bcs,
		"BucketsY":    nil,
		"BucketsM":    nil,
		"Movies":      resp.List,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"SortQuery":   sq,
		"CurrentSort": curOD,
	})
}
