package movie

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

const (
	minYear        = 1900
	maxYear        = 2100
	orderByRelease = "releasing_date"
	orderByBirth   = "birth_time"
	gbDiv          = 1024.0 * 1024.0 * 1024.0
)

type aggLevel int

const (
	levelRoot aggLevel = iota
	levelYear
	levelQuarter
	levelMonth
)

func (s *Service) BuildOwnedAggView(ctx context.Context, p AggParams) (*AggResult, error) {
	level := detectLevel(p.Year, p.Quarter, p.Month)
	req := &types.ListMovieFullRequest{
		MediaOwned: consts.OwnedAllNotRemoved,
		OrderBy:    p.OrderBy,
		Page:       int64(p.Page),
		PageSize:   int64(p.Size),
	}

	var start, end string
	switch level {
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
		req.MediaBirthTimeStart, req.MediaBirthTimeEnd = start, end
	} else {
		req.ReleasingDateStart, req.ReleasingDateEnd = start, end
	}

	fullReq := *req
	fullReq.Page, fullReq.PageSize = 1, 999999
	fullResp, err := s.ListMovieFull(ctx, &fullReq)
	if err != nil {
		return nil, err
	}
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

	switch level {
	case levelRoot:
		res.BucketsY = aggYearsOwned(fullResp.List, dateFn, p.Mode)
	case levelYear:
		res.BucketsQ = aggQuartersOwned(fullResp.List, dateFn, p.Year, p.Mode)
		res.BucketsM = aggMonthsWholeYearOwned(fullResp.List, dateFn, p.Year, p.Mode)
	case levelQuarter:
		res.BucketsM = aggMonthsOwned(fullResp.List, dateFn, p.Year, p.Quarter, p.Mode)
	}

	res.TopCasts = buildTopCasts(fullResp.List, p.TopN)
	res.TopDirectors = buildTopDirectors(fullResp.List, p.TopN)
	res.TopLabels = buildTopLabels(fullResp.List, p.TopN)
	res.TopPrefixes = buildTopPrefixes(fullResp.List, p.TopN)
	return res, nil
}

func (s *Service) BuildAllReleaseAggView(ctx context.Context, p AggParams) (*AggResult, error) {
	level := detectLevel(p.Year, p.Quarter, p.Month)
	req := &types.ListMovieFullRequest{
		MediaOwned: consts.MovieAll,
		OrderBy:    p.OrderBy,
		Page:       int64(p.Page),
		PageSize:   int64(p.Size),
	}

	var start, end string
	switch level {
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

	if level == levelRoot {
		alt := types.ListMovieFullRequest{
			MediaOwned: consts.OwnedAllNotRemoved,
			OrderBy:    orderByRelease,
			Page:       1,
			PageSize:   999999,
		}
		altResp, err := s.ListMovieFull(ctx, &alt)
		if err != nil {
			return nil, err
		}
		root := "/movie-agg-all/release"
		dateFn := func(m *types.MovieType) string { return m.ReleasingDate }
		buckets := aggYearsAll(altResp.List, dateFn, root)
		for i := range buckets {
			buckets[i].CountAll = 0
			buckets[i].SizeAllGB = 0
		}
		return &AggResult{
			Title:       "电影聚合（全量）· 上映日",
			Breadcrumbs: buildAllReleaseBreadcrumbs(0, 0, 0),
			Level:       "root",
			BucketsAll:  buckets,
		}, nil
	}

	fullReq := *req
	fullReq.Page, fullReq.PageSize = 1, 999999
	fullResp, err := s.ListMovieFull(ctx, &fullReq)
	if err != nil {
		return nil, err
	}
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

	res.TopCastsAll = buildTopCastsAll(fullResp.List, p.TopN)
	res.TopDirectorsAll = buildTopDirectorsAll(fullResp.List, p.TopN)
	res.TopLabelsAll = buildTopLabelsAll(fullResp.List, p.TopN)
	res.TopPrefixesAll = buildTopPrefixesAll(fullResp.List, p.TopN)
	return res, nil
}

func bytesToGB(b int64) float64 { return float64(b) / gbDiv }

func parseAggYMD(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
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
		return append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year)})
	}
	if month == 0 && quarter > 0 {
		return append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter)},
		)
	}
	if quarter == 0 && month > 0 {
		quarter = monthToQuarter(month)
	}
	qHref := fmt.Sprintf("%s/%d/q/%d", rootHref, year, quarter)
	return append(bcs,
		Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
		Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: qHref},
		Breadcrumb{Title: fmt.Sprintf("%02d 月", month)},
	)
}

func buildAllReleaseBreadcrumbs(year, quarter, month int) []Breadcrumb {
	rootHref := "/movie-agg-all/release"
	bcs := []Breadcrumb{{Title: "上映日（全量）", Href: rootHref}}
	if year == 0 {
		bcs[0].Href = ""
		return bcs
	}
	yearHref := fmt.Sprintf("%s/%d", rootHref, year)
	if quarter == 0 && month == 0 {
		return append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", year)})
	}
	if month == 0 && quarter > 0 {
		return append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter)},
		)
	}
	if quarter == 0 && month > 0 {
		quarter = monthToQuarter(month)
	}
	qHref := fmt.Sprintf("%s/%d/q/%d", rootHref, year, quarter)
	return append(bcs,
		Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
		Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: qHref},
		Breadcrumb{Title: fmt.Sprintf("%02d 月", month)},
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

func buildTopCasts(movies []*types.MovieType, topN int) []CastStat {
	if topN <= 0 {
		topN = 20
	}
	type agg struct {
		personID int64
		name     string
		n        int
		sc       int64
	}
	mp := make(map[string]*agg, 1024)
	for _, m := range movies {
		if m == nil || len(m.Cast) == 0 {
			continue
		}
		seen := make(map[string]struct{}, len(m.Cast))
		for _, c := range m.Cast {
			key, personID, name := movieAggCastKey(c)
			if key == "" || name == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			a := mp[key]
			if a == nil {
				a = &agg{personID: personID, name: name}
				mp[key] = a
			}
			a.n++
			a.sc += m.ScTimes
		}
	}
	out := make([]CastStat, 0, len(mp))
	for _, a := range mp {
		if a == nil || a.name == "" {
			continue
		}
		out = append(out, CastStat{PersonId: a.personID, Name: a.name, Count: a.n, ScSum: a.sc})
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

func aggYearsOwned(movies []*types.MovieType, dateFn func(*types.MovieType) string, mode string) []Bucket {
	const span = maxYear - minYear + 1
	cnt := make([]int, span)
	sz := make([]int64, span)
	for _, m := range movies {
		if t, ok := parseAggYMD(dateFn(m)); ok {
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
		if t, ok := parseAggYMD(dateFn(m)); ok && t.Year() == year {
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
		if t, ok := parseAggYMD(dateFn(m)); ok && t.Year() == year {
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
	cnt := make([]int, 13)
	sz := make([]int64, 13)
	for _, m := range movies {
		if t, ok := parseAggYMD(dateFn(m)); ok && t.Year() == year {
			mm := int(t.Month())
			cnt[mm]++
			if m.VFilm != nil {
				sz[mm] += m.VFilm.Size
			}
		}
	}
	root := rootPath(mode)
	out := make([]Bucket, 0, 12)
	for m := 1; m <= 12; m++ {
		if cnt[m] == 0 {
			continue
		}
		out = append(out, Bucket{
			Label:  fmt.Sprintf("%02d 月", m),
			Href:   fmt.Sprintf("%s/%d/%02d", root, year, m),
			Count:  cnt[m],
			SizeGB: bytesToGB(sz[m]),
		})
	}
	return out
}

func buildTopCastsAll(movies []*types.MovieType, topN int) []TopStatAll {
	if topN <= 0 {
		topN = 20
	}
	type agg struct {
		personID int64
		name     string
		all      int
		owned    int
	}
	mp := make(map[string]*agg, 1024)
	for _, m := range movies {
		if m == nil || len(m.Cast) == 0 {
			continue
		}
		seen := make(map[string]struct{}, len(m.Cast))
		for _, c := range m.Cast {
			key, personID, name := movieAggCastKey(c)
			if key == "" || name == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			a := mp[key]
			if a == nil {
				a = &agg{personID: personID, name: name}
				mp[key] = a
			}
			a.all++
			if isOwned(m) {
				a.owned++
			}
		}
	}
	out := make([]TopStatAll, 0, len(mp))
	for _, a := range mp {
		if a == nil || a.name == "" {
			continue
		}
		out = append(out, TopStatAll{PersonId: a.personID, Name: a.name, CountAll: a.all, CountOwned: a.owned})
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

func movieAggCastKey(c *types.CastInfo) (key string, personID int64, name string) {
	if c == nil {
		return "", 0, ""
	}
	name = strings.TrimSpace(c.DisplayName)
	if name == "" {
		name = strings.TrimSpace(c.Name)
	}
	if name == "" {
		name = strings.TrimSpace(c.NameShow)
	}
	if name == "" {
		return "", 0, ""
	}
	if c.PersonId > 0 {
		return "p:" + strconv.FormatInt(c.PersonId, 10), c.PersonId, name
	}
	return "n:" + name, 0, name
}

func buildTopByStringFieldAll(movies []*types.MovieType, pick func(*types.MovieType) string, topN int) []TopStatAll {
	if topN <= 0 {
		topN = 20
	}
	type agg struct {
		all   int
		owned int
	}
	mp := make(map[string]agg, 1024)
	for _, m := range movies {
		k := pick(m)
		if k == "" || k == "nil" {
			continue
		}
		a := mp[k]
		a.all++
		if isOwned(m) {
			a.owned++
		}
		mp[k] = a
	}
	out := make([]TopStatAll, 0, len(mp))
	for name, a := range mp {
		out = append(out, TopStatAll{Name: name, CountAll: a.all, CountOwned: a.owned})
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
	return buildTopByStringFieldAll(movies, func(m *types.MovieType) string { return m.Director }, topN)
}

func buildTopLabelsAll(movies []*types.MovieType, topN int) []TopStatAll {
	return buildTopByStringFieldAll(movies, func(m *types.MovieType) string { return m.Label }, topN)
}

func buildTopPrefixesAll(movies []*types.MovieType, topN int) []TopStatAll {
	return buildTopByStringFieldAll(movies, func(m *types.MovieType) string { return m.Prefix }, topN)
}

func aggYearsAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, root string) []AllBucket {
	type agg struct {
		all   int
		owned int
		size  int64
	}
	mp := make(map[int]agg, 64)
	for _, m := range movies {
		if t, ok := parseAggYMD(dateFn(m)); ok {
			y := t.Year()
			if y < minYear || y > maxYear {
				continue
			}
			a := mp[y]
			a.all++
			if isOwned(m) {
				a.owned++
			}
			if m.VFilm != nil {
				a.size += m.VFilm.Size
			}
			mp[y] = a
		}
	}
	out := make([]AllBucket, 0, len(mp))
	for y, a := range mp {
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%d 年", y),
			Href:       fmt.Sprintf("%s/%d", root, y),
			CountAll:   a.all,
			CountOwned: a.owned,
			SizeAllGB:  bytesToGB(a.size),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label > out[j].Label })
	return out
}

func aggQuartersAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, root string) []AllBucket {
	type agg struct {
		all   int
		owned int
		size  int64
	}
	mp := make(map[int]agg, 4)
	for _, m := range movies {
		if t, ok := parseAggYMD(dateFn(m)); ok && t.Year() == year {
			q := monthToQuarter(int(t.Month()))
			a := mp[q]
			a.all++
			if isOwned(m) {
				a.owned++
			}
			if m.VFilm != nil {
				a.size += m.VFilm.Size
			}
			mp[q] = a
		}
	}
	out := make([]AllBucket, 0, len(mp))
	for q, a := range mp {
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("Q%d 季", q),
			Href:       fmt.Sprintf("%s/%d/q/%d", root, year, q),
			CountAll:   a.all,
			CountOwned: a.owned,
			SizeAllGB:  bytesToGB(a.size),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

func aggMonthsAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year, quarter int, root string) []AllBucket {
	startMonth := 3*(quarter-1) + 1
	type agg struct {
		all   int
		owned int
		size  int64
	}
	mp := make(map[int]agg, 3)
	for _, m := range movies {
		if t, ok := parseAggYMD(dateFn(m)); ok && t.Year() == year {
			mm := int(t.Month())
			if mm < startMonth || mm > startMonth+2 {
				continue
			}
			a := mp[mm]
			a.all++
			if isOwned(m) {
				a.owned++
			}
			if m.VFilm != nil {
				a.size += m.VFilm.Size
			}
			mp[mm] = a
		}
	}
	out := make([]AllBucket, 0, len(mp))
	for mm, a := range mp {
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%02d 月", mm),
			Href:       fmt.Sprintf("%s/%d/%02d", root, year, mm),
			CountAll:   a.all,
			CountOwned: a.owned,
			SizeAllGB:  bytesToGB(a.size),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

func aggMonthsYearAll(movies []*types.MovieType, dateFn func(*types.MovieType) string, year int, root string) []AllBucket {
	type agg struct {
		all   int
		owned int
		size  int64
	}
	mp := make(map[int]agg, 12)
	for _, m := range movies {
		if t, ok := parseAggYMD(dateFn(m)); ok && t.Year() == year {
			mm := int(t.Month())
			a := mp[mm]
			a.all++
			if isOwned(m) {
				a.owned++
			}
			if m.VFilm != nil {
				a.size += m.VFilm.Size
			}
			mp[mm] = a
		}
	}
	out := make([]AllBucket, 0, len(mp))
	for mm, a := range mp {
		out = append(out, AllBucket{
			Label:      fmt.Sprintf("%02d 月", mm),
			Href:       fmt.Sprintf("%s/%d/%02d", root, year, mm),
			CountAll:   a.all,
			CountOwned: a.owned,
			SizeAllGB:  bytesToGB(a.size),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}
