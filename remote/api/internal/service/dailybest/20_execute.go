package dailybest

import (
	"context"
	"log"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (s *Service) execute(ctx context.Context, runID int64, businessDate string, category int64, triggerType string) {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	startedAt := time.Now().Unix()
	if err := s.repository.MarkFetchRunRunning(ctx, runID, startedAt); err != nil {
		log.Printf("DailyBest 抓取任务 %d 标记运行失败: %v", runID, err)
		return
	}
	if err := s.collectAndPublish(ctx, runID, businessDate, category); err != nil {
		if markErr := s.repository.MarkFetchRunFailed(context.WithoutCancel(ctx), runID, time.Now().Unix(), err.Error()); markErr != nil {
			log.Printf("DailyBest 抓取任务 %d 标记失败异常: %v；原始错误: %v", runID, markErr, err)
		}
	}
	if triggerType == dailybesttypes.TriggerScheduled || triggerType == dailybesttypes.TriggerRecovery {
		s.signalRecovery()
	}
}
