package loop

import (
	"context"
	"fmt"
	"time"

	"rudy_gc/internal/taskctx"
)

func (l *FetchLoopService) currentRootContext() context.Context {
	l.rootMu.Lock()
	defer l.rootMu.Unlock()
	if l.rootCtx != nil {
		return l.rootCtx
	}
	return context.Background()
}

func (l *FetchLoopService) beginExclusiveTask(jobID, taskType string) error {
	l.exclusiveMu.Lock()
	defer l.exclusiveMu.Unlock()
	if l.exclusiveJobID != "" {
		return fmt.Errorf("已有运行中任务：%s", l.exclusiveTaskType)
	}
	l.exclusiveJobID = jobID
	l.exclusiveTaskType = taskType
	return nil
}

func (l *FetchLoopService) finishExclusiveTask(jobID string) {
	l.exclusiveMu.Lock()
	defer l.exclusiveMu.Unlock()
	if l.exclusiveJobID == jobID {
		l.exclusiveJobID = ""
		l.exclusiveTaskType = ""
	}
}

func (l *FetchLoopService) publishTaskProgress(jobID, taskType string, progress taskctx.Progress) {
	event := JobProgress{
		JobID:        jobID,
		TaskType:     taskType,
		Stage:        progress.Stage,
		Message:      progress.Message,
		HandledCount: progress.HandledCount,
		SuccessCount: progress.SuccessCount,
		FailedCount:  progress.FailedCount,
		QueuedCount:  progress.QueuedCount,
		At:           time.Now().Unix(),
	}
	if snapshot := l.findJob(jobID); snapshot != nil {
		event.StartedAt = snapshot.startedAt
		if progress.HandledCount == 0 && snapshot.lastEvent != nil {
			event.HandledCount = snapshot.lastEvent.HandledCount
		}
		if progress.SuccessCount == 0 && snapshot.lastEvent != nil {
			event.SuccessCount = snapshot.lastEvent.SuccessCount
		}
		if progress.FailedCount == 0 && snapshot.lastEvent != nil {
			event.FailedCount = snapshot.lastEvent.FailedCount
		}
		if progress.QueuedCount == 0 && snapshot.lastEvent != nil {
			event.QueuedCount = snapshot.lastEvent.QueuedCount
		}
	}
	l.jobs.publish(jobID, event)
}

func (l *FetchLoopService) findJob(jobID string) *managedJob {
	l.jobs.mu.Lock()
	defer l.jobs.mu.Unlock()
	return l.jobs.jobs[jobID]
}
