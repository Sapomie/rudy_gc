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

	RandomCount string
	Explicit    map[string]bool
	HideDirs    bool

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
	MediaBirthTimeStart string
	MediaBirthTimeEnd   string

	CastAgeMin string
	CastAgeMax string

	StartRankingDateStart string
	StartRankingDateEnd   string

	DaysInRankMin string
	NeedDownload  string
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

	MediaDir1 string
	MediaDir2 string
	MediaDir3 string
	MediaDir4 string
}

func buildMovieCardFilterView(c *gin.Context, req types.ListMovieFullRequest, currentOD string, randomN *int) *movieCardFilterView {
	path := c.Request.URL.Path

	view := &movieCardFilterView{
		Action:    path,
		ClearHref: buildMovieCardFilterClearHref(path, randomN),
		OrderBy:   currentOD,
		Order:     queryOrFallbackString(c, "order", req.Order),
		PageSize:  strconv.FormatInt(req.PageSize, 10),
		Explicit:  buildMovieCardFilterExplicit(c, randomN != nil),
		HideDirs:  false,

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
		MediaBirthTimeStart: queryOrFallbackString(c, "mbs", req.MediaBirthTimeStart),
		MediaBirthTimeEnd:   queryOrFallbackString(c, "mbe", req.MediaBirthTimeEnd),

		CastAgeMin: queryOrFallbackFloat(c, "cay", req.CastAgeMin),
		CastAgeMax: queryOrFallbackFloat(c, "cao", req.CastAgeMax),

		StartRankingDateStart: queryOrFallbackString(c, "srds", req.StartRankingDateStart),
		StartRankingDateEnd:   queryOrFallbackString(c, "srde", req.StartRankingDateEnd),

		DaysInRankMin: queryOrFallbackInt(c, "drkmin", req.DaysInRankMin, false),
		NeedDownload:  queryOrFallbackInt(c, "nd", req.NeedDownload, false),
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

		MediaDir1: queryOrFallbackString(c, "md1", req.MediaDir1),
		MediaDir2: queryOrFallbackString(c, "md2", req.MediaDir2),
		MediaDir3: queryOrFallbackString(c, "md3", req.MediaDir3),
		MediaDir4: queryOrFallbackString(c, "md4", req.MediaDir4),
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
		"mowned", "nd", "drkmin",
		"md1", "md2", "md3", "md4",
		"rs", "re", "mbs", "mbe", "srds", "srde", "lsctmin", "lsctmax",
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
