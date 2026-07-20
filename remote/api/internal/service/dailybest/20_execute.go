package dailybest

import (
	"context"
	"time"
)

func (s *Service) execute(ctx context.Context, runID int64, businessDate string, category int64) {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	startedAt := time.Now().Unix()
	if err := s.repository.MarkFetchRunRunning(ctx, runID, startedAt); err != nil {
		return
	}
	if err := s.collectAndPublish(ctx, runID, businessDate, category); err != nil {
		_ = s.repository.MarkFetchRunFailed(context.WithoutCancel(ctx), runID, time.Now().Unix(), err.Error())
	}
}
