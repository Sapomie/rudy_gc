package movie

import (
	"fmt"
	"sort"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/consts"
)

const (
	defaultRankPeriodWeekTopN    int64 = 200
	defaultRankPeriodMonthTopN   int64 = 300
	defaultRankPeriodQuarterTopN int64 = 500
	defaultRankPeriodYearTopN    int64 = 1000
	defaultPeriodPageSize              = 18
	maxPeriodPageSize                  = 20000
)

type rankPeriodSpec struct {
	PeriodType     int64
	PeriodKey      string
	StartDayNumber int64
	EndDayNumber   int64
	PickDays       int64
	TopN           int64
}

type rankPeriodCandidate struct {
	MovieJavId         string
	Score              float64
	DaysInRank         int64
	UsedPickDays       int64
	FirstRankDayNumber int64
	LastRankDayNumber  int64
	BestRank           int64
	WorstPickedRank    int64
}

func buildCurrentRankPeriodSpecs(latestDay int64) []rankPeriodSpec {
	latestTime := consts.GetTimeByRankDayNumber(latestDay).In(time.Local)

	weekStart := startOfISOWeek(latestTime)
	monthStart := time.Date(latestTime.Year(), latestTime.Month(), 1, 0, 0, 0, 0, time.Local)
	quarterStart := startOfQuarter(latestTime)
	yearStart := time.Date(latestTime.Year(), time.January, 1, 0, 0, 0, 0, time.Local)

	return []rankPeriodSpec{
		newRankPeriodSpec(consts.RankPeriodTypeWeek, weekPeriodKey(latestTime), weekStart, latestTime),
		newRankPeriodSpec(consts.RankPeriodTypeMonth, latestTime.Format("2006-01"), monthStart, latestTime),
		newRankPeriodSpec(consts.RankPeriodTypeQuarter, quarterPeriodKey(latestTime), quarterStart, latestTime),
		newRankPeriodSpec(consts.RankPeriodTypeYear, latestTime.Format("2006"), yearStart, latestTime),
	}
}

func buildHistoricalRankPeriodSpecs(earliestDay, latestDay int64) []rankPeriodSpec {
	if earliestDay <= 0 || latestDay <= 0 || earliestDay > latestDay {
		return []rankPeriodSpec{}
	}

	specMap := make(map[string]rankPeriodSpec)
	for day := earliestDay; day <= latestDay; day++ {
		specs := buildCurrentRankPeriodSpecs(day)
		for _, spec := range specs {
			key := fmt.Sprintf("%d:%s", spec.PeriodType, spec.PeriodKey)
			specMap[key] = spec
		}
	}

	out := make([]rankPeriodSpec, 0, len(specMap))
	for _, spec := range specMap {
		out = append(out, spec)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].PeriodType != out[j].PeriodType {
			return out[i].PeriodType < out[j].PeriodType
		}
		if out[i].StartDayNumber != out[j].StartDayNumber {
			return out[i].StartDayNumber < out[j].StartDayNumber
		}
		if out[i].EndDayNumber != out[j].EndDayNumber {
			return out[i].EndDayNumber < out[j].EndDayNumber
		}
		return out[i].PeriodKey < out[j].PeriodKey
	})

	return out
}

func newRankPeriodSpec(periodType int64, periodKey string, startTime, endTime time.Time) rankPeriodSpec {
	startDay := consts.GetRankDayNumber(startTime.Format(time.DateOnly))
	endDay := consts.GetRankDayNumber(endTime.Format(time.DateOnly))

	return rankPeriodSpec{
		PeriodType:     periodType,
		PeriodKey:      periodKey,
		StartDayNumber: startDay,
		EndDayNumber:   endDay,
		PickDays:       0,
		TopN:           defaultRankPeriodTopN(periodType),
	}
}

func buildRankPeriodItems(spec rankPeriodSpec, rows []*moviex.CRank, prevRankMap, peakRankMap map[string]int64) (rankPeriodSpec, []*moviex.CRankPeriodItem) {
	grouped := make(map[string][]*moviex.CRank)
	for _, row := range rows {
		if row == nil || row.MovieJavId == "" {
			continue
		}
		grouped[row.MovieJavId] = append(grouped[row.MovieJavId], row)
	}

	spec.PickDays = calcRankPeriodPickDays(grouped, spec.TopN)
	candidates := make([]rankPeriodCandidate, 0, len(grouped))
	for movieJavID, movieRows := range grouped {
		candidate, ok := buildRankPeriodCandidate(movieJavID, movieRows, spec.PickDays)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score < candidates[j].Score
		}
		if candidates[i].BestRank != candidates[j].BestRank {
			return candidates[i].BestRank < candidates[j].BestRank
		}
		if candidates[i].DaysInRank != candidates[j].DaysInRank {
			return candidates[i].DaysInRank > candidates[j].DaysInRank
		}
		return candidates[i].MovieJavId < candidates[j].MovieJavId
	})

	if spec.TopN > 0 && int64(len(candidates)) > spec.TopN {
		candidates = candidates[:spec.TopN]
	}

	items := make([]*moviex.CRankPeriodItem, 0, len(candidates))
	for idx, candidate := range candidates {
		rankPos := int64(idx + 1)
		prevRank := prevRankMap[candidate.MovieJavId]
		rankChange := int64(0)
		if prevRank > 0 {
			rankChange = prevRank - rankPos
		}
		peakRank := rankPos
		if oldPeak := peakRankMap[candidate.MovieJavId]; oldPeak > 0 && oldPeak < peakRank {
			peakRank = oldPeak
		}

		items = append(items, &moviex.CRankPeriodItem{
			MovieJavId:         candidate.MovieJavId,
			RankPos:            rankPos,
			Score:              candidate.Score,
			DaysInRank:         candidate.DaysInRank,
			UsedPickDays:       candidate.UsedPickDays,
			FirstRankDayNumber: candidate.FirstRankDayNumber,
			LastRankDayNumber:  candidate.LastRankDayNumber,
			BestRank:           candidate.BestRank,
			WorstPickedRank:    candidate.WorstPickedRank,
			PrevRank:           prevRank,
			RankChange:         rankChange,
			PeakRank:           peakRank,
		})
	}

	return spec, items
}

func buildRankPeriodCandidate(movieJavID string, rows []*moviex.CRank, pickDays int64) (rankPeriodCandidate, bool) {
	if pickDays <= 0 || int64(len(rows)) < pickDays {
		return rankPeriodCandidate{}, false
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].RankPos != rows[j].RankPos {
			return rows[i].RankPos < rows[j].RankPos
		}
		if rows[i].DayNumber != rows[j].DayNumber {
			return rows[i].DayNumber < rows[j].DayNumber
		}
		return rows[i].Id < rows[j].Id
	})

	firstDay := rows[0].DayNumber
	lastDay := rows[0].DayNumber
	bestRank := rows[0].RankPos

	for _, row := range rows[1:] {
		if row.DayNumber < firstDay {
			firstDay = row.DayNumber
		}
		if row.DayNumber > lastDay {
			lastDay = row.DayNumber
		}
		if row.RankPos < bestRank {
			bestRank = row.RankPos
		}
	}

	usedPickDays := pickDays
	picked := rows[:usedPickDays]
	var scoreSum float64
	worstPickedRank := picked[0].RankPos
	for _, row := range picked {
		scoreSum += float64(row.RankPos)
		if row.RankPos > worstPickedRank {
			worstPickedRank = row.RankPos
		}
	}

	return rankPeriodCandidate{
		MovieJavId:         movieJavID,
		Score:              scoreSum / float64(usedPickDays),
		DaysInRank:         int64(len(rows)),
		UsedPickDays:       usedPickDays,
		FirstRankDayNumber: firstDay,
		LastRankDayNumber:  lastDay,
		BestRank:           bestRank,
		WorstPickedRank:    worstPickedRank,
	}, true
}

func calcRankPeriodPickDays(grouped map[string][]*moviex.CRank, topN int64) int64 {
	if len(grouped) == 0 {
		return 0
	}

	counts := make([]int64, 0, len(grouped))
	for _, rows := range grouped {
		counts = append(counts, int64(len(rows)))
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i] > counts[j]
	})

	if topN <= 0 {
		return counts[len(counts)-1]
	}
	if topN > int64(len(counts)) {
		return counts[len(counts)-1]
	}
	return counts[topN-1]
}

func defaultRankPeriodTopN(periodType int64) int64 {
	switch periodType {
	case consts.RankPeriodTypeWeek:
		return defaultRankPeriodWeekTopN
	case consts.RankPeriodTypeMonth:
		return defaultRankPeriodMonthTopN
	case consts.RankPeriodTypeQuarter:
		return defaultRankPeriodQuarterTopN
	case consts.RankPeriodTypeYear:
		return defaultRankPeriodYearTopN
	default:
		return defaultRankPeriodMonthTopN
	}
}

func startOfISOWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-(weekday-1), 0, 0, 0, 0, time.Local)
}

func startOfQuarter(t time.Time) time.Time {
	month := ((int(t.Month())-1)/3)*3 + 1
	return time.Date(t.Year(), time.Month(month), 1, 0, 0, 0, 0, time.Local)
}

func weekPeriodKey(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%04d-W%02d", year, week)
}

func quarterPeriodKey(t time.Time) string {
	quarter := ((int(t.Month()) - 1) / 3) + 1
	return fmt.Sprintf("%04d-Q%d", t.Year(), quarter)
}
