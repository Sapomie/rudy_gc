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

/* ==================== 聚合卡片用的桶结构（直接带 SizeGB 字符串，模板无需函数） ==================== */

type YearBucket struct {
	Year   int
	Count  int
	SizeGB string // 直接渲染，比如 "12.3 GB"
	Href   string
	Label  string
}

type MonthBucket struct {
	Month  int
	Count  int
	SizeGB string // 直接渲染，比如 "7.8 GB"
	Href   string
	Label  string
}

/* ==================== Handler ==================== */

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

// 聚合页默认分页：18
const defaultAggPageSize = 18

/* ------------------------- 上映日（release） ------------------------- */

// /movie-agg/release
func (h *MovieAggHTMLHandler) MovieAggReleaseYears(c *gin.Context) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := atoiDef(c.DefaultQuery("ps", fmt.Sprint(defaultAggPageSize)), defaultAggPageSize)

	// 排序（上映日默认 rd）
	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	// 年桶（全量 + 内存分桶 + 累加 Size）
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
		count int
		size  int64
	}
	ym := map[int]*agg{}
	for _, m := range allResp.List {
		t, ok := parseYMD(m.ReleasingDate)
		if !ok {
			continue
		}
		a := ym[t.Year()]
		if a == nil {
			a = &agg{}
			ym[t.Year()] = a
		}
		a.count++
		// 只在 owned 集合中统计 Size（VFilm 才有）
		if m.VFilm != nil {
			a.size += m.VFilm.Size
		}
	}
	buckets := make([]YearBucket, 0, len(ym))
	for y, a := range ym {
		buckets = append(buckets, YearBucket{
			Year:   y,
			Count:  a.count,
			SizeGB: toGB(a.size),
			Href:   fmt.Sprintf("/movie-agg/release/%d", y),
			Label:  fmt.Sprintf("%d 年", y),
		})
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Year > buckets[j].Year })

	// 年页卡片（分页 + 排序）
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

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       "电影聚合 · 上映日",
		"Buckets":     buckets,
		"Movies":      listResp.List,
		"Total":       listResp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"CurrentSort": curOD,
	})
}

// /movie-agg/release/:year
func (h *MovieAggHTMLHandler) MovieAggReleaseMonths(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := atoiDef(c.DefaultQuery("ps", fmt.Sprint(defaultAggPageSize)), defaultAggPageSize)

	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	rReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: fmt.Sprintf("%04d-01-01", year),
		ReleasingDateEnd:   fmt.Sprintf("%04d-12-31", year),
		Page:               1,
		PageSize:           999999,
	}
	rResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), rReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	type agg struct {
		count int
		size  int64
	}
	mm := map[int]*agg{}
	for _, m := range rResp.List {
		t, ok := parseYMD(m.ReleasingDate)
		if !ok {
			continue
		}
		a := mm[int(t.Month())]
		if a == nil {
			a = &agg{}
			mm[int(t.Month())] = a
		}
		a.count++
		if m.VFilm != nil {
			a.size += m.VFilm.Size
		}
	}
	mb := make([]MonthBucket, 0, len(mm))
	for mth, a := range mm {
		mb = append(mb, MonthBucket{
			Month:  mth,
			Count:  a.count,
			SizeGB: toGB(a.size),
			Href:   fmt.Sprintf("/movie-agg/release/%d/%02d", year, mth),
			Label:  fmt.Sprintf("%02d 月", mth),
		})
	}
	sort.Slice(mb, func(i, j int) bool { return mb[i].Month < mb[j].Month })

	listReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: fmt.Sprintf("%04d-01-01", year),
		ReleasingDateEnd:   fmt.Sprintf("%04d-12-31", year),
		OrderBy:            curOD,
		Page:               int64(page),
		PageSize:           int64(size),
	}
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       fmt.Sprintf("上映日 · %d 年", year),
		"Buckets":     mb,
		"Movies":      listResp.List,
		"Total":       listResp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"CurrentSort": curOD,
	})
}

// /movie-agg/release/:year/:month
func (h *MovieAggHTMLHandler) MovieAggReleaseMonth(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := atoiDef(c.DefaultQuery("ps", fmt.Sprint(defaultAggPageSize)), defaultAggPageSize)

	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByReleasingDate), consts.OrderByReleasingDate)
	sq := buildSortQuery(c, curOD)

	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := fmt.Sprintf("%04d-%02d-31", year, month)

	req := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		ReleasingDateStart: start,
		ReleasingDateEnd:   end,
		OrderBy:            curOD,
		Page:               int64(page),
		PageSize:           int64(size),
	}
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, resp.Total, int64(page), int64(size), pageWindow)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       fmt.Sprintf("上映日 · %d 年 %02d 月", year, month),
		"Movies":      resp.List,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"CurrentSort": curOD,
	})
}

/* ------------------------- 拍摄日（birth） ------------------------- */

// /movie-agg/birth
func (h *MovieAggHTMLHandler) MovieAggBirthYears(c *gin.Context) {
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := atoiDef(c.DefaultQuery("ps", fmt.Sprint(defaultAggPageSize)), defaultAggPageSize)

	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

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
		count int
		size  int64
	}
	ym := map[int]*agg{}
	for _, m := range allResp.List {
		t, ok := parseYMD(m.FilmBirthDate)
		if !ok {
			continue
		}
		a := ym[t.Year()]
		if a == nil {
			a = &agg{}
			ym[t.Year()] = a
		}
		a.count++
		if m.VFilm != nil {
			a.size += m.VFilm.Size
		}
	}
	buckets := make([]YearBucket, 0, len(ym))
	for y, a := range ym {
		buckets = append(buckets, YearBucket{
			Year:   y,
			Count:  a.count,
			SizeGB: toGB(a.size),
			Href:   fmt.Sprintf("/movie-agg/birth/%d", y),
			Label:  fmt.Sprintf("%d 年", y),
		})
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Year > buckets[j].Year })

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

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       "电影聚合 · 拍摄日",
		"Buckets":     buckets,
		"Movies":      listResp.List,
		"Total":       listResp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"CurrentSort": curOD,
	})
}

// /movie-agg/birth/:year
func (h *MovieAggHTMLHandler) MovieAggBirthMonths(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := atoiDef(c.DefaultQuery("ps", fmt.Sprint(defaultAggPageSize)), defaultAggPageSize)

	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	rReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: fmt.Sprintf("%04d-01-01", year),
		FilmBirthTimeEnd:   fmt.Sprintf("%04d-12-31", year),
		Page:               1,
		PageSize:           999999,
	}
	rResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), rReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	type agg struct {
		count int
		size  int64
	}
	mm := map[int]*agg{}
	for _, m := range rResp.List {
		t, ok := parseYMD(m.FilmBirthDate)
		if !ok {
			continue
		}
		a := mm[int(t.Month())]
		if a == nil {
			a = &agg{}
			mm[int(t.Month())] = a
		}
		a.count++
		if m.VFilm != nil {
			a.size += m.VFilm.Size
		}
	}
	mb := make([]MonthBucket, 0, len(mm))
	for mth, a := range mm {
		mb = append(mb, MonthBucket{
			Month:  mth,
			Count:  a.count,
			SizeGB: toGB(a.size),
			Href:   fmt.Sprintf("/movie-agg/birth/%d/%02d", year, mth),
			Label:  fmt.Sprintf("%02d 月", mth),
		})
	}
	sort.Slice(mb, func(i, j int) bool { return mb[i].Month < mb[j].Month })

	listReq := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: fmt.Sprintf("%04d-01-01", year),
		FilmBirthTimeEnd:   fmt.Sprintf("%04d-12-31", year),
		OrderBy:            curOD,
		Page:               int64(page),
		PageSize:           int64(size),
	}
	listResp, err := h.movieSvc.ListMovieFull(c.Request.Context(), listReq)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, listResp.Total, int64(page), int64(size), pageWindow)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       fmt.Sprintf("拍摄日 · %d 年", year),
		"Buckets":     mb,
		"Movies":      listResp.List,
		"Total":       listResp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"CurrentSort": curOD,
	})
}

// /movie-agg/birth/:year/:month
func (h *MovieAggHTMLHandler) MovieAggBirthMonth(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	page := atoiDef(c.DefaultQuery("p", "1"), 1)
	size := atoiDef(c.DefaultQuery("ps", fmt.Sprint(defaultAggPageSize)), defaultAggPageSize)

	curOD := normalizeOrderBy(c.DefaultQuery("od", consts.OrderByBirthTime), consts.OrderByBirthTime)
	sq := buildSortQuery(c, curOD)

	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := fmt.Sprintf("%04d-%02d-31", year, month)

	req := &types.ListMovieFullRequest{
		Owned:              consts.OwnedAllNotRemoved,
		FilmBirthTimeStart: start,
		FilmBirthTimeEnd:   end,
		OrderBy:            curOD,
		Page:               int64(page),
		PageSize:           int64(size),
	}
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), req)
	if err != nil {
		c.String(500, "加载失败: %v", err)
		return
	}
	pi := BuildPageInfo(c, resp.Total, int64(page), int64(size), pageWindow)

	c.HTML(200, "page.movie_agg_time", gin.H{
		"Title":       fmt.Sprintf("拍摄日 · %d 年 %02d 月", year, month),
		"Movies":      resp.List,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"sortQuery":   sq,
		"CurrentSort": curOD,
	})
}

/* ------------------------- 工具函数 ------------------------- */

func atoiDef(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
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

// 不用模板函数，直接在后端把字节转成 "x.y GB" 字符串
func toGB(bytes int64) string {
	const gb = 1024.0 * 1024.0 * 1024.0
	return fmt.Sprintf("%.1f GB", float64(bytes)/gb)
}
