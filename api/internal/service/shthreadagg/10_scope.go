package shthreadagg

import (
	"fmt"
	"time"
)

type bucketScope struct {
	Key       string
	Level     string
	Year      int
	Quarter   int
	Month     int
	Week      int
	Day       int
	StartUnix int64
	EndUnix   int64
}

func buildScopes(postTime int64) []bucketScope {
	if postTime <= 0 {
		return nil
	}
	t := time.Unix(postTime, 0).In(time.Local)
	year, month, day := t.Date()
	quarter := ((int(month) - 1) / 3) + 1
	isoYear, isoWeek := t.ISOWeek()

	dayStart := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)
	weekOffset := (int(dayStart.Weekday()) + 6) % 7
	weekStart := dayStart.AddDate(0, 0, -weekOffset)
	weekEnd := weekStart.AddDate(0, 0, 7).Add(-time.Second)
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
	quarterStartMonth := time.Month(3*(quarter-1) + 1)
	quarterStart := time.Date(year, quarterStartMonth, 1, 0, 0, 0, 0, time.Local)
	quarterEnd := quarterStart.AddDate(0, 3, 0).Add(-time.Second)
	yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	yearEnd := yearStart.AddDate(1, 0, 0).Add(-time.Second)

	return []bucketScope{
		{Key: fmt.Sprintf("day:%04d-%02d-%02d", year, month, day), Level: levelDay, Year: year, Quarter: quarter, Month: int(month), Day: day, StartUnix: dayStart.Unix(), EndUnix: dayEnd.Unix()},
		{Key: fmt.Sprintf("week:%04d-W%02d", isoYear, isoWeek), Level: levelWeek, Year: isoYear, Week: isoWeek, StartUnix: weekStart.Unix(), EndUnix: weekEnd.Unix()},
		{Key: fmt.Sprintf("month:%04d-%02d", year, month), Level: levelMonth, Year: year, Quarter: quarter, Month: int(month), StartUnix: monthStart.Unix(), EndUnix: monthEnd.Unix()},
		{Key: fmt.Sprintf("quarter:%04d-Q%d", year, quarter), Level: levelQuarter, Year: year, Quarter: quarter, StartUnix: quarterStart.Unix(), EndUnix: quarterEnd.Unix()},
		{Key: fmt.Sprintf("year:%04d", year), Level: levelYear, Year: year, StartUnix: yearStart.Unix(), EndUnix: yearEnd.Unix()},
	}
}
