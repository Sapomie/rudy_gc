package types

import (
	"fmt"
	"strings"
	"time"
)

// CalendarPeriod uses local calendar boundaries; weeks follow ISO 8601.
type CalendarPeriod struct {
	Kind  string
	Key   string
	Year  int
	Start time.Time
	End   time.Time
}

func CalendarPeriodAt(kind string, date time.Time) (CalendarPeriod, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	p := CalendarPeriod{Kind: kind, Year: date.Year()}
	var end time.Time
	switch kind {
	case "year":
		start = time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
		p.Key = fmt.Sprintf("year:%04d", p.Year)
		end = start.AddDate(1, 0, 0)
	case "quarter":
		quarter := (int(date.Month())-1)/3 + 1
		start = time.Date(date.Year(), time.Month(3*(quarter-1)+1), 1, 0, 0, 0, 0, date.Location())
		p.Key = fmt.Sprintf("quarter:%04d-Q%d", p.Year, quarter)
		end = start.AddDate(0, 3, 0)
	case "month":
		start = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		p.Key = "month:" + start.Format("2006-01")
		end = start.AddDate(0, 1, 0)
	case "week":
		year, week := date.ISOWeek()
		p.Year = year
		start = start.AddDate(0, 0, -(int(start.Weekday())+6)%7)
		p.Key = fmt.Sprintf("week:%04d-W%02d", year, week)
		end = start.AddDate(0, 0, 7)
	case "day":
		p.Key = "day:" + start.Format(time.DateOnly)
		end = start.AddDate(0, 0, 1)
	default:
		return p, fmt.Errorf("无效周期类型: %s", kind)
	}
	p.Start, p.End = start, end.Add(-time.Second)
	return p, nil
}

// ParseCalendarPeriod rejects keys that normalize into a different period.
func ParseCalendarPeriod(kind, key string, location *time.Location) (CalendarPeriod, error) {
	if location == nil {
		location = time.Local
	}
	value, found := strings.CutPrefix(key, kind+":")
	if !found {
		return CalendarPeriod{}, fmt.Errorf("周期键与周期类型不匹配")
	}
	var date time.Time
	var err error
	switch kind {
	case "year":
		date, err = time.ParseInLocation("2006", value, location)
	case "quarter":
		var year, quarter int
		if _, err = fmt.Sscanf(value, "%04d-Q%d", &year, &quarter); err == nil && quarter >= 1 && quarter <= 4 {
			date = time.Date(year, time.Month(3*(quarter-1)+1), 1, 0, 0, 0, 0, location)
		}
	case "month":
		date, err = time.ParseInLocation("2006-01", value, location)
	case "week":
		var year, week int
		if _, err = fmt.Sscanf(value, "%04d-W%02d", &year, &week); err == nil && week >= 1 && week <= 53 {
			jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, location)
			date = jan4.AddDate(0, 0, -(int(jan4.Weekday())+6)%7+7*(week-1))
		}
	case "day":
		date, err = time.ParseInLocation(time.DateOnly, value, location)
	default:
		return CalendarPeriod{}, fmt.Errorf("无效周期类型: %s", kind)
	}
	if err != nil || date.IsZero() {
		return CalendarPeriod{}, fmt.Errorf("无效周期键: %s", key)
	}
	p, err := CalendarPeriodAt(kind, date)
	if err != nil || p.Key != key {
		return CalendarPeriod{}, fmt.Errorf("无效周期键: %s", key)
	}
	return p, nil
}
