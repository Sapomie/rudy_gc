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
	event := JobEvent{
		Kind:              JobEventKindProgress,
		JobID:             jobID,
		TaskType:          taskType,
		Stage:             progress.Stage,
		Message:           progress.Message,
		HandledCount:      progress.HandledCount,
		SuccessCount:      progress.SuccessCount,
		FailedCount:       progress.FailedCount,
		QueuedCount:       progress.QueuedCount,
		CurrentPhaseKey:   progress.CurrentPhaseKey,
		PhaseKey:          progress.PhaseKey,
		PhaseHandledCount: progress.PhaseHandledCount,
		PhaseTotalCount:   progress.PhaseTotalCount,
		PhaseSuccessCount: progress.PhaseSuccessCount,
		PhaseFailedCount:  progress.PhaseFailedCount,
		At:                time.Now().Unix(),
	}
	if snapshot := l.findJob(jobID); snapshot != nil {
		event.StartedAt = snapshot.startedAt
	}
	l.jobs.publish(jobID, event)
}

func (l *FetchLoopService) publishTaskLog(jobID, taskType string, logEvent taskctx.Log) {
	if logEvent.Message == "" && logEvent.Line == "" {
		return
	}

	event := JobEvent{
		Kind:      JobEventKindLog,
		JobID:     jobID,
		TaskType:  taskType,
		Message:   logEvent.Message,
		Level:     logEvent.Level,
		Line:      logEvent.Line,
		At:        logEvent.At,
		StartedAt: 0,
	}
	if snapshot := l.findJob(jobID); snapshot != nil {
		event.StartedAt = snapshot.startedAt
	}
	l.jobs.publish(jobID, event)
}

func (l *FetchLoopService) findJob(jobID string) *managedJob {
	l.jobs.mu.Lock()
	defer l.jobs.mu.Unlock()
	return l.jobs.jobs[jobID]
}
