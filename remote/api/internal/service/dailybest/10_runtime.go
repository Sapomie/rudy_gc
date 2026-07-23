package dailybest

import (
	"context"
	"errors"
	"fmt"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (s *Service) PrepareRuntime(ctx context.Context) error {
	if s.runtimeRepository == nil {
		return fmt.Errorf("DailyBest 运行状态仓库为空")
	}
	now := time.Now().Unix()
	interruptedRuns, interruptedLogs, err := s.repository.InterruptActiveFetches(
		ctx, now, "Remote API 重启时结束上一进程遗留的抓取任务",
	)
	if err != nil {
		return err
	}
	if interruptedRuns > 0 || interruptedLogs > 0 {
		if _, err := s.runtimeRepository.CreateRuntimeEvent(ctx, dailybesttypes.RuntimeEventDraft{
			Component: dailybesttypes.RuntimeComponentAPI,
			EventType: "fetch_interrupted",
			Level:     dailybesttypes.RuntimeLevelError,
			Message: fmt.Sprintf(
				"Remote API 重启时结束了 %d 个遗留抓取任务和 %d 条遗留请求日志",
				interruptedRuns, interruptedLogs,
			),
			OccurredAt: now,
		}); err != nil {
			return err
		}
	}
	interrupted, err := s.runtimeRepository.InterruptRunningTargetProbes(ctx, now, "Remote API 重启时结束上一进程遗留的探测记录")
	if err != nil {
		return err
	}
	if interrupted > 0 {
		if _, err := s.runtimeRepository.CreateRuntimeEvent(ctx, dailybesttypes.RuntimeEventDraft{
			Component:  dailybesttypes.RuntimeComponentAPI,
			EventType:  "target_probe_interrupted",
			Level:      dailybesttypes.RuntimeLevelError,
			Message:    "Remote API 重启时结束了遗留的运行中目标站探测",
			OccurredAt: now,
		}); err != nil {
			return err
		}
	}
	latest, err := s.runtimeRepository.FindLatestTargetProbe(ctx)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if err == nil {
		s.runtimeMu.Lock()
		s.latestProbe = &latest
		if latest.Status == dailybesttypes.TargetProbeStatusSucceeded {
			s.targetStatus = dailybesttypes.TargetStatusAvailable
			s.lastProbeSuccess = latest.FinishedAt
		} else {
			s.targetStatus = dailybesttypes.TargetStatusUnavailable
			s.lastProbeFailure = latest.FinishedAt
		}
		s.runtimeMu.Unlock()
	}
	return nil
}

func (s *Service) RuntimeStatus(ctx context.Context) dailybesttypes.RuntimeStatus {
	status := dailybesttypes.RuntimeStatus{APIStatus: "ready", DatabaseStatus: "ready"}
	if err := s.repository.Ping(ctx); err != nil {
		status.DatabaseStatus = "unavailable"
	}
	if s.agent != nil {
		browserStatus, err := s.agent.Status(ctx)
		status.BrowserStatus = browserStatus
		if err != nil {
			status.BrowserError = err.Error()
		}
	}
	s.runtimeMu.Lock()
	status.TargetStatus = s.targetStatus
	status.NextProbeAt = s.nextProbeAt
	status.LastProbeSuccess = s.lastProbeSuccess
	status.LastProbeFailure = s.lastProbeFailure
	if s.latestProbe != nil {
		latest := *s.latestProbe
		status.LatestProbe = &latest
	}
	s.runtimeMu.Unlock()
	return status
}

func (s *Service) ListTargetProbes(ctx context.Context, filter dailybesttypes.TargetProbeFilter) ([]dailybesttypes.TargetProbe, int64, error) {
	if s.runtimeRepository == nil {
		return nil, 0, fmt.Errorf("DailyBest 运行状态仓库为空")
	}
	return s.runtimeRepository.ListTargetProbes(ctx, filter)
}

func (s *Service) ListRuntimeEvents(ctx context.Context, filter dailybesttypes.RuntimeEventFilter) ([]dailybesttypes.RuntimeEvent, int64, error) {
	if s.runtimeRepository == nil {
		return nil, 0, fmt.Errorf("DailyBest 运行状态仓库为空")
	}
	return s.runtimeRepository.ListRuntimeEvents(ctx, filter)
}
