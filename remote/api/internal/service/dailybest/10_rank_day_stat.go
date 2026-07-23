package dailybest

import (
	"context"
	"sort"
	"time"

	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) EnsureRankDayStats(ctx context.Context) error {
	existingDays, err := s.repository.ListRankDayStatDayNumbers(ctx)
	if err != nil {
		return err
	}
	existing := make(map[int64]struct{}, len(existingDays))
	for _, dayNumber := range existingDays {
		existing[dayNumber] = struct{}{}
	}
	ranks, err := s.repository.ListAllRanks(ctx)
	if err != nil {
		return err
	}
	missingDays := make(map[int64]struct{})
	for _, rank := range ranks {
		if _, ok := existing[rank.DayNumber]; !ok {
			missingDays[rank.DayNumber] = struct{}{}
		}
	}
	dayNumbers := make([]int64, 0, len(missingDays))
	for dayNumber := range missingDays {
		dayNumbers = append(dayNumbers, dayNumber)
	}
	sort.Slice(dayNumbers, func(i, j int) bool { return dayNumbers[i] < dayNumbers[j] })
	now := time.Now().Unix()
	stats := make([]contracts.DailyBestRankDayStat, 0, len(dayNumbers))
	for _, dayNumber := range dayNumbers {
		stat, err := buildRankDayStat(ranks, dayNumber, 0, 0, now)
		if err != nil {
			return err
		}
		stats = append(stats, stat)
	}
	return s.repository.InsertRankDayStatsIfMissing(ctx, stats)
}
