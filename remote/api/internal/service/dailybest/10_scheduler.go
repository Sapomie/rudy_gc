package dailybest

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) StartScheduler(ctx context.Context, dailyFetchTime string) error {
	hour, minute, err := parseDailyFetchTime(dailyFetchTime)
	if err != nil {
		return err
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
				_, _ = s.enqueueAutomatic(ctx, dailybesttypes.TriggerScheduled)
				s.signalRecovery()
			}
		}
	}()
	return nil
}

func (s *Service) StartFailureRecovery(ctx context.Context, retryDelay time.Duration) error {
	if retryDelay <= 0 {
		return fmt.Errorf("DailyBest 失败恢复时间必须为正数")
	}
	if s.schedulerRepository == nil {
		return fmt.Errorf("DailyBest 调度仓库为空")
	}
	s.retryDelay = retryDelay
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runFailureRecovery(ctx)
	}()
	s.signalRecovery()
	return nil
}

func (s *Service) runFailureRecovery(ctx context.Context) {
	for {
		date := s.CurrentBusinessDate(time.Now())
		state, err := s.schedulerRepository.FindAutomaticFetchState(ctx, date, contracts.DailyBestCategoryMonth)
		if err != nil {
			if !s.waitForRecovery(ctx, s.retryDelay) {
				return
			}
			continue
		}
		if state.Successful != nil || state.Active != nil || state.Failed == nil {
			if !s.waitForRecovery(ctx, 0) {
				return
			}
			continue
		}
		retryAt := time.Unix(state.Failed.FinishedAt, 0).Add(s.retryDelay)
		if wait := time.Until(retryAt); wait > 0 {
			if !s.waitForRecovery(ctx, wait) {
				return
			}
			continue
		}
		_, _ = s.enqueueAutomatic(ctx, dailybesttypes.TriggerRecovery)
		if !s.waitForRecovery(ctx, 0) {
			return
		}
	}
}

func (s *Service) waitForRecovery(ctx context.Context, wait time.Duration) bool {
	if wait <= 0 {
		select {
		case <-ctx.Done():
			return false
		case <-s.retryWake:
			return true
		}
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-s.retryWake:
		return true
	case <-timer.C:
		return true
	}
}

func (s *Service) signalRecovery() {
	select {
	case s.retryWake <- struct{}{}:
	default:
	}
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

func (s *Service) Wait() {
	s.wg.Wait()
}
