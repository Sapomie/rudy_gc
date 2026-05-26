package consts

import (
	"strconv"
	"time"
)

const (
	RankPeriodTypeWeek int64 = 1 + iota
	RankPeriodTypeMonth
	RankPeriodTypeQuarter
	RankPeriodTypeYear
)

const (
	BestCategoryMonth   int64 = 1
	BestCategoryAllTime int64 = 2
)

func RankPeriodTypeName(periodType int64) string {
	switch periodType {
	case RankPeriodTypeWeek:
		return "week"
	case RankPeriodTypeMonth:
		return "month"
	case RankPeriodTypeQuarter:
		return "quarter"
	case RankPeriodTypeYear:
		return "year"
	default:
		return "month"
	}
}

func RankPeriodTypeLabel(periodType int64) string {
	switch periodType {
	case RankPeriodTypeWeek:
		return "Week"
	case RankPeriodTypeMonth:
		return "Month"
	case RankPeriodTypeQuarter:
		return "Quarter"
	case RankPeriodTypeYear:
		return "Year"
	default:
		return "Month"
	}
}

func RankPeriodTypeFromName(name string) int64 {
	switch name {
	case "week":
		return RankPeriodTypeWeek
	case "month":
		return RankPeriodTypeMonth
	case "quarter":
		return RankPeriodTypeQuarter
	case "year":
		return RankPeriodTypeYear
	default:
		return RankPeriodTypeMonth
	}
}

func BestCategoryLabel(category int64) string {
	switch category {
	case BestCategoryMonth:
		return "MonthSource"
	case BestCategoryAllTime:
		return "AllTimeSource"
	default:
		return "Unknown"
	}
}

func GetDateStringByRankDayNumber(dayNumber int64) string {
	if dayNumber <= 0 {
		return ""
	}
	base := time.Date(2011, 11, 21, 0, 0, 0, 0, time.Local)
	return base.AddDate(0, 0, int(dayNumber-1)).Format(time.DateOnly)
}

func GetRankDayNumber(dateString string) int64 {
	parsed, err := time.ParseInLocation(time.DateOnly, dateString, time.Local)
	if err != nil {
		return 0
	}
	base := time.Date(2011, 11, 21, 0, 0, 0, 0, time.Local)
	return int64(parsed.Sub(base).Hours()/24) + 1
}

func ParseCategory(raw string, fallback int64) int64 {
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
