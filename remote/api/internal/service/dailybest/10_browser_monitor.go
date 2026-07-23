package dailybest

import (
	"context"
	"fmt"
	"log"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (s *Service) StartBrowserMonitor(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("FlareSolverr 监控间隔必须为正数")
	}
	if s.agent == nil {
		return fmt.Errorf("FlareSolverr 客户端为空")
	}
	if s.runtimeRepository == nil {
		return fmt.Errorf("DailyBest 运行状态仓库为空")
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.observeBrowserState(ctx)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.observeBrowserState(ctx)
			}
		}
	}()
	return nil
}

func (s *Service) observeBrowserState(ctx context.Context) {
	status, err := s.agent.Status(ctx)
	stateKey := "unavailable"
	level := dailybesttypes.RuntimeLevelError
	eventType := "browser_solver_unavailable"
	message := "FlareSolverr 状态接口不可用"
	if err == nil {
		stateKey = fmt.Sprintf(
			"%s:%s:%s",
			status.State, status.Version, status.UserAgent,
		)
		level = dailybesttypes.RuntimeLevelInfo
		eventType = "browser_solver_state_changed"
		message = fmt.Sprintf(
			"FlareSolverr=%s, 版本=%s, User-Agent=%s",
			status.State, status.Version, status.UserAgent,
		)
	} else {
		message += ": " + err.Error()
	}
	s.runtimeMu.Lock()
	previous := s.browserMonitorState
	if previous == stateKey {
		s.runtimeMu.Unlock()
		return
	}
	s.browserMonitorState = stateKey
	s.runtimeMu.Unlock()
	if _, eventErr := s.runtimeRepository.CreateRuntimeEvent(context.WithoutCancel(ctx), dailybesttypes.RuntimeEventDraft{
		Component:  dailybesttypes.RuntimeComponentBrowserAgent,
		EventType:  eventType,
		Level:      level,
		Message:    message,
		OccurredAt: time.Now().Unix(),
	}); eventErr != nil {
		log.Printf("FlareSolverr 状态事件写入失败: %v", eventErr)
	}
}
