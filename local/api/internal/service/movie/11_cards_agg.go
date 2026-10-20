package movie

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/types"
)

const CardsAggMaxMovieCount int64 = 10000

type CardsAggView struct {
	Total         int64
	Skipped       bool
	SkipReason    string
	Message       string
	Limit         int64
	Summary       CardsAggSummary
	Groups        CardsAggGroups
	Timeline      CardsAggTimeline
	Distributions CardsAggDistributions
	TopMovies     CardsAggTopMovies
}

type CardsAggSummary struct {
	Owned []*CardsAggItem
	Sc    []*CardsAggItem
	Come  []*CardsAggItem
	Rank  []*CardsAggItem
}

type CardsAggGroups struct {
	Casts     []*CardsAggCastItem
	Directors []*CardsAggItem
	Prefixes  []*CardsAggItem
	Makers    []*CardsAggItem
	Labels    []*CardsAggItem
}

type CardsAggTimeline struct {
	ReleaseYears         []*CardsAggItem
	ReleaseMonths        []*CardsAggItem
	MediaBirthYears      []*CardsAggItem
	MediaBirthMonths     []*CardsAggItem
	LastSelectedScMonths []*CardsAggItem
}

type CardsAggDistributions struct {
	Scores       []*CardsAggItem
	Viewers      []*CardsAggItem
	Lengths      []*CardsAggItem
	ScTimes      []*CardsAggItem
	ComeTimes    []*CardsAggItem
	HighestRanks []*CardsAggItem
}

type CardsAggTopMovies struct {
	ScTimes          []*CardsAggMovieItem
	ComeTimes        []*CardsAggMovieItem
	Viewers          []*CardsAggMovieItem
	Scores           []*CardsAggMovieItem
	HighestRanks     []*CardsAggMovieItem
	RecentRelease    []*CardsAggMovieItem
	RecentMedia      []*CardsAggMovieItem
	RecentSelectedSc []*CardsAggMovieItem
}

type CardsAggItem struct {
	Key     string
	Label   string
	Count   int64
	Href    string
	Value   float64
	Percent float64
}

type CardsAggCastItem struct {
	Key            string
	Label          string
	Count          int64
	Href           string
	Percent        float64
	ScCount        int64
	NoScCount      int64
	MOwnedCount    int64
	MRemovedCount  int64
	MNotOwnedCount int64
}

type CardsAggMovieItem struct {
	JavID      string
	Name       string
	Title      string
	Href       string
	Value      float64
	ValueLabel string
	Date       string
}

func (s *Service) BuildMovieCardsAgg(ctx context.Context, req *types.ListMovieFullRequest) (*CardsAggView, error) {
	ids, err := s.movieListQuery().ListFullIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	total := int64(len(ids))
	if total > CardsAggMaxMovieCount {
		return &CardsAggView{
			Total:      total,
			Skipped:    true,
			SkipReason: "too_many_movies",
			Message:    fmt.Sprintf("筛选结果 %d 部，超过 %d 部，已跳过聚合。", total, CardsAggMaxMovieCount),
			Limit:      CardsAggMaxMovieCount,
		}, nil
	}

	movies := make([]*types.MovieType, 0, len(ids))
	for _, javID := range ids {
		mt, err := s.GetMovieType(ctx, javID)
		if err != nil {
			return nil, err
		}
		movies = append(movies, mt)
	}
	return buildCardsAggView(movies), nil
}

func buildCardsAggView(movies []*types.MovieType) *CardsAggView {
	total := int64(len(movies))
	owned := newCardsAggCounter(total)
	scPresence := newCardsAggCounter(total)
	comePresence := newCardsAggCounter(total)
	rankPresence := newCardsAggCounter(total)

	casts := newCardsAggCastCounter(total)
	directors := newCardsAggCounter(total)
	prefixes := newCardsAggCounter(total)
	makers := newCardsAggCounter(total)
	labels := newCardsAggCounter(total)

	releaseYears := newCardsAggCounter(total)
	releaseMonths := newCardsAggCounter(total)
	mediaBirthYears := newCardsAggCounter(total)
	mediaBirthMonths := newCardsAggCounter(total)
	lastSelectedScMonths := newCardsAggCounter(total)

	scores := newCardsAggCounter(total)
	viewers := newCardsAggCounter(total)
	lengths := newCardsAggCounter(total)
	scTimes := newCardsAggCounter(total)
	comeTimes := newCardsAggCounter(total)
	highestRanks := newCardsAggCounter(total)

	top := CardsAggTopMovies{}

	for _, mt := range movies {
		if mt == nil {
			continue
		}
		owned.add(cardsAggOwnedKey(mt.OwnedWMedia), cardsAggOwnedLabel(mt.OwnedWMedia), cardsAggOwnedHref(mt.OwnedWMedia), float64(mt.OwnedWMedia))
		if mt.ScTimes > 0 {
			scPresence.add("1", "有 SC", "/listcardagg?scmin=1", 1)
		} else {
			scPresence.add("0", "无 SC", "/listcardagg?scmax=0", 0)
		}
		if mt.ComeTimes > 0 {
			comePresence.add("1", "有 Come", "/listcardagg?comin=1", 1)
		} else {
			comePresence.add("0", "无 Come", "/listcardagg?comax=0", 0)
		}
		if mt.HighestRank > 0 {
			rankPresence.add("1", "有榜单", "/listcardrank", 1)
		} else {
			rankPresence.add("0", "无榜单", "", 0)
		}

		for _, cast := range mt.Cast {
			if cast == nil {
				continue
			}
			key := strings.TrimSpace(cast.Name)
			href := "/listcard?cn=" + url.QueryEscape(key)
			if cast.PersonId > 0 {
				key = "p:" + strconv.FormatInt(cast.PersonId, 10)
				href = "/castdetail/" + strconv.FormatInt(cast.PersonId, 10)
			}
			label := firstNonEmpty(cast.DisplayName, cast.Name)
			casts.add(key, label, href, mt)
		}
		directors.add(mt.Director, emptyLabel(mt.Director), "/listcard?dn="+url.QueryEscape(mt.Director), 0)
		prefixes.add(mt.Prefix, emptyLabel(mt.Prefix), "/listcard?pn="+url.QueryEscape(mt.Prefix), 0)
		makers.add(mt.Maker, emptyLabel(mt.Maker), "/listcard?mn="+url.QueryEscape(mt.Maker), 0)
		labels.add(mt.Label, emptyLabel(mt.Label), "/listcard?ln="+url.QueryEscape(mt.Label), 0)

		addDateAgg(releaseYears, releaseMonths, mt.ReleasingDate, "rs", "re")
		addDateAgg(mediaBirthYears, mediaBirthMonths, mt.FilmBirthDateWMedia, "mbs", "mbe")
		addUnixMonthAgg(lastSelectedScMonths, mt.LastSelectedScTime)

		scores.add(cardsAggFloatKey(mt.Score), cardsAggFloatLabel(mt.Score), "", mt.Score)
		viewers.add(strconv.FormatInt(mt.ViewersNumberWatched, 10), strconv.FormatInt(mt.ViewersNumberWatched, 10), "", float64(mt.ViewersNumberWatched))
		lengths.add(strconv.FormatInt(mt.Length, 10), strconv.FormatInt(mt.Length, 10), "", float64(mt.Length))
		scTimes.add(strconv.FormatInt(mt.ScTimes, 10), strconv.FormatInt(mt.ScTimes, 10), "", float64(mt.ScTimes))
		comeTimes.add(strconv.FormatInt(mt.ComeTimes, 10), strconv.FormatInt(mt.ComeTimes, 10), "", float64(mt.ComeTimes))
		highestRanks.add(cardsAggRankKey(mt.HighestRank), cardsAggRankLabel(mt.HighestRank), "", float64(mt.HighestRank))

		if mt.ScTimes > 0 {
			top.ScTimes = append(top.ScTimes, cardsAggMovieItem(mt, float64(mt.ScTimes), strconv.FormatInt(mt.ScTimes, 10), ""))
		}
		if mt.ComeTimes > 0 {
			top.ComeTimes = append(top.ComeTimes, cardsAggMovieItem(mt, float64(mt.ComeTimes), strconv.FormatInt(mt.ComeTimes, 10), ""))
		}
		if mt.ViewersNumberWatched > 0 {
			top.Viewers = append(top.Viewers, cardsAggMovieItem(mt, float64(mt.ViewersNumberWatched), strconv.FormatInt(mt.ViewersNumberWatched, 10), ""))
		}
		if mt.Score > 0 {
			top.Scores = append(top.Scores, cardsAggMovieItem(mt, mt.Score, cardsAggFloatLabel(mt.Score), ""))
		}
		if mt.HighestRank > 0 {
			top.HighestRanks = append(top.HighestRanks, cardsAggMovieItem(mt, float64(mt.HighestRank), "#"+strconv.FormatInt(mt.HighestRank, 10), ""))
		}
		if unix := parseCardsAggDate(mt.ReleasingDate); unix > 0 {
			top.RecentRelease = append(top.RecentRelease, cardsAggMovieItem(mt, float64(unix), mt.ReleasingDate, mt.ReleasingDate))
		}
		if unix := parseCardsAggDate(mt.FilmBirthDateWMedia); unix > 0 {
			top.RecentMedia = append(top.RecentMedia, cardsAggMovieItem(mt, float64(unix), mt.FilmBirthDateWMedia, mt.FilmBirthDateWMedia))
		}
		if mt.LastSelectedScTime > 0 {
			date := cardsAggUnixDate(mt.LastSelectedScTime)
			top.RecentSelectedSc = append(top.RecentSelectedSc, cardsAggMovieItem(mt, float64(mt.LastSelectedScTime), date, date))
		}
	}

	sortCardsAggMoviesDesc(top.ScTimes)
	sortCardsAggMoviesDesc(top.ComeTimes)
	sortCardsAggMoviesDesc(top.Viewers)
	sortCardsAggMoviesDesc(top.Scores)
	sortCardsAggMoviesAsc(top.HighestRanks)
	sortCardsAggMoviesDesc(top.RecentRelease)
	sortCardsAggMoviesDesc(top.RecentMedia)
	sortCardsAggMoviesDesc(top.RecentSelectedSc)

	return &CardsAggView{
		Total: total,
		Summary: CardsAggSummary{
			Owned: owned.itemsByCount(),
			Sc:    scPresence.itemsByValue(),
			Come:  comePresence.itemsByValue(),
			Rank:  rankPresence.itemsByValue(),
		},
		Groups: CardsAggGroups{
			Casts:     casts.itemsByCount(),
			Directors: directors.itemsByCount(),
			Prefixes:  prefixes.itemsByCount(),
			Makers:    makers.itemsByCount(),
			Labels:    labels.itemsByCount(),
		},
		Timeline: CardsAggTimeline{
			ReleaseYears:         releaseYears.itemsByKeyDesc(),
			ReleaseMonths:        releaseMonths.itemsByKeyDesc(),
			MediaBirthYears:      mediaBirthYears.itemsByKeyDesc(),
			MediaBirthMonths:     mediaBirthMonths.itemsByKeyDesc(),
			LastSelectedScMonths: lastSelectedScMonths.itemsByKeyDesc(),
		},
		Distributions: CardsAggDistributions{
			Scores:       scores.itemsByValueDesc(),
			Viewers:      viewers.itemsByValueDesc(),
			Lengths:      lengths.itemsByValueDesc(),
			ScTimes:      scTimes.itemsByValueDesc(),
			ComeTimes:    comeTimes.itemsByValueDesc(),
			HighestRanks: highestRanks.itemsByRank(),
		},
		TopMovies: top,
	}
}

type cardsAggCounter struct {
	total int64
	m     map[string]*CardsAggItem
}

func newCardsAggCounter(total int64) *cardsAggCounter {
	return &cardsAggCounter{total: total, m: make(map[string]*CardsAggItem)}
}

func (c *cardsAggCounter) add(key, label, href string, value float64) {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "(empty)"
	}
	label = emptyLabel(label)
	item, ok := c.m[key]
	if !ok {
		item = &CardsAggItem{
			Key:   key,
			Label: label,
			Href:  href,
			Value: value,
		}
		c.m[key] = item
	}
	item.Count++
	if c.total > 0 {
		item.Percent = round1(float64(item.Count) * 100 / float64(c.total))
	}
}

func (c *cardsAggCounter) items() []*CardsAggItem {
	out := make([]*CardsAggItem, 0, len(c.m))
	for _, item := range c.m {
		out = append(out, item)
	}
	return out
}

func (c *cardsAggCounter) itemsByCount() []*CardsAggItem {
	out := c.items()
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Label < out[j].Label
	})
	return out
}

func (c *cardsAggCounter) itemsByKeyDesc() []*CardsAggItem {
	out := c.items()
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Key != out[j].Key {
			return out[i].Key > out[j].Key
		}
		return out[i].Label < out[j].Label
	})
	return out
}

func (c *cardsAggCounter) itemsByValue() []*CardsAggItem {
	out := c.items()
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Value != out[j].Value {
			return out[i].Value > out[j].Value
		}
		return out[i].Label < out[j].Label
	})
	return out
}

func (c *cardsAggCounter) itemsByValueDesc() []*CardsAggItem {
	return c.itemsByValue()
}

func (c *cardsAggCounter) itemsByRank() []*CardsAggItem {
	out := c.items()
	sort.SliceStable(out, func(i, j int) bool {
		iv := out[i].Value
		jv := out[j].Value
		if iv == 0 && jv != 0 {
			return false
		}
		if iv != 0 && jv == 0 {
			return true
		}
		if iv != jv {
			return iv < jv
		}
		return out[i].Label < out[j].Label
	})
	return out
}

type cardsAggCastCounter struct {
	total int64
	m     map[string]*CardsAggCastItem
}

func newCardsAggCastCounter(total int64) *cardsAggCastCounter {
	return &cardsAggCastCounter{total: total, m: make(map[string]*CardsAggCastItem)}
}

func (c *cardsAggCastCounter) add(key, label, href string, mt *types.MovieType) {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "(empty)"
	}
	label = emptyLabel(label)
	item, ok := c.m[key]
	if !ok {
		item = &CardsAggCastItem{
			Key:   key,
			Label: label,
			Href:  href,
		}
		c.m[key] = item
	}
	item.Count++
	if c.total > 0 {
		item.Percent = round1(float64(item.Count) * 100 / float64(c.total))
	}
	if mt == nil {
		return
	}
	if mt.ScTimes > 0 {
		item.ScCount++
	} else {
		item.NoScCount++
	}
	switch mt.OwnedWMedia {
	case consts.OwnedHasSubNotRemoved, consts.OwnedNoSubNotRemoved:
		item.MOwnedCount++
	case consts.OwnedRemoved:
		item.MRemovedCount++
	default:
		item.MNotOwnedCount++
	}
}

func (c *cardsAggCastCounter) items() []*CardsAggCastItem {
	out := make([]*CardsAggCastItem, 0, len(c.m))
	for _, item := range c.m {
		out = append(out, item)
	}
	return out
}

func (c *cardsAggCastCounter) itemsByCount() []*CardsAggCastItem {
	out := c.items()
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Label < out[j].Label
	})
	return out
}

func cardsAggMovieItem(mt *types.MovieType, value float64, valueLabel, date string) *CardsAggMovieItem {
	return &CardsAggMovieItem{
		JavID:      mt.JavId,
		Name:       mt.Name,
		Title:      mt.Title,
		Href:       "/moviedetail/" + url.PathEscape(mt.Name),
		Value:      value,
		ValueLabel: valueLabel,
		Date:       date,
	}
}

func sortCardsAggMoviesDesc(items []*CardsAggMovieItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Value != items[j].Value {
			return items[i].Value > items[j].Value
		}
		return items[i].Name < items[j].Name
	})
}

func sortCardsAggMoviesAsc(items []*CardsAggMovieItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Value != items[j].Value {
			return items[i].Value < items[j].Value
		}
		return items[i].Name < items[j].Name
	})
}

func addDateAgg(years, months *cardsAggCounter, raw, startParam, endParam string) {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 4 {
		year := raw[:4]
		years.add(year, year, "/listcard?"+startParam+"="+url.QueryEscape(year+"-01-01")+"&"+endParam+"="+url.QueryEscape(year+"-12-31"), 0)
	} else {
		years.add("", "(空)", "", 0)
	}
	if len(raw) >= 7 {
		month := raw[:7]
		start, end := monthRange(month)
		months.add(month, month, "/listcard?"+startParam+"="+url.QueryEscape(start)+"&"+endParam+"="+url.QueryEscape(end), 0)
	} else {
		months.add("", "(空)", "", 0)
	}
}

func addUnixMonthAgg(months *cardsAggCounter, ts int64) {
	if ts <= 0 {
		months.add("", "(空)", "", 0)
		return
	}
	t := time.Unix(ts, 0)
	month := t.Format("2006-01")
	months.add(month, month, "/listcard?lsstmin="+url.QueryEscape(month+"-01")+"&lsstmax="+url.QueryEscape(monthRangeEnd(month)), 0)
}

func monthRange(month string) (string, string) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "", ""
	}
	start := t.Format(time.DateOnly)
	end := t.AddDate(0, 1, -1).Format(time.DateOnly)
	return start, end
}

func monthRangeEnd(month string) string {
	_, end := monthRange(month)
	return end
}

func parseCardsAggDate(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	t, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func cardsAggUnixDate(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format(time.DateOnly)
}

func cardsAggOwnedKey(v int64) string {
	if v == 0 {
		return "0"
	}
	return strconv.FormatInt(v, 10)
}

func cardsAggOwnedLabel(v int64) string {
	switch v {
	case consts.OwnedHasSubNotRemoved:
		return "有字幕未移除"
	case consts.OwnedNoSubNotRemoved:
		return "无字幕未移除"
	case consts.OwnedRemoved:
		return "已移除"
	case consts.OwnedAllNotRemoved:
		return "已拥有未移除"
	case consts.OwnedNotOwned, 0:
		return "无 w_media 记录"
	default:
		return "未知"
	}
}

func cardsAggOwnedHref(v int64) string {
	if v == 0 {
		v = consts.OwnedNotOwned
	}
	return "/listcardagg?mowned=" + strconv.FormatInt(v, 10)
}

func cardsAggFloatKey(v float64) string {
	return fmt.Sprintf("%.1f", v)
}

func cardsAggFloatLabel(v float64) string {
	return fmt.Sprintf("%.1f", v)
}

func cardsAggRankKey(v int64) string {
	if v <= 0 {
		return "0"
	}
	return strconv.FormatInt(v, 10)
}

func cardsAggRankLabel(v int64) string {
	if v <= 0 {
		return "未上榜"
	}
	return "#" + strconv.FormatInt(v, 10)
}

func emptyLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(空)"
	}
	return value
}
