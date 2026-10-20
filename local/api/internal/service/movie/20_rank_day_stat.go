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
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
)

func (s *Service) BackfillMissingRankStats(ctx context.Context) error {
	if err := s.BackfillMissingRankDayStats(ctx); err != nil {
		return fmt.Errorf("backfill missing rank day stats failed: %w", err)
	}
	if err := s.BackfillRankPeriodStats(ctx); err != nil {
		return fmt.Errorf("rebuild rank period stats failed: %w", err)
	}
	return nil
}

func (s *Service) BackfillMissingRankDayStats(ctx context.Context) error {
	rows, err := s.deps.RankModel.All(ctx)
	if err != nil {
		return err
	}
	existing, err := s.deps.RankDayStatModel.ListAll(ctx)
	if err != nil {
		return err
	}
	existingDays := make(map[int64]struct{}, len(existing))
	for _, row := range existing {
		if row != nil {
			existingDays[row.DayNumber] = struct{}{}
		}
	}
	missingDays := make(map[int64]struct{})
	for _, row := range rows {
		if row == nil || row.DayNumber <= 0 {
			continue
		}
		if _, ok := existingDays[row.DayNumber]; !ok {
			missingDays[row.DayNumber] = struct{}{}
		}
	}
	timings := make([]dailybesttypes.RankDayTiming, 0, len(missingDays))
	for _, dayNumber := range sortedDayNumbers(missingDays) {
		timings = append(timings, dailybesttypes.RankDayTiming{DayNumber: dayNumber})
	}
	return s.insertMissingRankDayStats(ctx, rows, timings)
}

func (s *Service) CreateRankDayStats(ctx context.Context, timings []dailybesttypes.RankDayTiming) error {
	if len(timings) == 0 {
		return nil
	}
	rows, err := s.deps.RankModel.All(ctx)
	if err != nil {
		return err
	}
	return s.insertMissingRankDayStats(ctx, rows, timings)
}

func (s *Service) insertMissingRankDayStats(
	ctx context.Context,
	rows []*moviex.CRank,
	timings []dailybesttypes.RankDayTiming,
) error {
	timingByDay := make(map[int64]dailybesttypes.RankDayTiming, len(timings))
	for _, timing := range timings {
		if timing.DayNumber <= 0 {
			continue
		}
		current := timingByDay[timing.DayNumber]
		current.DayNumber = timing.DayNumber
		if timing.FetchedOn > current.FetchedOn {
			current.FetchedOn = timing.FetchedOn
		}
		if timing.SyncedOn > current.SyncedOn {
			current.SyncedOn = timing.SyncedOn
		}
		timingByDay[timing.DayNumber] = current
	}
	daySet := make(map[int64]struct{}, len(timingByDay))
	for dayNumber := range timingByDay {
		daySet[dayNumber] = struct{}{}
	}
	orderedDays := sortedDayNumbers(daySet)
	calculatedOn := time.Now().Unix()
	stats := make([]*moviex.CRankDayStat, 0, len(orderedDays))
	for _, dayNumber := range orderedDays {
		stat, err := buildRankDayStat(rows, timingByDay[dayNumber], calculatedOn)
		if err != nil {
			return err
		}
		stats = append(stats, stat)
	}
	return s.deps.RankDayStatModel.InsertMissing(ctx, stats)
}

func buildRankDayStat(
	rows []*moviex.CRank,
	timing dailybesttypes.RankDayTiming,
	calculatedOn int64,
) (*moviex.CRankDayStat, error) {
	dayNumber := timing.DayNumber
	ordered := make([]*moviex.CRank, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.DayNumber > 0 {
			ordered = append(ordered, row)
		}
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
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

	dayRows := make([]*moviex.CRank, 0)
	prevDayNumber := int64(0)
	for _, row := range ordered {
		if row.DayNumber == dayNumber {
			dayRows = append(dayRows, row)
		}
		if row.DayNumber < dayNumber && row.DayNumber > prevDayNumber {
			prevDayNumber = row.DayNumber
		}
	}
	if len(dayRows) == 0 {
		return nil, fmt.Errorf("c_rank day_number=%d 没有排行数据", dayNumber)
	}
	rankStoredOn := int64(0)
	for _, row := range dayRows {
		if row.CreatedOn > 0 && (rankStoredOn == 0 || row.CreatedOn < rankStoredOn) {
			rankStoredOn = row.CreatedOn
		}
	}
	if timing.FetchedOn <= 0 {
		timing.FetchedOn = rankStoredOn
	}
	if timing.SyncedOn <= 0 {
		timing.SyncedOn = rankStoredOn
	}

	seenMovies := make(map[string]struct{})
	prevMovies := make(map[string]struct{})
	for _, row := range ordered {
		if row.DayNumber >= dayNumber {
			break
		}
		seenMovies[row.MovieJavId] = struct{}{}
		if row.DayNumber == prevDayNumber {
			prevMovies[row.MovieJavId] = struct{}{}
		}
	}

	currentMovies := make(map[string]struct{})
	hasher := sha256.New()
	for _, row := range dayRows {
		currentMovies[row.MovieJavId] = struct{}{}
		_, _ = fmt.Fprintf(
			hasher,
			"%d:%d:%d:%s\n",
			row.DayNumber,
			row.Category,
			row.RankPos,
			row.MovieJavId,
		)
	}

	stat := &moviex.CRankDayStat{
		DayNumber:      dayNumber,
		RankDate:       consts.GetDateStringByRankDayNumber(dayNumber),
		PrevDayNumber:  prevDayNumber,
		RankRowCount:   int64(len(dayRows)),
		MovieCount:     int64(len(currentMovies)),
		SourceChecksum: hex.EncodeToString(hasher.Sum(nil)),
		FetchedOn:      timing.FetchedOn,
		SyncedOn:       timing.SyncedOn,
		CalculatedOn:   calculatedOn,
		CreatedOn:      calculatedOn,
		UpdatedOn:      calculatedOn,
	}
	for movieJavID := range currentMovies {
		if _, ok := prevMovies[movieJavID]; ok {
			stat.ContinuedMovieCount++
			continue
		}
		if _, ok := seenMovies[movieJavID]; ok {
			stat.ReturningMovieCount++
		} else {
			stat.FirstSeenMovieCount++
		}
	}
	for movieJavID := range prevMovies {
		if _, ok := currentMovies[movieJavID]; !ok {
			stat.DroppedMovieCount++
		}
	}
	return stat, nil
}

func sortedDayNumbers(daySet map[int64]struct{}) []int64 {
	days := make([]int64, 0, len(daySet))
	for dayNumber := range daySet {
		days = append(days, dayNumber)
	}
	sort.Slice(days, func(i, j int) bool { return days[i] < days[j] })
	return days
}
