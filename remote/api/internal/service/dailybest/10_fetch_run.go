package dailybest

import (
	"context"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) ListFetchRuns(ctx context.Context, page dailybesttypes.PageQuery) ([]contracts.DailyBestFetchRun, int64, error) {
	return s.repository.ListFetchRuns(ctx, page)
}

func (s *Service) FindFetchRun(ctx context.Context, runID int64) (contracts.DailyBestFetchRun, error) {
	return s.repository.FindFetchRun(ctx, runID)
}
