package loop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/taskctx"
)

type exclusiveTaskSlot struct {
	JobID    string
	TaskType string
}

type taskRuntimePolicy struct {
	ExclusiveGroup         string
	PauseDetailLoop        bool
	RegistersRefreshOldest bool
	PreemptsRefreshOldest  bool
}

func (l *FetchLoopService) currentRootContext() context.Context {
	l.rootMu.Lock()
	defer l.rootMu.Unlock()
	if l.rootCtx != nil {
		return l.rootCtx
	}
	return context.Background()
}

func (l *FetchLoopService) beginTaskRuntime(jobID, taskType string, policy taskRuntimePolicy) error {
	l.exclusiveMu.Lock()
	defer l.exclusiveMu.Unlock()
	group := strings.TrimSpace(policy.ExclusiveGroup)
	if group != "" {
		slot := l.exclusiveGroups[group]
		if slot.JobID != "" {
			return fmt.Errorf("已有运行中任务：%s", slot.TaskType)
		}
		l.exclusiveGroups[group] = exclusiveTaskSlot{
			JobID:    jobID,
			TaskType: taskType,
		}
	}
	if policy.RegistersRefreshOldest {
		l.refreshOldestJobID = jobID
		l.refreshOldestAutoPause = false
	}
	return nil
}

func (l *FetchLoopService) finishTaskRuntime(jobID string, policy taskRuntimePolicy) {
	l.exclusiveMu.Lock()
	defer l.exclusiveMu.Unlock()
	group := strings.TrimSpace(policy.ExclusiveGroup)
	if group != "" {
		slot := l.exclusiveGroups[group]
		if slot.JobID == jobID {
			delete(l.exclusiveGroups, group)
		}
	}
	if policy.RegistersRefreshOldest && l.refreshOldestJobID == jobID {
		l.refreshOldestJobID = ""
		l.refreshOldestAutoPause = false
	}
}

func (l *FetchLoopService) isExclusiveGroupRunning(group string) bool {
	group = strings.TrimSpace(group)
	if group == "" {
		return false
	}
	l.exclusiveMu.Lock()
	defer l.exclusiveMu.Unlock()
	slot := l.exclusiveGroups[group]
	return slot.JobID != ""
}

func (l *FetchLoopService) pauseRefreshOldestIfRunning() {
	l.exclusiveMu.Lock()
	jobID := l.refreshOldestJobID
	autoPaused := l.refreshOldestAutoPause
	l.exclusiveMu.Unlock()
	if jobID == "" || autoPaused {
		return
	}

	snapshot, ok := l.jobs.snapshot(jobID)
	if !ok || snapshot.Done || snapshot.Paused {
		return
	}
	event, err := l.jobs.pause(jobID)
	if err != nil {
		return
	}

	l.exclusiveMu.Lock()
	if l.refreshOldestJobID == jobID {
		l.refreshOldestAutoPause = true
	}
	l.exclusiveMu.Unlock()

	if event != nil {
		l.jobs.publish(jobID, *event)
	}
}

func (l *FetchLoopService) resumeRefreshOldestIfAutoPaused() {
	l.exclusiveMu.Lock()
	jobID := l.refreshOldestJobID
	autoPaused := l.refreshOldestAutoPause
	l.exclusiveMu.Unlock()
	if jobID == "" || !autoPaused {
		return
	}

	snapshot, ok := l.jobs.snapshot(jobID)
	if !ok || snapshot.Done {
		l.exclusiveMu.Lock()
		if l.refreshOldestJobID == jobID {
			l.refreshOldestAutoPause = false
		}
		l.exclusiveMu.Unlock()
		return
	}
	if !snapshot.Paused {
		l.exclusiveMu.Lock()
		if l.refreshOldestJobID == jobID {
			l.refreshOldestAutoPause = false
		}
		l.exclusiveMu.Unlock()
		return
	}

	event, err := l.jobs.resume(jobID)
	if err != nil {
		return
	}

	l.exclusiveMu.Lock()
	if l.refreshOldestJobID == jobID {
		l.refreshOldestAutoPause = false
	}
	l.exclusiveMu.Unlock()

	if event != nil {
		l.jobs.publish(jobID, *event)
	}
}

func (l *FetchLoopService) applyTaskStartPolicy(jobID string, policy taskRuntimePolicy) {
	if policy.PreemptsRefreshOldest {
		l.pauseRefreshOldestIfRunning()
		return
	}
	if policy.RegistersRefreshOldest && l.isExclusiveGroupRunning(taskGroupFetchPriority) {
		l.pauseRefreshOldestIfRunning()
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
