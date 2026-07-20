package dailybestcollector

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"rudy_gc/remote/api/internal/modules/dailybest/domain"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) StartScheduler(ctx context.Context, dailyFetchTime string, fetchOnStart bool) error {
	hour, minute, err := parseDailyFetchTime(dailyFetchTime)
	if err != nil {
		return err
	}
	if fetchOnStart {
		_, _ = s.Enqueue(ctx, contracts.DailyBestFetchRequest{}, dailybest.TriggerScheduled)
	}
	go func() {
		for {
			now := time.Now().In(s.loc)
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, s.loc)
			if !next.After(now) {
				next = next.AddDate(0, 0, 1)
			}
			timer := time.NewTimer(time.Until(next))
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return
			case <-timer.C:
				_, _ = s.Enqueue(ctx, contracts.DailyBestFetchRequest{}, dailybest.TriggerScheduled)
			}
		}
	}()
	return nil
}

func parseDailyFetchTime(value string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("DailyFetchTime 必须使用 HH:MM 格式")
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("DailyFetchTime 小时无效")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("DailyFetchTime 分钟无效")
	}
	return hour, minute, nil
}
