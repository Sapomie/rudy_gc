package common

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/types"
)

type MovieCardFilterView struct {
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
	LabelJavID   string
	AlbumName    string
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

	LastSelectedScTimeMin string
	LastSelectedScTimeMax string
	ScTimesMin            string
	ScTimesMax            string
	ComeTimesMin          string
	ComeTimesMax          string

	MediaDir1 string
	MediaDir2 string
	MediaDir3 string
	MediaDir4 string
}

func BuildMovieCardFilterView(c *gin.Context, request types.ListMovieFullRequest, currentOrderBy string, randomCount *int) *MovieCardFilterView {
	path := c.Request.URL.Path
	view := &MovieCardFilterView{
		Action:         path,
		ClearHref:      buildMovieCardFilterClearHref(path, randomCount),
		OrderBy:        currentOrderBy,
		Order:          queryOrFallbackString(c, "order", request.Order),
		PageSize:       strconv.FormatInt(request.PageSize, 10),
		Explicit:       buildMovieCardFilterExplicit(c, randomCount != nil),
		HideDirs:       false,
		TextDateInputs: true,

		CastNames:    queryOrFallbackString(c, "cn", request.CastNames),
		PersonIds:    queryOrFallbackString(c, "pid", request.PersonIds),
		GenreNames:   queryOrFallbackString(c, "gn", request.GenreNames),
		DirectorName: queryOrFallbackString(c, "dn", request.DirectorName),
		PrefixName:   queryOrFallbackString(c, "pn", request.PrefixName),
		MakerName:    queryOrFallbackString(c, "mn", request.MakerName),
		LabelName:    queryOrFallbackString(c, "ln", request.LabelName),
		LabelJavID:   queryOrFallbackString(c, "lj", request.LabelJavID),
		AlbumName:    queryOrFallbackString(c, "an", request.AlbumName),
		Word:         queryOrFallbackString(c, "wd", request.Word),

		ReleasingDateStart:  queryOrFallbackString(c, "rs", request.ReleasingDateStart),
		ReleasingDateEnd:    queryOrFallbackString(c, "re", request.ReleasingDateEnd),
		MediaBirthTimeStart: queryOrFallbackString(c, "mbs", request.MediaBirthTimeStart),
		MediaBirthTimeEnd:   queryOrFallbackString(c, "mbe", request.MediaBirthTimeEnd),

		CastAgeMin: queryOrFallbackFloat(c, "cay", request.CastAgeMin),
		CastAgeMax: queryOrFallbackFloat(c, "cao", request.CastAgeMax),

		StartRankingDateStart: queryOrFallbackString(c, "srds", request.StartRankingDateStart),
		StartRankingDateEnd:   queryOrFallbackString(c, "srde", request.StartRankingDateEnd),

		DaysInRankMin: queryOrFallbackInt(c, "drkmin", request.DaysInRankMin, false),
		NeedDownload:  queryOrFallbackInt(c, "nd", request.NeedDownload, false),
		MediaOwned:    queryOrFallbackInt(c, "mowned", request.MediaOwned, false),

		ViewWatchedMin: queryOrFallbackInt(c, "vwmin", request.ViewWatchedMin, false),
		ViewWatchedMax: queryOrFallbackInt(c, "vwmax", request.ViewWatchedMax, false),
		ScoreMin:       queryOrFallbackFloat(c, "smin", request.ScoreMin),
		ScoreMax:       queryOrFallbackFloat(c, "smax", request.ScoreMax),

		LastSelectedScTimeMin: queryOrFallbackString(c, "lsstmin", request.LastSelectedScTimeMin),
		LastSelectedScTimeMax: queryOrFallbackString(c, "lsstmax", request.LastSelectedScTimeMax),
		ScTimesMin:            queryOrFallbackInt(c, "scmin", request.ScTimesMin, false),
		ScTimesMax:            queryOrFallbackIntPtr(c, "scmax", request.ScTimesMax),
		ComeTimesMin:          queryOrFallbackInt(c, "comin", request.ComeTimesMin, false),
		ComeTimesMax:          queryOrFallbackIntPtr(c, "comax", request.ComeTimesMax),

		MediaDir1: queryOrFallbackString(c, "md1", request.MediaDir1),
		MediaDir2: queryOrFallbackString(c, "md2", request.MediaDir2),
		MediaDir3: queryOrFallbackString(c, "md3", request.MediaDir3),
		MediaDir4: queryOrFallbackString(c, "md4", request.MediaDir4),
	}

	if randomCount != nil {
		view.RandomCount = strconv.Itoa(*randomCount)
	}
	return view
}

func buildMovieCardFilterExplicit(c *gin.Context, hasRandom bool) map[string]bool {
	keys := []string{
		"ps", "od", "order",
		"cn", "pid", "gn", "dn", "pn", "mn", "ln", "lj", "an", "wd",
		"nd", "drkmin",
		"mowned",
		"md1", "md2", "md3", "md4",
		"rs", "re", "mbs", "mbe", "srds", "srde", "lsstmin", "lsstmax",
		"cay", "cao", "vwmin", "vwmax", "smin", "smax", "scmin", "scmax", "comin", "comax",
	}
	if hasRandom {
		keys = append(keys, "n")
	}

	explicit := make(map[string]bool, len(keys))
	for _, key := range keys {
		_, ok := c.GetQuery(key)
		explicit[key] = ok
	}
	return explicit
}

func buildMovieCardFilterClearHref(path string, randomCount *int) string {
	if randomCount == nil {
		return path
	}
	return path + "?n=" + strconv.Itoa(*randomCount)
}

func queryOrFallbackString(c *gin.Context, key, fallback string) string {
	if value, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func queryOrFallbackInt(c *gin.Context, key string, fallback int64, showZeroFallback bool) string {
	if value, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(value)
	}
	if fallback == 0 && !showZeroFallback {
		return ""
	}
	return strconv.FormatInt(fallback, 10)
}

func queryOrFallbackIntPtr(c *gin.Context, key string, fallback *int64) string {
	if value, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(value)
	}
	if fallback == nil {
		return ""
	}
	return strconv.FormatInt(*fallback, 10)
}

func queryOrFallbackFloat(c *gin.Context, key string, fallback float64) string {
	if value, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(value)
	}
	if fallback == 0 {
		return ""
	}
	return strconv.FormatFloat(fallback, 'f', -1, 64)
}

func UniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
