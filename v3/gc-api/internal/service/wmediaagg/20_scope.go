package wmediaagg

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"
)

const (
	levelRoot    = "root"
	levelYear    = "year"
	levelQuarter = "quarter"
	levelMonth   = "month"
	levelDay     = "day"

	aggTypeCast     = "cast"
	aggTypeDirector = "director"
	aggTypeLabel    = "label"
	aggTypePrefix   = "prefix"

	rootPath        = "/w-media-agg/birth"
	topPersistLimit = 200
	gbDiv           = 1024.0 * 1024.0 * 1024.0
)

type scope struct {
	Level     string
	Key       string
	Year      int
	Quarter   int
	Month     int
	Day       int
	StartUnix int64
	EndUnix   int64
	StartDate string
	EndDate   string
}

func detectLevel(year, quarter, month, day int) string {
	switch {
	case year <= 0:
		return levelRoot
	case day > 0:
		return levelDay
	case month > 0:
		return levelMonth
	case quarter > 0:
		return levelQuarter
	default:
		return levelYear
	}
}

func buildScope(year, quarter, month, day int) scope {
	level := detectLevel(year, quarter, month, day)
	switch level {
	case levelRoot:
		return scope{
			Level:     levelRoot,
			Key:       levelRoot,
			StartUnix: 1,
			EndUnix:   math.MaxInt64,
		}
	case levelYear:
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
		end := time.Date(year, time.December, 31, 23, 59, 59, 0, time.Local)
		return scope{
			Level:     levelYear,
			Key:       scopeKey(levelYear, year, 0, 0, 0),
			Year:      year,
			StartUnix: start.Unix(),
			EndUnix:   end.Unix(),
			StartDate: start.Format(time.DateOnly),
			EndDate:   end.Format(time.DateOnly),
		}
	case levelQuarter:
		startMonth := 3*(quarter-1) + 1
		start := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.Local)
		end := time.Date(year, time.Month(startMonth+3), 1, 0, 0, 0, 0, time.Local).Add(-time.Second)
		return scope{
			Level:     levelQuarter,
			Key:       scopeKey(levelQuarter, year, quarter, 0, 0),
			Year:      year,
			Quarter:   quarter,
			StartUnix: start.Unix(),
			EndUnix:   end.Unix(),
			StartDate: start.Format(time.DateOnly),
			EndDate:   end.Format(time.DateOnly),
		}
	case levelMonth:
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		end := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Second)
		return scope{
			Level:     levelMonth,
			Key:       scopeKey(levelMonth, year, monthToQuarter(month), month, 0),
			Year:      year,
			Quarter:   monthToQuarter(month),
			Month:     month,
			StartUnix: start.Unix(),
			EndUnix:   end.Unix(),
			StartDate: start.Format(time.DateOnly),
			EndDate:   end.Format(time.DateOnly),
		}
	default:
		start := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		end := start.Add(24*time.Hour - time.Second)
		return scope{
			Level:     levelDay,
			Key:       scopeKey(levelDay, year, monthToQuarter(month), month, day),
			Year:      year,
			Quarter:   monthToQuarter(month),
			Month:     month,
			Day:       day,
			StartUnix: start.Unix(),
			EndUnix:   end.Unix(),
			StartDate: start.Format(time.DateOnly),
			EndDate:   end.Format(time.DateOnly),
		}
	}
}

func scopeKey(level string, year, quarter, month, day int) string {
	switch level {
	case levelYear:
		return fmt.Sprintf("year:%04d", year)
	case levelQuarter:
		return fmt.Sprintf("quarter:%04d-Q%d", year, quarter)
	case levelMonth:
		return fmt.Sprintf("month:%04d-%02d", year, month)
	case levelDay:
		return fmt.Sprintf("day:%04d-%02d-%02d", year, month, day)
	default:
		return levelRoot
	}
}

func scopeFromBucketDay(bucketDay int64) scope {
	t := time.Unix(bucketDay, 0).In(time.Local)
	return buildScope(t.Year(), 0, int(t.Month()), t.Day())
}

func bucketDayFromBirthTime(birthTime int64) int64 {
	if birthTime <= 0 {
		return 0
	}
	t := time.Unix(birthTime, 0).In(time.Local)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
}

func monthToQuarter(month int) int {
	if month <= 0 {
		return 0
	}
	return ((month - 1) / 3) + 1
}

func bytesToGB(b int64) float64 {
	return float64(b) / gbDiv
}

func buildBirthAggHref(year, quarter, month, day int) string {
	q := url.Values{}
	if year > 0 {
		q.Set("year", strconv.Itoa(year))
	}
	if quarter > 0 {
		q.Set("quarter", strconv.Itoa(quarter))
	}
	if month > 0 {
		q.Set("month", strconv.Itoa(month))
	}
	if day > 0 {
		q.Set("day", strconv.Itoa(day))
	}
	if enc := q.Encode(); enc != "" {
		return rootPath + "?" + enc
	}
	return rootPath
}
