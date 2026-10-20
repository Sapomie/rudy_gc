package movie

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type rankStatPeriodSpec struct {
	PeriodType     int64
	PeriodKey      string
	StartDayNumber int64
	EndDayNumber   int64
}

func (s *Service) BackfillRankPeriodStats(ctx context.Context) error {
	rankRows, err := s.deps.RankModel.All(ctx)
	if err != nil {
		return err
	}
	dayStats, err := s.deps.RankDayStatModel.ListAll(ctx)
	if err != nil {
		return err
	}

	earliestDay, latestDay := rankStatSourceDayRange(rankRows)
	specs := buildHistoricalRankStatPeriodSpecs(earliestDay, latestDay)
	rows := buildRankPeriodStats(specs, rankRows, dayStats, time.Now().Unix())
	return s.deps.RankPeriodStatModel.Upsert(ctx, rows)
}

func (s *Service) RefreshRankPeriodStats(ctx context.Context, dayNumbers []int64) error {
	if len(dayNumbers) == 0 {
		return nil
	}
	rankRows, err := s.deps.RankModel.All(ctx)
	if err != nil {
		return err
	}
	dayStats, err := s.deps.RankDayStatModel.ListAll(ctx)
	if err != nil {
		return err
	}
	earliestDay, latestDay := rankStatSourceDayRange(rankRows)
	specs := rankStatPeriodSpecsForDays(
		buildHistoricalRankStatPeriodSpecs(earliestDay, latestDay),
		dayNumbers,
	)
	rows := buildRankPeriodStats(specs, rankRows, dayStats, time.Now().Unix())
	return s.deps.RankPeriodStatModel.Upsert(ctx, rows)
}

func rankStatPeriodSpecsForDays(specs []rankStatPeriodSpec, dayNumbers []int64) []rankStatPeriodSpec {
	daySet := make(map[int64]struct{}, len(dayNumbers))
	for _, dayNumber := range dayNumbers {
		if dayNumber > 0 {
			daySet[dayNumber] = struct{}{}
		}
	}
	out := make([]rankStatPeriodSpec, 0)
	for _, spec := range specs {
		for dayNumber := range daySet {
			if dayNumber >= spec.StartDayNumber && dayNumber <= spec.EndDayNumber {
				out = append(out, spec)
				break
			}
		}
	}
	return out
}

func buildRankPeriodStats(
	specs []rankStatPeriodSpec,
	rankRows []*moviex.CRank,
	dayStats []*moviex.CRankDayStat,
	calculatedOn int64,
) []*moviex.CRankPeriodStat {
	rows := make([]*moviex.CRankPeriodStat, 0, len(specs))
	for _, spec := range specs {
		periodRankRows := make([]*moviex.CRank, 0)
		for _, row := range rankRows {
			if row != nil && row.DayNumber >= spec.StartDayNumber && row.DayNumber <= spec.EndDayNumber {
				periodRankRows = append(periodRankRows, row)
			}
		}

		periodDayStats := make([]*moviex.CRankDayStat, 0)
		for _, row := range dayStats {
			if row != nil && row.DayNumber >= spec.StartDayNumber && row.DayNumber <= spec.EndDayNumber {
				periodDayStats = append(periodDayStats, row)
			}
		}
		rows = append(rows, buildRankPeriodStat(spec, periodRankRows, periodDayStats, calculatedOn))
	}
	return rows
}

func buildRankPeriodStat(
	spec rankStatPeriodSpec,
	rankRows []*moviex.CRank,
	dayStats []*moviex.CRankDayStat,
	calculatedOn int64,
) *moviex.CRankPeriodStat {
	stat := &moviex.CRankPeriodStat{
		PeriodType:     spec.PeriodType,
		PeriodKey:      spec.PeriodKey,
		StartDayNumber: spec.StartDayNumber,
		EndDayNumber:   spec.EndDayNumber,
		CalculatedOn:   calculatedOn,
		CreatedOn:      calculatedOn,
		UpdatedOn:      calculatedOn,
	}

	orderedRanks := append([]*moviex.CRank(nil), rankRows...)
	sort.SliceStable(orderedRanks, func(i, j int) bool {
		left, right := orderedRanks[i], orderedRanks[j]
		if left.DayNumber != right.DayNumber {
			return left.DayNumber < right.DayNumber
		}
		if left.Category != right.Category {
			return left.Category < right.Category
		}
		if left.RankPos != right.RankPos {
			return left.RankPos < right.RankPos
		}
		if left.MovieJavId != right.MovieJavId {
			return left.MovieJavId < right.MovieJavId
		}
		return left.Id < right.Id
	})

	movies := make(map[string]struct{})
	hasher := sha256.New()
	for _, row := range orderedRanks {
		if row == nil {
			continue
		}
		movies[row.MovieJavId] = struct{}{}
		if row.CreatedOn > 0 && (stat.SourceFetchedStartOn == 0 || row.CreatedOn < stat.SourceFetchedStartOn) {
			stat.SourceFetchedStartOn = row.CreatedOn
		}
		if row.CreatedOn > stat.SourceFetchedEndOn {
			stat.SourceFetchedEndOn = row.CreatedOn
		}
		_, _ = fmt.Fprintf(
			hasher,
			"rank:%d:%d:%d:%s:%d:%d\n",
			row.DayNumber,
			row.Category,
			row.RankPos,
			row.MovieJavId,
			row.CreatedOn,
			row.UpdatedOn,
		)
	}
	stat.SourceMovieCount = int64(len(movies))

	orderedDayStats := append([]*moviex.CRankDayStat(nil), dayStats...)
	sort.SliceStable(orderedDayStats, func(i, j int) bool {
		if orderedDayStats[i].DayNumber != orderedDayStats[j].DayNumber {
			return orderedDayStats[i].DayNumber < orderedDayStats[j].DayNumber
		}
		return orderedDayStats[i].Id < orderedDayStats[j].Id
	})
	for _, row := range orderedDayStats {
		if row == nil {
			continue
		}
		stat.SourceRankRowCount += row.RankRowCount
		stat.SourceRankDayCount++
		stat.PeriodMovieCount += row.MovieCount
		stat.NewEntryMovieCount += row.FirstSeenMovieCount
		_, _ = fmt.Fprintf(
			hasher,
			"day:%d:%d:%d:%d:%d:%s\n",
			row.DayNumber,
			row.RankRowCount,
			row.MovieCount,
			row.FirstSeenMovieCount,
			row.ReturningMovieCount,
			row.SourceChecksum,
		)
	}
	stat.MissingRankDayCount = spec.EndDayNumber - spec.StartDayNumber + 1 - stat.SourceRankDayCount
	if stat.MissingRankDayCount < 0 {
		stat.MissingRankDayCount = 0
	}
	stat.SourceChecksum = hex.EncodeToString(hasher.Sum(nil))
	return stat
}

func rankStatSourceDayRange(rows []*moviex.CRank) (int64, int64) {
	var earliestDay int64
	var latestDay int64
	for _, row := range rows {
		if row == nil || row.DayNumber <= 0 {
			continue
		}
		if earliestDay == 0 || row.DayNumber < earliestDay {
			earliestDay = row.DayNumber
		}
		if row.DayNumber > latestDay {
			latestDay = row.DayNumber
		}
	}
	return earliestDay, latestDay
}

func buildHistoricalRankStatPeriodSpecs(earliestDay, latestDay int64) []rankStatPeriodSpec {
	if earliestDay <= 0 || latestDay <= 0 || earliestDay > latestDay {
		return []rankStatPeriodSpec{}
	}

	specMap := make(map[string]rankStatPeriodSpec)
	for day := earliestDay; day <= latestDay; day++ {
		for _, spec := range buildCurrentRankStatPeriodSpecs(day) {
			key := fmt.Sprintf("%d:%s", spec.PeriodType, spec.PeriodKey)
			specMap[key] = spec
		}
	}

	rows := make([]rankStatPeriodSpec, 0, len(specMap))
	for _, spec := range specMap {
		rows = append(rows, spec)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].PeriodType != rows[j].PeriodType {
			return rows[i].PeriodType < rows[j].PeriodType
		}
		if rows[i].StartDayNumber != rows[j].StartDayNumber {
			return rows[i].StartDayNumber < rows[j].StartDayNumber
		}
		return rows[i].PeriodKey < rows[j].PeriodKey
	})
	return rows
}

func buildCurrentRankStatPeriodSpecs(dayNumber int64) []rankStatPeriodSpec {
	current := consts.GetTimeByRankDayNumber(dayNumber).In(time.Local)
	return []rankStatPeriodSpec{
		newRankStatPeriodSpec(consts.RankPeriodTypeWeek, weekPeriodKey(current), startOfISOWeek(current), current),
		newRankStatPeriodSpec(
			consts.RankPeriodTypeMonth,
			current.Format("2006-01"),
			time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, time.Local),
			current,
		),
		newRankStatPeriodSpec(consts.RankPeriodTypeQuarter, quarterPeriodKey(current), startOfQuarter(current), current),
		newRankStatPeriodSpec(
			consts.RankPeriodTypeYear,
			current.Format("2006"),
			time.Date(current.Year(), time.January, 1, 0, 0, 0, 0, time.Local),
			current,
		),
	}
}

func newRankStatPeriodSpec(periodType int64, periodKey string, startTime, endTime time.Time) rankStatPeriodSpec {
	return rankStatPeriodSpec{
		PeriodType:     periodType,
		PeriodKey:      periodKey,
		StartDayNumber: consts.GetRankDayNumber(startTime.Format(time.DateOnly)),
		EndDayNumber:   consts.GetRankDayNumber(endTime.Format(time.DateOnly)),
	}
}
