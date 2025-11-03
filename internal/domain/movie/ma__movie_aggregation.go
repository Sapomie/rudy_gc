package movie

import (
	"context"
	"fmt"
	"sort"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

const (
	minYear = 1900
	maxYear = 2100

	orderByRelease = "releasing_date"
	orderByBirth   = "birth_time"

	gbDiv = 1024.0 * 1024.0 * 1024.0
)

type aggLevel int

const (
	levelRoot aggLevel = iota
	levelYear
	levelQuarter
	levelMonth
)

// ---------- 业务主流程：Owned 聚合 ----------

func (s *MovieService) BuildOwnedAggView(ctx context.Context, p AggParams) (*AggResult, error) {
	level := detectLevel(p.Year, p.Quarter, p.Month)

	req := &types.ListMovieFullRequest{
		Owned:    consts.OwnedAllNotRemoved,
		OrderBy:  p.OrderBy,
		Page:     int64(p.Page),
		PageSize: int64(p.Size),
	}

	// 时间范围
	var start, end string
	switch level {
	case levelRoot:
	case levelYear:
		start = fmt.Sprintf("%04d-01-01", p.Year)
		end = fmt.Sprintf("%04d-12-31", p.Year)
	case levelQuarter:
		start, end = quarterRange(p.Year, p.Quarter)
	case levelMonth:
		start = fmt.Sprintf("%04d-%02d-01", p.Year, p.Month)
		end = lastDayOfMonth(p.Year, p.Month).Format("2006-01-02")
	}
	if p.Mode == "birth" {
		req.FilmBirthTimeStart, req.FilmBirthTimeEnd = start, end
	} else {
		req.ReleasingDateStart, req.ReleasingDateEnd = start, end
	}

	// 全量用于聚合与 Top
	fullReq := *req
	fullReq.Page, fullReq.PageSize = 1, 999999
	fullResp, err := s.ListMovieFull(ctx, &fullReq)
	if err != nil {
		return nil, err
	}

	// 列表分页
	listResp, err := s.ListMovieFull(ctx, req)
	if err != nil {
		return nil, err
	}

	dateFn := func(m *types.MovieType) string {
		if p.Mode == "birth" {
			return m.FilmBirthDate
		}
		return m.ReleasingDate
	}

	res := &AggResult{
		Breadcrumbs: buildAggBreadcrumbs(p.Mode, p.Year, p.Quarter, p.Month),
		Movies:      listResp.List,
		Total:       listResp.Total,
		RangeStart:  start,
		RangeEnd:    end,
	}

	// 标题 & Level
	switch level {
	case levelRoot:
		res.Level = "root"
		res.Title = "电影聚合 · " + modeTitle(p.Mode)
	case levelYear:
		res.Level = "year"
		res.Title = fmt.Sprintf("%s · %d 年", modeTitle(p.Mode), p.Year)
	case levelQuarter:
		res.Level = "quarter"
		res.Title = fmt.Sprintf("%s · %d 年 Q%d", modeTitle(p.Mode), p.Year, p.Quarter)
	case levelMonth:
		res.Level = "month"
		res.Title = fmt.Sprintf("%s · %d 年 %02d 月", modeTitle(p.Mode), p.Year, p.Month)
	}

	// Buckets
	switch level {
	case levelRoot:
		res.BucketsY = aggYearsOwned(fullResp.List, dateFn, p.Mode)
	case levelYear:
		res.BucketsQ = aggQuartersOwned(fullResp.List, dateFn, p.Year, p.Mode)
		// ★ 年页同时给全年 12 月
		res.BucketsM = aggMonthsWholeYearOwned(fullResp.List, dateFn, p.Year, p.Mode)
	case levelQuarter:
		res.BucketsM = aggMonthsOwned(fullResp.List, dateFn, p.Year, p.Quarter, p.Mode)
	}

	// Top（Owned）
	res.TopCasts = buildTopCasts(fullResp.List, p.TopN)
	res.TopDirectors = buildTopDirectors(fullResp.List, p.TopN)
	res.TopLabels = buildTopLabels(fullResp.List, p.TopN)
	res.TopPrefixes = buildTopPrefixes(fullResp.List, p.TopN)

	return res, nil
}

// ---------- 业务主流程：All/Release 聚合（模板 page.movie_agg_all_time） ----------

func (s *MovieService) BuildAllReleaseAggView(ctx context.Context, p AggParams) (*AggResult, error) {
	// 仅处理 release 全量页
	level := detectLevel(p.Year, p.Quarter, p.Month)

	req := &types.ListMovieFullRequest{
		Owned:    consts.MovieAll,
		OrderBy:  p.OrderBy,
		Page:     int64(p.Page),
		PageSize: int64(p.Size),
	}

	var start, end string
	switch level {
	case levelRoot:
	case levelYear:
		start = fmt.Sprintf("%04d-01-01", p.Year)
		end = fmt.Sprintf("%04d-12-31", p.Year)
	case levelQuarter:
		start, end = quarterRange(p.Year, p.Quarter)
	case levelMonth:
		start = fmt.Sprintf("%04d-%02d-01", p.Year, p.Month)
		end = lastDayOfMonth(p.Year, p.Month).Format("2006-01-02")
	}
	req.ReleasingDateStart, req.ReleasingDateEnd = start, end

	// 根页需要“年份入口”但你原逻辑用的是 owned 集合探测入口
	if level == levelRoot {
		alt := types.ListMovieFullRequest{
			Owned:    consts.OwnedAllNotRemoved,
			OrderBy:  orderByRelease,
			Page:     1,
			PageSize: 999999,
		}
		altResp, err := s.ListMovieFull(ctx, &alt)
		if err != nil {
			return nil, err
		}
		root := "/movie-agg-all/release"
		dateFn := func(m *types.MovieType) string { return m.ReleasingDate }
		buckets := aggYearsAll(altResp.List, dateFn, root)
		// 作为入口时抹掉 count/sizeAll 以免误导（保持你原先做法）
		for i := range buckets {
			buckets[i].CountAll = 0
			buckets[i].SizeAllGB = 0
		}
		return &AggResult{
			Title:       "电影聚合（全量）· 上映日",
			Breadcrumbs: buildAllReleaseBreadcrumbs(0, 0, 0),
			Level:       "root",
			BucketsAll:  buckets,
			Movies:      nil,
			Total:       0,
		}, nil
	}

	// 全量用于聚合/Top
	fullReq := *req
	fullReq.Page, fullReq.PageSize = 1, 999999
	fullResp, err := s.ListMovieFull(ctx, &fullReq)
	if err != nil {
		return nil, err
	}

	// 列表
	listResp, err := s.ListMovieFull(ctx, req)
	if err != nil {
		return nil, err
	}

	root := "/movie-agg-all/release"
	dateFn := func(m *types.MovieType) string { return m.ReleasingDate }

	res := &AggResult{
		Breadcrumbs: buildAllReleaseBreadcrumbs(p.Year, p.Quarter, p.Month),
		Movies:      listResp.List,
		Total:       listResp.Total,
		RangeStart:  start,
		RangeEnd:    end,
	}

	switch level {
	case levelYear:
		res.Level = "year"
		res.Title = fmt.Sprintf("上映日（全量）· %d 年", p.Year)
		res.BucketsQAll = aggQuartersAll(fullResp.List, dateFn, p.Year, root)
		res.BucketsMAll = aggMonthsYearAll(fullResp.List, dateFn, p.Year, root)
	case levelQuarter:
		res.Level = "quarter"
		res.Title = fmt.Sprintf("上映日（全量）· %d 年 Q%d", p.Year, p.Quarter)
		res.BucketsAll = aggMonthsAll(fullResp.List, dateFn, p.Year, p.Quarter, root)
	case levelMonth:
		res.Level = "month"
		res.Title = fmt.Sprintf("上映日（全量）· %d 年 %02d 月", p.Year, p.Month)
	}

	// Top（All）
	res.TopCastsAll = buildTopCastsAll(fullResp.List, p.TopN)
	res.TopDirectorsAll = buildTopDirectorsAll(fullResp.List, p.TopN)
	res.TopLabelsAll = buildTopLabelsAll(fullResp.List, p.TopN)
	res.TopPrefixesAll = buildTopPrefixesAll(fullResp.List, p.TopN)

	return res, nil
}

// ---------- 公共小工具 ----------

func bytesToGB(b int64) float64 { return float64(b) / gbDiv }

func parseYMD(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	// 用 UTC 解析以避免本地时区影响
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil || t.IsZero() {
		return time.Time{}, false
	}
	return t, true
}

func monthToQuarter(m int) int { return ((m - 1) / 3) + 1 }

func lastDayOfMonth(year, month int) time.Time {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
}

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

func rootPath(mode string) string {
	if mode == "birth" {
		return "/movie-agg-owned/birth"
	}
	return "/movie-agg-owned/release"
}
func modeTitle(mode string) string {
	if mode == "birth" {
		return "下载日"
	}
	return "上映日"
}

func buildAggBreadcrumbs(mode string, year, quarter, month int) []Breadcrumb {
	rootTitle, rootHref := modeTitle(mode), rootPath(mode)
	bcs := []Breadcrumb{{Title: rootTitle, Href: rootHref}}
	if year == 0 {
		bcs[0].Href = ""
		return bcs
	}
	yearHref := fmt.Sprintf("%s/%d", rootHref, year)
	if quarter == 0 && month == 0 {
		return append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: ""})
	}
	if month == 0 && quarter > 0 {
		return append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: ""},
		)
	}
	if quarter == 0 && month > 0 {
		quarter = monthToQuarter(month)
	}
	qHref := fmt.Sprintf("%s/%d/q/%d", rootHref, year, quarter)
	return append(bcs,
		Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
		Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: qHref},
		Breadcrumb{Title: fmt.Sprintf("%02d 月", month), Href: ""},
	)
}

func detectLevel(year, quarter, month int) aggLevel {
	switch {
	case year == 0:
		return levelRoot
	case quarter == 0 && month == 0:
		return levelYear
	case month == 0:
		return levelQuarter
	default:
		return levelMonth
	}
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

// ---------- Top（Owned） ----------

func buildTopCasts(movies []*types.MovieType, topN int) []CastStat {
	if topN <= 0 {
		topN = 20
	}
	type agg struct {
		n  int
		sc int64
	}
	mp := make(map[string]agg, 1024)
	for _, m := range movies {
		if m == nil || len(m.Cast) == 0 {
			continue
		}
		for _, c := range m.Cast {
			if c == nil || c.Name == "" {
				continue
			}
			a := mp[c.Name]
			a.n++
			a.sc += m.ScTimes
			mp[c.Name] = a
		}
	}
	out := make([]CastStat, 0, len(mp))
	for name, a := range mp {
		out = append(out, CastStat{Name: name, Count: a.n, ScSum: a.sc})
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
		n  int
		sc int64
	}
	mp := make(map[string]agg, 1024)
	for _, m := range movies {
		k := pick(m)
		if k == "" || k == "nil" {
			continue
		}
		a := mp[k]
		a.n++
		a.sc += m.ScTimes
		mp[k] = a
	}
	out := make([]TopStat, 0, len(mp))
	for name, a := range mp {
		out = append(out, TopStat{Name: name, Count: a.n, ScSum: a.sc})
	}
	sortTopStats(out)
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}
func buildTopDirectors(movies []*types.MovieType, topN int) []TopStat {
	return buildTopByStringField(movies, func(m *types.MovieType) string { return m.Director }, topN)
}
func buildTopLabels(movies []*types.MovieType, topN int) []TopStat {
	return buildTopByStringField(movies, func(m *types.MovieType) string { return m.Label }, topN)
}
func buildTopPrefixes(movies []*types.MovieType, topN int) []TopStat {
	return buildTopByStringField(movies, func(m *types.MovieType) string { return m.Prefix }, topN)
}

// ---------- Buckets（Owned） ----------

func aggYearsOwned(movies []*types.MovieType, dateFn func(*types.MovieType) string, mode string) []Bucket {
	const span = maxYear - minYear + 1
	cnt := make([]int, span)
	sz := make([]int64, span)
	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok {
			y := t.Year()
			if y >= minYear && y <= maxYear {
				i := y - minYear
				cnt[i]++
				if m.VFilm != nil {
					sz[i] += m.VFilm.Size
				}
			}
		}
	}
	root := rootPath(mode)
	out := make([]Bucket, 0, 60)
	for i := len(cnt) - 1; i >= 0; i-- {
		if cnt[i] == 0 {
			continue
		}
		y := minYear + i
		out = append(out, Bucket{
			Label:  fmt.Sprintf("%d 年", y),
			Href:   fmt.Sprintf("%s/%d", root, y),
			Count:  cnt[i],
			SizeGB: bytesToGB(sz[i]),
		})
	}
	return out
}

func aggQuartersOwned(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, mode string) []Bucket {
	cnt := make([]int, 5)
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
	root := rootPath(mode)
	out := make([]Bucket, 0, 4)
	for q := 1; q <= 4; q++ {
		if cnt[q] == 0 {
			continue
		}
		out = append(out, Bucket{
			Label:  fmt.Sprintf("Q%d 季", q),
			Href:   fmt.Sprintf("%s/%d/q/%d", root, year, q),
			Count:  cnt[q],
			SizeGB: bytesToGB(sz[q]),
		})
	}
	return out
}

func aggMonthsOwned(movies []*types.MovieType, dateFn func(*types.MovieType) string, year, quarter int, mode string) []Bucket {
	startMonth := 3*(quarter-1) + 1
	cnt := make([]int, 3)
	sz := make([]int64, 3)
	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok && t.Year() == year {
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
	root := rootPath(mode)
	out := make([]Bucket, 0, 3)
	for i := 0; i < 3; i++ {
		if cnt[i] == 0 {
			continue
		}
		m := startMonth + i
		out = append(out, Bucket{
			Label:  fmt.Sprintf("%02d 月", m),
			Href:   fmt.Sprintf("%s/%d/%02d", root, year, m),
			Count:  cnt[i],
			SizeGB: bytesToGB(sz[i]),
		})
	}
	return out
}

func aggMonthsWholeYearOwned(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, mode string) []Bucket {
	var cnt [13]int
	var sz [13]int64
	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok && t.Year() == year {
			mm := int(t.Month())
			cnt[mm]++
			if m.VFilm != nil {
				sz[mm] += m.VFilm.Size
			}
		}
	}
	root := rootPath(mode)
	out := make([]Bucket, 0, 12)
	for mm := 1; mm <= 12; mm++ {
		if cnt[mm] == 0 {
			continue
		}
		out = append(out, Bucket{
			Label:  fmt.Sprintf("%02d 月", mm),
			Href:   fmt.Sprintf("%s/%d/%02d", root, year, mm),
			Count:  cnt[mm],
			SizeGB: bytesToGB(sz[mm]),
		})
	}
	return out
}

// ---------- Top & Buckets（All/Release） ----------

func buildAllReleaseBreadcrumbs(year, quarter, month int) []Breadcrumb {
	rootTitle, rootHref := "上映日（全量）", "/movie-agg-all/release"
	bcs := []Breadcrumb{{Title: rootTitle, Href: rootHref}}
	if year == 0 {
		bcs[0].Href = ""
		return bcs
	}
	yearHref := fmt.Sprintf("%s/%d", rootHref, year)
	if quarter == 0 && month == 0 {
		return append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: ""})
	}
	if month == 0 && quarter > 0 {
		return append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: ""},
		)
	}
	if quarter == 0 && month > 0 {
		quarter = monthToQuarter(month)
	}
	qHref := fmt.Sprintf("%s/%d/q/%d", rootHref, year, quarter)
	return append(bcs,
		Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
		Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: qHref},
		Breadcrumb{Title: fmt.Sprintf("%02d 月", month), Href: ""},
	)
}

func aggYearsAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, root string) []AllBucket {
	const span = maxYear - minYear + 1
	var cntAll, cntOwned [span]int
	var szAll [span]int64
	for _, m := range movies {
		if t, ok := parseYMD(dateFn(m)); ok {
			y := t.Year()
			if y < minYear || y > maxYear {
				continue
			}
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
	out := make([]AllBucket, 0, 60)
	for i := span - 1; i >= 0; i-- {
		if cntAll[i] == 0 {
			continue
		}
		y := minYear + i
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%d 年", y),
			Href:       fmt.Sprintf("%s/%d", root, y),
			CountAll:   cntAll[i],
			CountOwned: cntOwned[i],
			SizeAllGB:  bytesToGB(szAll[i]),
		})
	}
	return out
}
func aggQuartersAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, root string) []AllBucket {
	var cntAll, cntOwned [5]int
	var szAll [5]int64
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
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("Q%d 季", q),
			Href:       fmt.Sprintf("%s/%d/q/%d", root, year, q),
			CountAll:   cntAll[q],
			CountOwned: cntOwned[q],
			SizeAllGB:  bytesToGB(szAll[q]),
		})
	}
	return out
}
func aggMonthsAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year, quarter int, root string) []AllBucket {
	startMonth := 3*(quarter-1) + 1
	var cntAll, cntOwned [3]int
	var szAll [3]int64
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
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%02d 月", mm),
			Href:       fmt.Sprintf("%s/%d/%02d", root, year, mm),
			CountAll:   cntAll[i],
			CountOwned: cntOwned[i],
			SizeAllGB:  bytesToGB(szAll[i]),
		})
	}
	return out
}
func aggMonthsYearAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, root string) []AllBucket {
	var cntAll, cntOwned [13]int
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
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%02d 月", mm),
			Href:       fmt.Sprintf("%s/%d/%02d", root, year, mm),
			CountAll:   cntAll[mm],
			CountOwned: cntOwned[mm],
			SizeAllGB:  bytesToGB(szAll[mm]),
		})
	}
	return out
}

// Top（All）
func buildTopCastsAll(movies []*types.MovieType, topN int) []TopStatAll {
	if topN <= 0 {
		topN = 20
	}
	type agg struct{ all, own int }
	mp := make(map[string]agg, 1024)
	for _, m := range movies {
		if m == nil || len(m.Cast) == 0 {
			continue
		}
		own := isOwned(m)
		for _, c := range m.Cast {
			if c == nil || c.Name == "" {
				continue
			}
			a := mp[c.Name]
			a.all++
			if own {
				a.own++
			}
			mp[c.Name] = a
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
	mp := make(map[string]agg, 1024)
	for _, m := range movies {
		k := pick(m)
		if k == "" || k == "nil" {
			continue
		}
		a := mp[k]
		a.all++
		if isOwned(m) {
			a.own++
		}
		mp[k] = a
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
