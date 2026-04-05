package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/types"
)

type movieCardFilterView struct {
	Action    string
	ClearHref string

	OrderBy  string
	Order    string
	PageSize string

	RandomCount    string
	Explicit       map[string]bool
	HideDirs       bool
	TextDateInputs bool

	CastNames    string
	PersonIds    string
	GenreNames   string
	DirectorName string
	PrefixName   string
	MakerName    string
	LabelName    string
	Word         string

	ReleasingDateStart  string
	ReleasingDateEnd    string
	FilmBirthTimeStart  string
	FilmBirthTimeEnd    string
	MediaBirthTimeStart string
	MediaBirthTimeEnd   string

	CastAgeMin string
	CastAgeMax string

	StartRankingDateStart string
	StartRankingDateEnd   string

	DaysInRankMin string
	NeedDownload  string
	Owned         string
	MediaOwned    string

	ViewWatchedMin string
	ViewWatchedMax string
	ScoreMin       string
	ScoreMax       string

	LastScTimeMin string
	LastScTimeMax string
	ScTimesMin    string
	ScTimesMax    string
	ComeTimesMin  string
	ComeTimesMax  string

	Dir1 string
	Dir2 string
	Dir3 string
	Dir4 string
}

func buildMovieCardFilterView(c *gin.Context, req types.ListMovieFullRequest, currentOD string, randomN *int) *movieCardFilterView {
	path := c.Request.URL.Path

	view := &movieCardFilterView{
		Action:         path,
		ClearHref:      buildMovieCardFilterClearHref(path, randomN),
		OrderBy:        currentOD,
		Order:          queryOrFallbackString(c, "order", req.Order),
		PageSize:       strconv.FormatInt(req.PageSize, 10),
		Explicit:       buildMovieCardFilterExplicit(c, randomN != nil),
		HideDirs:       false,
		TextDateInputs: true,

		CastNames:    queryOrFallbackString(c, "cn", req.CastNames),
		PersonIds:    queryOrFallbackString(c, "pid", req.PersonIds),
		GenreNames:   queryOrFallbackString(c, "gn", req.GenreNames),
		DirectorName: queryOrFallbackString(c, "dn", req.DirectorName),
		PrefixName:   queryOrFallbackString(c, "pn", req.PrefixName),
		MakerName:    queryOrFallbackString(c, "mn", req.MakerName),
		LabelName:    queryOrFallbackString(c, "ln", req.LabelName),
		Word:         queryOrFallbackString(c, "wd", req.Word),

		ReleasingDateStart:  queryOrFallbackString(c, "rs", req.ReleasingDateStart),
		ReleasingDateEnd:    queryOrFallbackString(c, "re", req.ReleasingDateEnd),
		FilmBirthTimeStart:  queryOrFallbackString(c, "bs", req.FilmBirthTimeStart),
		FilmBirthTimeEnd:    queryOrFallbackString(c, "be", req.FilmBirthTimeEnd),
		MediaBirthTimeStart: queryOrFallbackString(c, "mbs", req.MediaBirthTimeStart),
		MediaBirthTimeEnd:   queryOrFallbackString(c, "mbe", req.MediaBirthTimeEnd),

		CastAgeMin: queryOrFallbackFloat(c, "cay", req.CastAgeMin),
		CastAgeMax: queryOrFallbackFloat(c, "cao", req.CastAgeMax),

		StartRankingDateStart: queryOrFallbackString(c, "srds", req.StartRankingDateStart),
		StartRankingDateEnd:   queryOrFallbackString(c, "srde", req.StartRankingDateEnd),

		DaysInRankMin: queryOrFallbackInt(c, "drkmin", req.DaysInRankMin, false),
		NeedDownload:  queryOrFallbackInt(c, "nd", req.NeedDownload, false),
		Owned:         queryOrFallbackInt(c, "owned", req.Owned, false),
		MediaOwned:    queryOrFallbackInt(c, "mowned", req.MediaOwned, false),

		ViewWatchedMin: queryOrFallbackInt(c, "vwmin", req.ViewWatchedMin, false),
		ViewWatchedMax: queryOrFallbackInt(c, "vwmax", req.ViewWatchedMax, false),
		ScoreMin:       queryOrFallbackFloat(c, "smin", req.ScoreMin),
		ScoreMax:       queryOrFallbackFloat(c, "smax", req.ScoreMax),

		LastScTimeMin: queryOrFallbackString(c, "lsctmin", req.LastScTimeMin),
		LastScTimeMax: queryOrFallbackString(c, "lsctmax", req.LastScTimeMax),
		ScTimesMin:    queryOrFallbackInt(c, "scmin", req.ScTimesMin, false),
		ScTimesMax:    queryOrFallbackIntPtr(c, "scmax", req.ScTimesMax),
		ComeTimesMin:  queryOrFallbackInt(c, "comin", req.ComeTimesMin, false),
		ComeTimesMax:  queryOrFallbackIntPtr(c, "comax", req.ComeTimesMax),

		Dir1: queryOrFallbackString(c, "d1", req.Dir1),
		Dir2: queryOrFallbackString(c, "d2", req.Dir2),
		Dir3: queryOrFallbackString(c, "d3", req.Dir3),
		Dir4: queryOrFallbackString(c, "d4", req.Dir4),
	}

	if randomN != nil {
		view.RandomCount = strconv.Itoa(*randomN)
	}
	return view
}

func buildMovieCardFilterExplicit(c *gin.Context, hasRandom bool) map[string]bool {
	keys := []string{
		"ps", "od", "order",
		"cn", "pid", "gn", "dn", "pn", "mn", "ln", "wd",
		"owned", "nd", "drkmin",
		"mowned",
		"d1", "d2", "d3", "d4",
		"rs", "re", "bs", "be", "mbs", "mbe", "srds", "srde", "lsctmin", "lsctmax",
		"cay", "cao", "vwmin", "vwmax", "smin", "smax", "scmin", "scmax", "comin", "comax",
	}
	if hasRandom {
		keys = append(keys, "n")
	}

	out := make(map[string]bool, len(keys))
	for _, key := range keys {
		_, ok := c.GetQuery(key)
		out[key] = ok
	}
	return out
}

func buildMovieCardFilterClearHref(path string, randomN *int) string {
	if randomN == nil {
		return path
	}
	return path + "?n=" + strconv.Itoa(*randomN)
}

func queryOrFallbackString(c *gin.Context, key, fallback string) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	return fallback
}

func queryOrFallbackInt(c *gin.Context, key string, fallback int64, showZeroFallback bool) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	if fallback == 0 && !showZeroFallback {
		return ""
	}
	return strconv.FormatInt(fallback, 10)
}

func queryOrFallbackIntPtr(c *gin.Context, key string, fallback *int64) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	if fallback == nil {
		return ""
	}
	return strconv.FormatInt(*fallback, 10)
}

func queryOrFallbackFloat(c *gin.Context, key string, fallback float64) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	if fallback == 0 {
		return ""
	}
	return strconv.FormatFloat(fallback, 'f', -1, 64)
}
