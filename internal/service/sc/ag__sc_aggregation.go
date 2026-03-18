package sc

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

const scEventNameLayout = "2006-01-02-15-04"

type AggParams struct {
	Year    int
	Quarter int
	Month   int
	TopN    int
}

type scAggCount struct {
	events map[string]struct{}
	movies int
}

func (s *ScService) BuildAggView(ctx context.Context, p AggParams) (*types.ScAggResult, error) {
	level := detectScAggLevel(p.Year, p.Quarter, p.Month)
	topN := p.TopN
	if topN <= 0 {
		topN = 12
	}

	allEvents, err := s.scFindAll(ctx)
	if err != nil {
		return nil, err
	}
	allLists, err := s.glFindAll(ctx)
	if err != nil {
		return nil, err
	}

	filteredEvents := filterScEventsByScope(allEvents, p.Year, p.Quarter, p.Month, level)
	eventNames := collectScEventNames(filteredEvents)
	filteredLists := filterGListsByEventNames(allLists, eventNames)
	movieMap := s.loadMovieTypesByJavIDs(ctx, collectMovieJavIDs(filteredLists))

	res := &types.ScAggResult{
		Title:                 buildScAggTitle(level, p.Year, p.Quarter, p.Month),
		Level:                 level,
		Breadcrumbs:           buildScAggBreadcrumbs(level, p.Year, p.Quarter, p.Month),
		TotalEvents:           len(filteredEvents),
		TotalMovieAppearances: countMovieAppearances(filteredLists),
		TotalUniqueMovies:     countUniqueMovies(filteredLists),
		RecentTrend:           buildScAggTrend(allEvents),
		TopCasts:              buildScAggTopCasts(filteredLists, movieMap, topN),
		TopLabels:             buildScAggTopLabels(filteredLists, movieMap, topN),
		TopPrefixes:           buildScAggTopPrefixes(filteredLists, movieMap, topN),
	}

	switch level {
	case "root":
		res.BucketsY = buildScYearBuckets(allEvents)
	case "year":
		res.BucketsQ = buildScQuarterBuckets(filteredEvents, p.Year)
		res.BucketsM = buildScYearMonthBuckets(filteredEvents, p.Year)
	case "quarter":
		res.BucketsM = buildScQuarterMonthBuckets(filteredEvents, p.Year, p.Quarter)
	case "month":
		res.Events = buildScAggEventItems(filteredEvents)
		res.Movies = collectScopedMovies(filteredLists, movieMap)
	}

	return res, nil
}

func detectScAggLevel(year, quarter, month int) string {
	switch {
	case year > 0 && month > 0:
		return "month"
	case year > 0 && quarter > 0:
		return "quarter"
	case year > 0:
		return "year"
	default:
		return "root"
	}
}

func buildScAggTitle(level string, year, quarter, month int) string {
	switch level {
	case "year":
		return fmt.Sprintf("SC 统计 · %d 年", year)
	case "quarter":
		return fmt.Sprintf("SC 统计 · %d 年 Q%d", year, quarter)
	case "month":
		return fmt.Sprintf("SC 统计 · %d 年 %02d 月", year, month)
	default:
		return "SC 统计"
	}
}

func buildScAggBreadcrumbs(level string, year, quarter, month int) []types.ScAggBreadcrumb {
	items := []types.ScAggBreadcrumb{
		{Title: "SC 统计", Href: "/sc-agg"},
	}

	switch level {
	case "year":
		items = append(items, types.ScAggBreadcrumb{Title: fmt.Sprintf("%d 年", year)})
	case "quarter":
		items = append(items,
			types.ScAggBreadcrumb{Title: fmt.Sprintf("%d 年", year), Href: fmt.Sprintf("/sc-agg/%d", year)},
			types.ScAggBreadcrumb{Title: fmt.Sprintf("Q%d", quarter)},
		)
	case "month":
		items = append(items,
			types.ScAggBreadcrumb{Title: fmt.Sprintf("%d 年", year), Href: fmt.Sprintf("/sc-agg/%d", year)},
			types.ScAggBreadcrumb{Title: fmt.Sprintf("Q%d", quarterOfMonth(month)), Href: fmt.Sprintf("/sc-agg/%d/q/%d", year, quarterOfMonth(month))},
			types.ScAggBreadcrumb{Title: fmt.Sprintf("%02d 月", month)},
		)
	default:
		items[0].Href = ""
	}

	return items
}

func filterScEventsByScope(events []*types.GSc, year, quarter, month int, level string) []*types.GSc {
	out := make([]*types.GSc, 0, len(events))
	for _, event := range events {
		t, ok := scEventTime(event)
		if !ok {
			continue
		}
		switch level {
		case "year":
			if t.Year() != year {
				continue
			}
		case "quarter":
			if t.Year() != year || quarterOfMonth(int(t.Month())) != quarter {
				continue
			}
		case "month":
			if t.Year() != year || int(t.Month()) != month {
				continue
			}
		}
		out = append(out, event)
	}
	return out
}

func buildScYearBuckets(events []*types.GSc) []types.ScAggBucket {
	grouped := make(map[int][]*types.GSc)
	years := make([]int, 0)
	seen := make(map[int]struct{})
	for _, event := range events {
		t, ok := scEventTime(event)
		if !ok {
			continue
		}
		year := t.Year()
		grouped[year] = append(grouped[year], event)
		if _, ok := seen[year]; ok {
			continue
		}
		seen[year] = struct{}{}
		years = append(years, year)
	}
	sort.Slice(years, func(i, j int) bool { return years[i] > years[j] })

	buckets := make([]types.ScAggBucket, 0, len(years))
	for _, year := range years {
		eventCount, cooldownDays, comeRate := summarizeScEvents(grouped[year])
		buckets = append(buckets, types.ScAggBucket{
			Label:           fmt.Sprintf("%d 年", year),
			Href:            fmt.Sprintf("/sc-agg/%d", year),
			EventCount:      eventCount,
			AvgCooldownDays: cooldownDays,
			ComeRate:        comeRate,
		})
	}
	return buckets
}

func buildScQuarterBuckets(events []*types.GSc, year int) []types.ScAggBucket {
	grouped := make(map[int][]*types.GSc)
	for _, event := range events {
		t, ok := scEventTime(event)
		if !ok || t.Year() != year {
			continue
		}
		q := quarterOfMonth(int(t.Month()))
		grouped[q] = append(grouped[q], event)
	}

	buckets := make([]types.ScAggBucket, 0, 4)
	for q := 1; q <= 4; q++ {
		eventCount, cooldownDays, comeRate := summarizeScEvents(grouped[q])
		buckets = append(buckets, types.ScAggBucket{
			Label:           fmt.Sprintf("Q%d", q),
			Href:            fmt.Sprintf("/sc-agg/%d/q/%d", year, q),
			EventCount:      eventCount,
			AvgCooldownDays: cooldownDays,
			ComeRate:        comeRate,
		})
	}
	return buckets
}

func buildScYearMonthBuckets(events []*types.GSc, year int) []types.ScAggBucket {
	grouped := make(map[int][]*types.GSc)
	for _, event := range events {
		t, ok := scEventTime(event)
		if !ok || t.Year() != year {
			continue
		}
		m := int(t.Month())
		grouped[m] = append(grouped[m], event)
	}

	buckets := make([]types.ScAggBucket, 0, 12)
	for month := 1; month <= 12; month++ {
		eventCount, cooldownDays, comeRate := summarizeScEvents(grouped[month])
		buckets = append(buckets, types.ScAggBucket{
			Label:           fmt.Sprintf("%02d 月", month),
			Href:            fmt.Sprintf("/sc-agg/%d/%02d", year, month),
			EventCount:      eventCount,
			AvgCooldownDays: cooldownDays,
			ComeRate:        comeRate,
		})
	}
	return buckets
}

func buildScQuarterMonthBuckets(events []*types.GSc, year, quarter int) []types.ScAggBucket {
	grouped := make(map[int][]*types.GSc)
	for _, event := range events {
		t, ok := scEventTime(event)
		if !ok || t.Year() != year || quarterOfMonth(int(t.Month())) != quarter {
			continue
		}
		m := int(t.Month())
		grouped[m] = append(grouped[m], event)
	}

	startMonth := (quarter-1)*3 + 1
	buckets := make([]types.ScAggBucket, 0, 3)
	for month := startMonth; month < startMonth+3; month++ {
		eventCount, cooldownDays, comeRate := summarizeScEvents(grouped[month])
		buckets = append(buckets, types.ScAggBucket{
			Label:           fmt.Sprintf("%02d 月", month),
			Href:            fmt.Sprintf("/sc-agg/%d/%02d", year, month),
			EventCount:      eventCount,
			AvgCooldownDays: cooldownDays,
			ComeRate:        comeRate,
		})
	}
	return buckets
}

func buildScAggTrend(events []*types.GSc) []types.ScAggTrend {
	now := time.Now().In(time.Local)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	grouped := make(map[string][]*types.GSc)
	for _, event := range events {
		t, ok := scEventTime(event)
		if !ok {
			continue
		}
		key := fmt.Sprintf("%04d-%02d", t.Year(), int(t.Month()))
		grouped[key] = append(grouped[key], event)
	}

	trends := make([]types.ScAggTrend, 0, 6)
	for i := 5; i >= 0; i-- {
		t := startOfMonth.AddDate(0, -i, 0)
		key := fmt.Sprintf("%04d-%02d", t.Year(), int(t.Month()))
		eventCount, cooldownDays, comeRate := summarizeScEvents(grouped[key])
		trends = append(trends, types.ScAggTrend{
			Label:           key,
			EventCount:      eventCount,
			AvgCooldownDays: cooldownDays,
			ComeRate:        comeRate,
		})
	}
	return trends
}

func buildScAggTopCasts(rows []*types.GList, movieMap map[string]*types.MovieType, limit int) []types.ScAggTopStat {
	stats := make(map[string]*scAggCount)
	for _, row := range rows {
		if row == nil || row.ScName == "" || row.MovieJavId == "" {
			continue
		}
		movieType := movieMap[row.MovieJavId]
		if movieType == nil {
			continue
		}
		seen := make(map[string]struct{})
		for _, cast := range movieType.Cast {
			name := normalizeCastName(cast)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			acc := ensureScAggCount(stats, name)
			acc.movies++
			acc.events[row.ScName] = struct{}{}
		}
	}
	return flattenScAggCounts(stats, limit, func(name string) string {
		return "/cast?name=" + url.QueryEscape(name)
	})
}

func buildScAggTopLabels(rows []*types.GList, movieMap map[string]*types.MovieType, limit int) []types.ScAggTopStat {
	stats := make(map[string]*scAggCount)
	for _, row := range rows {
		if row == nil || row.ScName == "" || row.MovieJavId == "" {
			continue
		}
		movieType := movieMap[row.MovieJavId]
		if movieType == nil {
			continue
		}
		name := strings.TrimSpace(movieType.Label)
		if name == "" {
			continue
		}
		acc := ensureScAggCount(stats, name)
		acc.movies++
		acc.events[row.ScName] = struct{}{}
	}
	return flattenScAggCounts(stats, limit, func(name string) string {
		return "/cards?ln=" + url.QueryEscape(name)
	})
}

func buildScAggTopPrefixes(rows []*types.GList, movieMap map[string]*types.MovieType, limit int) []types.ScAggTopStat {
	stats := make(map[string]*scAggCount)
	for _, row := range rows {
		if row == nil || row.ScName == "" || row.MovieJavId == "" {
			continue
		}
		movieType := movieMap[row.MovieJavId]
		if movieType == nil {
			continue
		}
		name := strings.TrimSpace(movieType.Prefix)
		if name == "" {
			continue
		}
		acc := ensureScAggCount(stats, name)
		acc.movies++
		acc.events[row.ScName] = struct{}{}
	}
	return flattenScAggCounts(stats, limit, func(name string) string {
		return "/cards?pn=" + url.QueryEscape(name)
	})
}

func flattenScAggCounts(stats map[string]*scAggCount, limit int, hrefFn func(string) string) []types.ScAggTopStat {
	out := make([]types.ScAggTopStat, 0, len(stats))
	for name, acc := range stats {
		out = append(out, types.ScAggTopStat{
			Name:       name,
			EventCount: len(acc.events),
			MovieCount: acc.movies,
			Href:       hrefFn(name),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].EventCount == out[j].EventCount {
			if out[i].MovieCount == out[j].MovieCount {
				return out[i].Name < out[j].Name
			}
			return out[i].MovieCount > out[j].MovieCount
		}
		return out[i].EventCount > out[j].EventCount
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func ensureScAggCount(stats map[string]*scAggCount, name string) *scAggCount {
	acc := stats[name]
	if acc != nil {
		return acc
	}
	acc = &scAggCount{events: make(map[string]struct{})}
	stats[name] = acc
	return acc
}

func buildScAggEventItems(events []*types.GSc) []*types.ScAggEventItem {
	items := make([]*types.ScAggEventItem, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}
		items = append(items, &types.ScAggEventItem{
			Event: event,
			Href:  buildScEventHref(event.Name),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Event.ScTime == items[j].Event.ScTime {
			return items[i].Event.Name > items[j].Event.Name
		}
		return items[i].Event.ScTime > items[j].Event.ScTime
	})
	return items
}

func collectScopedMovies(rows []*types.GList, movieMap map[string]*types.MovieType) []*types.MovieType {
	seen := make(map[string]struct{})
	movies := make([]*types.MovieType, 0)
	for _, row := range rows {
		if row == nil || row.MovieJavId == "" {
			continue
		}
		if _, ok := seen[row.MovieJavId]; ok {
			continue
		}
		seen[row.MovieJavId] = struct{}{}
		movieType := movieMap[row.MovieJavId]
		if movieType == nil {
			continue
		}
		movies = append(movies, movieType)
	}
	sortMoviesForScDisplay(movies)
	return movies
}

func (s *ScService) loadMovieTypesByJavIDs(ctx context.Context, javIDs []string) map[string]*types.MovieType {
	out := make(map[string]*types.MovieType, len(javIDs))
	for _, javID := range javIDs {
		if javID == "" {
			continue
		}
		movieType, err := s.movieSvc.GetMovieType(ctx, javID)
		if err != nil {
			if s.deps.Log != nil {
				s.deps.Log.WithError(err).Warnf("skip movie_type for sc aggregate: %s", javID)
			}
			continue
		}
		if movieType == nil {
			continue
		}
		out[javID] = movieType
	}
	return out
}

func summarizeScEvents(events []*types.GSc) (eventCount int, avgCooldownDays float64, comeRate float64) {
	if len(events) == 0 {
		return 0, 0, 0
	}

	var cooldownSum int64
	var cooldownCount int
	var comeCount int
	for _, event := range events {
		if event == nil {
			continue
		}
		if event.Cooldown > 0 {
			cooldownSum += event.Cooldown
			cooldownCount++
		}
		if strings.TrimSpace(event.ComeMovieName) != "" {
			comeCount++
		}
	}

	eventCount = len(events)
	if cooldownCount > 0 {
		avgCooldownDays = float64(cooldownSum) / 86400.0 / float64(cooldownCount)
	}
	comeRate = float64(comeCount) * 100.0 / float64(eventCount)
	return eventCount, avgCooldownDays, comeRate
}

func collectScEventNames(events []*types.GSc) map[string]struct{} {
	out := make(map[string]struct{}, len(events))
	for _, event := range events {
		if event == nil || event.Name == "" {
			continue
		}
		out[event.Name] = struct{}{}
	}
	return out
}

func filterGListsByEventNames(rows []*types.GList, names map[string]struct{}) []*types.GList {
	if len(names) == 0 {
		return nil
	}
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ScName == "" {
			continue
		}
		if _, ok := names[row.ScName]; !ok {
			continue
		}
		out = append(out, row)
	}
	return out
}

func countMovieAppearances(rows []*types.GList) int {
	total := 0
	for _, row := range rows {
		if row == nil || row.MovieJavId == "" {
			continue
		}
		total++
	}
	return total
}

func countUniqueMovies(rows []*types.GList) int {
	return len(collectMovieJavIDs(rows))
}

func collectMovieJavIDs(rows []*types.GList) []string {
	seen := make(map[string]struct{}, len(rows))
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.MovieJavId == "" {
			continue
		}
		if _, ok := seen[row.MovieJavId]; ok {
			continue
		}
		seen[row.MovieJavId] = struct{}{}
		out = append(out, row.MovieJavId)
	}
	return out
}

func scEventTime(event *types.GSc) (time.Time, bool) {
	if event == nil {
		return time.Time{}, false
	}
	if event.ScTime > 0 {
		return time.Unix(event.ScTime, 0).In(time.Local), true
	}
	if event.Name == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(scEventNameLayout, event.Name, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func quarterOfMonth(month int) int {
	return ((month - 1) / 3) + 1
}

func buildScEventHref(name string) string {
	return "/sc-events/" + url.PathEscape(name)
}

func normalizeCastName(cast *types.CastInfo) string {
	if cast == nil {
		return ""
	}
	name := strings.TrimSpace(cast.Name)
	if name != "" {
		return name
	}
	return strings.TrimSpace(cast.NameShow)
}

func actorMovieIsCome(gl *types.GList) bool {
	return gl != nil && gl.IsCome == consts.GListIsCome
}
