package loop

import (
	"fmt"
	"strconv"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

var allowedFetchSiteOrderBy = map[string]struct{}{
	consts.OrderByDetailUpdateTime: {},
	consts.OrderByCastAgeDesc:      {},
	consts.OrderByCastAgeAsc:       {},
	consts.OrderByViewerWatched:    {},
	consts.OrderByReleasingDate:    {},
	consts.OrderByRankDate:         {},
	consts.OrderByHighestRank:      {},
	consts.OrderByDaysInRank:       {},
	consts.OrderByBirthTime:        {},
	consts.OrderByScTimes:          {},
	consts.OrderByComeTimes:        {},
	consts.OrderByLastScTime:       {},
}

const (
	fetchSiteDefaultNumber       int64 = 1000000
	fetchSiteDefaultDurationDays int64 = 5
)

func buildFetchSiteMovieRequest(req StartTaskRequest) (*types.ListMovieFullRequest, error) {
	out := &types.ListMovieFullRequest{
		CastNames:             strings.TrimSpace(req.CastNames),
		PersonIds:             strings.TrimSpace(req.PersonIds),
		GenreNames:            strings.TrimSpace(req.GenreNames),
		DirectorName:          strings.TrimSpace(req.DirectorName),
		PrefixName:            strings.TrimSpace(req.PrefixName),
		MakerName:             strings.TrimSpace(req.MakerName),
		LabelName:             strings.TrimSpace(req.LabelName),
		ReleasingDateStart:    strings.TrimSpace(req.ReleasingDateStart),
		ReleasingDateEnd:      strings.TrimSpace(req.ReleasingDateEnd),
		MediaBirthTimeStart:   strings.TrimSpace(req.MediaBirthTimeStart),
		MediaBirthTimeEnd:     strings.TrimSpace(req.MediaBirthTimeEnd),
		StartRankingDateStart: strings.TrimSpace(req.StartRankingDateFrom),
		StartRankingDateEnd:   strings.TrimSpace(req.StartRankingDateTo),
		Word:                  strings.TrimSpace(req.Word),
		LastScTimeMin:         strings.TrimSpace(req.LastScTimeMin),
		LastScTimeMax:         strings.TrimSpace(req.LastScTimeMax),
		MediaDir1:             strings.TrimSpace(req.MediaDir1),
		MediaDir2:             strings.TrimSpace(req.MediaDir2),
		MediaDir3:             strings.TrimSpace(req.MediaDir3),
		MediaDir4:             strings.TrimSpace(req.MediaDir4),
		OrderBy:               normalizeFetchSiteOrderBy(req.OrderBy),
		Order:                 normalizeFetchSiteOrder(req.Order),
		Page:                  1,
		PageSize:              normalizeFetchSiteNumber(req.Number),
	}

	var err error
	if out.CastAgeMin, err = parseFetchSiteFloat(req.CastAgeMin, "cay"); err != nil {
		return nil, err
	}
	if out.CastAgeMax, err = parseFetchSiteFloat(req.CastAgeMax, "cao"); err != nil {
		return nil, err
	}
	if out.DaysInRankMin, err = parseFetchSiteInt(req.DaysInRankMin, "drkmin"); err != nil {
		return nil, err
	}
	if out.NeedDownload, err = parseFetchSiteInt(req.NeedDownload, "nd"); err != nil {
		return nil, err
	}
	if out.MediaOwned, err = parseFetchSiteInt(req.MediaOwned, "mowned"); err != nil {
		return nil, err
	}
	if out.ViewWatchedMin, err = parseFetchSiteInt(req.ViewWatchedMin, "vwmin"); err != nil {
		return nil, err
	}
	if out.ViewWatchedMax, err = parseFetchSiteInt(req.ViewWatchedMax, "vwmax"); err != nil {
		return nil, err
	}
	if out.ScoreMin, err = parseFetchSiteFloat(req.ScoreMin, "smin"); err != nil {
		return nil, err
	}
	if out.ScoreMax, err = parseFetchSiteFloat(req.ScoreMax, "smax"); err != nil {
		return nil, err
	}
	if out.ScTimesMin, err = parseFetchSiteInt(req.ScTimesMin, "scmin"); err != nil {
		return nil, err
	}
	if out.ComeTimesMin, err = parseFetchSiteInt(req.ComeTimesMin, "comin"); err != nil {
		return nil, err
	}
	if out.ScTimesMax, err = parseFetchSiteIntPtr(req.ScTimesMax, "scmax"); err != nil {
		return nil, err
	}
	if out.ComeTimesMax, err = parseFetchSiteIntPtr(req.ComeTimesMax, "comax"); err != nil {
		return nil, err
	}

	return out, nil
}

type fetchSiteDurationFilter struct {
	LastFetchDurationDays   int64
	LastSuccessDurationDays int64
}

func buildFetchSiteDurationFilter(req StartTaskRequest) (fetchSiteDurationFilter, error) {
	var out fetchSiteDurationFilter
	var err error

	if out.LastFetchDurationDays, err = parseFetchSiteNonNegativeInt(req.LastFetchDurationDays, "last_fetch_duration_days"); err != nil {
		return fetchSiteDurationFilter{}, err
	}
	if out.LastSuccessDurationDays, err = parseFetchSiteNonNegativeInt(req.LastSuccessDurationDays, "last_success_duration_days"); err != nil {
		return fetchSiteDurationFilter{}, err
	}
	if out.LastFetchDurationDays <= 0 {
		out.LastFetchDurationDays = fetchSiteDefaultDurationDays
	}
	if out.LastSuccessDurationDays <= 0 {
		out.LastSuccessDurationDays = fetchSiteDefaultDurationDays
	}
	return out, nil
}

func normalizeFetchSiteNumber(number int64) int64 {
	if number > 0 {
		return number
	}
	return fetchSiteDefaultNumber
}

func normalizeFetchSiteOrderBy(input string) string {
	current := strings.TrimSpace(input)
	if _, ok := allowedFetchSiteOrderBy[current]; ok {
		return current
	}
	return consts.OrderByReleasingDate
}

func normalizeFetchSiteOrder(input string) string {
	current := strings.ToLower(strings.TrimSpace(input))
	switch current {
	case "asc", "desc":
		return current
	default:
		return ""
	}
}

func parseFetchSiteInt(raw string, field string) (int64, error) {
	current := strings.TrimSpace(raw)
	if current == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(current, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid", field)
	}
	return value, nil
}

func parseFetchSiteNonNegativeInt(raw string, field string) (int64, error) {
	value, err := parseFetchSiteInt(raw, field)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("%s is invalid", field)
	}
	return value, nil
}

func parseFetchSiteIntPtr(raw string, field string) (*int64, error) {
	current := strings.TrimSpace(raw)
	if current == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(current, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s is invalid", field)
	}
	return &value, nil
}

func parseFetchSiteFloat(raw string, field string) (float64, error) {
	current := strings.TrimSpace(raw)
	if current == "" {
		return 0, nil
	}
	value, err := strconv.ParseFloat(current, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid", field)
	}
	return value, nil
}
