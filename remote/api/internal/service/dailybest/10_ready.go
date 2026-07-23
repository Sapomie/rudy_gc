package dailybest

import (
	"context"
	"fmt"

	browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"
)

func (s *Service) Ready(ctx context.Context) error {
	if err := s.repository.Ping(ctx); err != nil {
		return err
	}
	if s.agent == nil {
		return fmt.Errorf("FlareSolverr 客户端为空")
	}
	status, err := s.agent.Status(ctx)
	if err != nil {
		return err
	}
	if status.State != browseragenttypes.StateReady && status.State != browseragenttypes.StateFetching {
		return fmt.Errorf("FlareSolverr 当前状态为 %s", status.State)
	}
	return nil
}
