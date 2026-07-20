package dailybest

import (
	"context"
	"fmt"
	"strings"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) CurrentBusinessDate(now time.Time) string {
	return now.In(s.loc).Add(-8 * time.Hour).Format(time.DateOnly)
}

func (s *Service) Enqueue(ctx context.Context, req contracts.DailyBestFetchRequest, triggerType string) (contracts.DailyBestFetchRun, error) {
	date := strings.TrimSpace(req.BusinessDate)
	if date == "" {
		date = s.CurrentBusinessDate(time.Now())
	}
	if _, err := time.ParseInLocation(time.DateOnly, date, s.loc); err != nil {
		return contracts.DailyBestFetchRun{}, fmt.Errorf("business_date 格式错误: %w", err)
	}
	currentDate := s.CurrentBusinessDate(time.Now())
	if date != currentDate {
		return contracts.DailyBestFetchRun{}, fmt.Errorf("源站只提供当前榜单，business_date 必须为 %s", currentDate)
	}
	category := req.Category
	if category == 0 {
		category = contracts.DailyBestCategoryMonth
	}
	if category != contracts.DailyBestCategoryMonth {
		return contracts.DailyBestFetchRun{}, fmt.Errorf("当前只支持 DailyBest 月榜 category=%d", contracts.DailyBestCategoryMonth)
	}
	if triggerType == "" {
		triggerType = dailybesttypes.TriggerManual
	}
	run, err := s.repository.CreateFetchRun(ctx, date, category, triggerType)
	if err != nil {
		return contracts.DailyBestFetchRun{}, err
	}
	s.wg.Add(1)
	go func(runID int64, businessDate string, category int64) {
		defer s.wg.Done()
		s.execute(s.root, runID, businessDate, category)
	}(run.ID, date, category)
	return run, nil
}
