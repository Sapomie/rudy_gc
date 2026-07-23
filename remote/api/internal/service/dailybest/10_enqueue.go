package dailybest

import (
	"context"
	"errors"
	"fmt"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) CurrentBusinessDate(now time.Time) string {
	return now.In(s.loc).Add(-8 * time.Hour).Format(time.DateOnly)
}

func (s *Service) Enqueue(ctx context.Context, triggerType string) (contracts.DailyBestFetchRun, error) {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	date := s.CurrentBusinessDate(time.Now())
	category := contracts.DailyBestCategoryMonth
	if triggerType == "" {
		triggerType = dailybesttypes.TriggerManual
	}
	run, err := s.repository.CreateFetchRun(ctx, date, category, triggerType)
	if err != nil {
		return contracts.DailyBestFetchRun{}, err
	}
	s.wg.Add(1)
	go func(runID int64, businessDate string, category int64, triggerType string) {
		defer s.wg.Done()
		s.execute(s.root, runID, businessDate, category, triggerType)
	}(run.ID, date, category, triggerType)
	return run, nil
}

func (s *Service) enqueueAutomatic(ctx context.Context, triggerType string) (bool, error) {
	if s.schedulerRepository == nil {
		return false, fmt.Errorf("DailyBest 调度仓库为空")
	}
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	date := s.CurrentBusinessDate(time.Now())
	category := contracts.DailyBestCategoryMonth
	state, err := s.schedulerRepository.FindAutomaticFetchState(ctx, date, category)
	if err != nil {
		return false, err
	}
	if state.Successful != nil || state.Active != nil {
		return false, nil
	}
	run, err := s.repository.CreateFetchRun(ctx, date, category, triggerType)
	if err != nil {
		if errors.Is(err, dailybesttypes.ErrActiveFetchRun) {
			return false, nil
		}
		return false, err
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.execute(s.root, run.ID, date, category, triggerType)
	}()
	return true, nil
}
