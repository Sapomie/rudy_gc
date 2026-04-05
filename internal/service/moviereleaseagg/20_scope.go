package moviereleaseagg

import (
	"fmt"
	"time"
)

const (
	rootPath = "/movie-agg-all/release"

	levelRoot    = "root"
	levelYear    = "year"
	levelQuarter = "quarter"
	levelMonth   = "month"
	levelDay     = "day"

	aggTypeCast     = "cast"
	aggTypeDirector = "director"
	aggTypeLabel    = "label"
	aggTypePrefix   = "prefix"

	topPersistLimit = 30
	gbDiv           = 1024.0 * 1024.0 * 1024.0
)

type scope struct {
	Key       string
	Level     string
	Year      int
	Quarter   int
	Month     int
	Day       int
	StartUnix int64
	EndUnix   int64
}

func buildScope(year, quarter, month, day int) scope {
	if year <= 0 {
		return scope{
			Key:       levelRoot,
			Level:     levelRoot,
			StartUnix: 1,
			EndUnix:   time.Date(2100, time.December, 31, 23, 59, 59, 0, time.Local).Unix(),
		}
	}

	if day > 0 {
		start := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		end := start.Add(24*time.Hour - time.Second)
		return scope{
			Key:       fmt.Sprintf("day:%04d-%02d-%02d", year, month, day),
			Level:     levelDay,
			Year:      year,
			Quarter:   monthToQuarter(month),
			Month:     month,
			Day:       day,
			StartUnix: start.Unix(),
			EndUnix:   end.Unix(),
		}
	}

	if month > 0 {
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return scope{
			Key:       fmt.Sprintf("month:%04d-%02d", year, month),
			Level:     levelMonth,
			Year:      year,
			Quarter:   monthToQuarter(month),
			Month:     month,
			StartUnix: start.Unix(),
			EndUnix:   end.Unix(),
		}
	}

	if quarter > 0 {
		startMonth := 3*(quarter-1) + 1
		start := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 3, 0).Add(-time.Second)
		return scope{
			Key:       fmt.Sprintf("quarter:%04d-Q%d", year, quarter),
			Level:     levelQuarter,
			Year:      year,
			Quarter:   quarter,
			StartUnix: start.Unix(),
			EndUnix:   end.Unix(),
		}
	}

	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.Local).Add(-time.Second)
	return scope{
		Key:       fmt.Sprintf("year:%04d", year),
		Level:     levelYear,
		Year:      year,
		StartUnix: start.Unix(),
		EndUnix:   end.Unix(),
	}
}

func scopeFromBucketMonth(bucketMonth int64) scope {
	t := time.Unix(bucketMonth, 0).In(time.Local)
	return buildScope(t.Year(), 0, int(t.Month()), 0)
}

func scopeFromBucketDay(bucketDay int64) scope {
	t := time.Unix(bucketDay, 0).In(time.Local)
	return buildScope(t.Year(), 0, int(t.Month()), t.Day())
}

func bucketMonthFromReleaseTime(releaseTime int64) int64 {
	if releaseTime <= 0 {
		return 0
	}
	t := time.Unix(releaseTime, 0).In(time.Local)
	monthStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	return monthStart.Unix()
}

func monthToQuarter(month int) int {
	return ((month - 1) / 3) + 1
}

func dateOnly(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).In(time.Local).Format(time.DateOnly)
}

func bytesToGB(sizeBytes int64) float64 {
	return float64(sizeBytes) / gbDiv
}

func buildBreadcrumbs(year, quarter, month, day int) []Breadcrumb {
	bcs := []Breadcrumb{{Title: "上映日（全量）", Href: rootPath}}
	if year == 0 {
		bcs[0].Href = ""
		return bcs
	}
	yearHref := fmt.Sprintf("%s/%d", rootPath, year)
	if day > 0 {
		if quarter == 0 {
			quarter = monthToQuarter(month)
		}
		quarterHref := fmt.Sprintf("%s/%d/q/%d", rootPath, year, quarter)
		monthHref := fmt.Sprintf("%s/%d/%02d", rootPath, year, month)
		return append(bcs,
			Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: quarterHref},
			Breadcrumb{Title: fmt.Sprintf("%02d 月", month), Href: monthHref},
			Breadcrumb{Title: fmt.Sprintf("%02d 日", day)},
		)
	}
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
	quarterHref := fmt.Sprintf("%s/%d/q/%d", rootPath, year, quarter)
	return append(bcs,
		Breadcrumb{Title: fmt.Sprintf("%d 年", year), Href: yearHref},
		Breadcrumb{Title: fmt.Sprintf("Q%d 季", quarter), Href: quarterHref},
		Breadcrumb{Title: fmt.Sprintf("%02d 月", month)},
	)
}
