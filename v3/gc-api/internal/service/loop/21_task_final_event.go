package loop

import "time"

func (l *FetchLoopService) buildManagedTaskFinalEvent(jobID string, taskType string, stage string, message string) JobEvent {
	event := JobEvent{
		Kind:     JobEventKindProgress,
		JobID:    jobID,
		TaskType: taskType,
		Stage:    stage,
		Message:  message,
		Done:     true,
		At:       time.Now().Unix(),
	}
	l.jobs.mu.Lock()
	job := l.jobs.jobs[jobID]
	if job == nil || job.lastProgress == nil {
		l.jobs.mu.Unlock()
		return event
	}

	progress := job.lastProgress
	event.HandledCount = progress.HandledCount
	event.SuccessCount = progress.SuccessCount
	event.FailedCount = progress.FailedCount
	event.QueuedCount = progress.QueuedCount
	event.PendingItems = cloneManagedQueueItems(progress.PendingItems)
	event.RunningItems = cloneManagedQueueItems(progress.RunningItems)
	event.DoneItems = cloneManagedQueueItems(progress.DoneItems)
	event.CurrentPhaseKey = progress.CurrentPhaseKey
	event.PhaseKey = progress.PhaseKey
	event.PhaseHandledCount = progress.PhaseHandledCount
	event.PhaseTotalCount = progress.PhaseTotalCount
	event.PhaseSuccessCount = progress.PhaseSuccessCount
	event.PhaseFailedCount = progress.PhaseFailedCount
	event.PhaseStats = cloneManagedPhaseStats(job.phaseStats)
	event.StartedAt = progress.StartedAt
	l.jobs.mu.Unlock()
	return event
}
