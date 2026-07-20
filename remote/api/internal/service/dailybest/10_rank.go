package dailybest

import (
	"context"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) ListRankSyncManifests(ctx context.Context) ([]contracts.DailyBestRankSyncManifest, error) {
	return s.repository.ListRankSyncManifests(ctx)
}

func (s *Service) ListRanks(ctx context.Context, filter dailybesttypes.RankFilter) ([]contracts.DailyBestRank, int64, error) {
	return s.repository.ListRanks(ctx, filter)
}

func (s *Service) ListRankDayStats(ctx context.Context, filter dailybesttypes.RankDayStatFilter) ([]contracts.DailyBestRankDayStat, int64, error) {
	return s.repository.ListRankDayStats(ctx, filter)
}

func (s *Service) ListRankPeriods(ctx context.Context, filter dailybesttypes.RankPeriodFilter) ([]contracts.DailyBestRankPeriod, int64, error) {
	return s.repository.ListRankPeriods(ctx, filter)
}

func (s *Service) ListRankPeriodItems(ctx context.Context, filter dailybesttypes.RankPeriodItemFilter) ([]contracts.DailyBestRankPeriodItem, int64, error) {
	return s.repository.ListRankPeriodItems(ctx, filter)
}

func (s *Service) ListRankPeriodStats(ctx context.Context, filter dailybesttypes.RankPeriodStatFilter) ([]contracts.DailyBestRankPeriodStat, int64, error) {
	return s.repository.ListRankPeriodStats(ctx, filter)
}
