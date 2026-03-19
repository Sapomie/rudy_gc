package consts

const (
	RankPeriodTypeWeek int64 = 1 + iota
	RankPeriodTypeMonth
	RankPeriodTypeQuarter
	RankPeriodTypeYear
)

const (
	RankPeriodStatusProcessing int64 = 1 + iota
	RankPeriodStatusReady
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
