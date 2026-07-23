package dailybest

import (
	"context"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) ListRankDays(ctx context.Context, filter dailybesttypes.RankDayFilter) ([]contracts.DailyBestRankDay, int64, error) {
	return s.repository.ListRankDays(ctx, filter)
}

func (s *Service) ListRanks(ctx context.Context, filter dailybesttypes.RankFilter) ([]contracts.DailyBestRank, int64, error) {
	return s.repository.ListRanks(ctx, filter)
}

func (s *Service) ListRankDayStats(ctx context.Context, filter dailybesttypes.RankDayStatFilter) ([]contracts.DailyBestRankDayStat, int64, error) {
	return s.repository.ListRankDayStats(ctx, filter)
}
