package loop

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	JobEventKindProgress  = "progress"
	JobEventKindLog       = "log"
	managedJobHistorySize = 300
)

type JobEvent struct {
	ID                int64                   `json:"id"`
	Kind              string                  `json:"kind"`
	JobID             string                  `json:"job_id"`
	TaskType          string                  `json:"task_type"`
	Stage             string                  `json:"stage,omitempty"`
	Message           string                  `json:"message,omitempty"`
	HandledCount      int                     `json:"handled_count,omitempty"`
	SuccessCount      int                     `json:"success_count,omitempty"`
	FailedCount       int                     `json:"failed_count,omitempty"`
	QueuedCount       int                     `json:"queued_count,omitempty"`
	CurrentPhaseKey   string                  `json:"current_phase_key,omitempty"`
	PhaseKey          string                  `json:"phase_key,omitempty"`
	PhaseHandledCount int                     `json:"phase_handled_count,omitempty"`
	PhaseTotalCount   int                     `json:"phase_total_count,omitempty"`
	PhaseSuccessCount int                     `json:"phase_success_count,omitempty"`
	PhaseFailedCount  int                     `json:"phase_failed_count,omitempty"`
	PhaseStats        map[string]JobPhaseStat `json:"phase_stats,omitempty"`
	Level             string                  `json:"level,omitempty"`
	Line              string                  `json:"line,omitempty"`
	StartedAt         int64                   `json:"started_at"`
	At                int64                   `json:"at"`
	Done              bool                    `json:"done"`
}

type JobPhaseStat struct {
	HandledCount int `json:"handled_count"`
	TotalCount   int `json:"total_count"`
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
}

type JobSnapshot struct {
	JobID           string                  `json:"job_id"`
	TaskType        string                  `json:"task_type"`
	StartedAt       int64                   `json:"started_at"`
	Done            bool                    `json:"done"`
	Paused          bool                    `json:"paused"`
	Stage           string                  `json:"stage"`
	Message         string                  `json:"message"`
	HandledCount    int                     `json:"handled_count"`
	SuccessCount    int                     `json:"success_count"`
	FailedCount     int                     `json:"failed_count"`
	QueuedCount     int                     `json:"queued_count"`
	CurrentPhaseKey string                  `json:"current_phase_key,omitempty"`
	PhaseStats      map[string]JobPhaseStat `json:"phase_stats,omitempty"`
	At              int64                   `json:"at"`
}

type managedJob struct {
	id              string
	taskType        string
	startedAt       int64
	lastProgress    *JobEvent
	done            bool
	paused          bool
	pauseCh         chan struct{}
	finalEvent      *JobEvent
	cancel          func()
	currentPhaseKey string
	phaseStats      map[string]JobPhaseStat
	history         []JobEvent
	subscribers     map[chan JobEvent]struct{}
}

type managedProgressEventPusher func(chan JobEvent, JobEvent)

type managedProgressJobManager struct {
	prefix          string
	subscribeBuffer int
	push            managedProgressEventPusher

	seed uint64
	mu   sync.Mutex
	jobs map[string]*managedJob

	eventSeq          int64
	globalHistory     []JobEvent
	globalSubscribers map[chan JobEvent]struct{}
}

func newManagedProgressJobManager(prefix string, subscribeBuffer int, push managedProgressEventPusher) *managedProgressJobManager {
	if subscribeBuffer <= 0 {
		subscribeBuffer = 64
	}
	if push == nil {
		push = pushManagedProgressEvent
	}
	return &managedProgressJobManager{
		prefix:            prefix,
		subscribeBuffer:   subscribeBuffer,
		push:              push,
		jobs:              make(map[string]*managedJob),
		globalHistory:     make([]JobEvent, 0, managedJobHistorySize),
		globalSubscribers: make(map[chan JobEvent]struct{}),
	}
}

func (m *managedProgressJobManager) create(taskType string) string {
	id := fmt.Sprintf("%s_%d_%d", m.prefix, time.Now().UnixNano(), atomic.AddUint64(&m.seed, 1))
	now := time.Now().Unix()
	m.mu.Lock()
	m.jobs[id] = &managedJob{
		id:          id,
		taskType:    taskType,
		startedAt:   now,
		phaseStats:  make(map[string]JobPhaseStat),
		history:     make([]JobEvent, 0, managedJobHistorySize),
		subscribers: make(map[chan JobEvent]struct{}),
	}
	m.mu.Unlock()
	return id
}

func (m *managedProgressJobManager) subscribe(jobID string, afterID int64) ([]JobEvent, <-chan JobEvent, func(), error) {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return nil, nil, nil, fmt.Errorf("job_id not found")
	}

	history := cloneManagedJobEventsAfter(job.history, afterID)
	ch := make(chan JobEvent, m.subscribeBuffer)
	if job.done {
		close(ch)
		m.mu.Unlock()
		return history, ch, func() {}, nil
	}

	job.subscribers[ch] = struct{}{}
	m.mu.Unlock()

	cancel := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if current, ok := m.jobs[jobID]; ok {
			delete(current.subscribers, ch)
		}
	}
	return history, ch, cancel, nil
}

func (m *managedProgressJobManager) subscribeAll(afterID int64) ([]JobEvent, <-chan JobEvent, func()) {
	m.mu.Lock()
	history := cloneManagedJobEventsAfter(m.globalHistory, afterID)
	ch := make(chan JobEvent, m.subscribeBuffer)
	m.globalSubscribers[ch] = struct{}{}
	m.mu.Unlock()

	cancel := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.globalSubscribers, ch)
	}
	return history, ch, cancel
}

func (m *managedProgressJobManager) publish(jobID string, event JobEvent) {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok || job.done {
		m.mu.Unlock()
		return
	}
	event = prepareManagedJobEvent(job, event)
	event = applyManagedJobPhaseEvent(job, event)
	event = assignManagedJobEventID(m, event)
	recordManagedJobEvent(job, event)
	recordManagedGlobalEvent(m, event)
	subs := snapshotManagedJobSubscribers(job)
	globalSubs := snapshotManagedGlobalSubscribers(m)
	m.mu.Unlock()

	for _, ch := range subs {
		m.push(ch, event)
	}
	for _, ch := range globalSubs {
		m.push(ch, event)
	}
}

func (m *managedProgressJobManager) finish(jobID string, event JobEvent) {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return
	}

	job.done = true
	job.cancel = nil
	if job.pauseCh != nil {
		close(job.pauseCh)
		job.pauseCh = nil
	}
	job.paused = false

	event.Done = true
	event = prepareManagedJobEvent(job, event)
	event = applyManagedJobPhaseEvent(job, event)
	event = assignManagedJobEventID(m, event)
	recordManagedJobEvent(job, event)
	recordManagedGlobalEvent(m, event)
	if event.Kind == JobEventKindProgress {
		cloned := event
		job.finalEvent = &cloned
	}

	subs := snapshotManagedJobSubscribers(job)
	globalSubs := snapshotManagedGlobalSubscribers(m)
	job.subscribers = make(map[chan JobEvent]struct{})
	m.mu.Unlock()

	for _, ch := range subs {
		m.push(ch, event)
	}
	for _, ch := range globalSubs {
		m.push(ch, event)
	}

	time.AfterFunc(2*time.Minute, func() {
		m.mu.Lock()
		delete(m.jobs, jobID)
		m.mu.Unlock()
	})
}

func (m *managedProgressJobManager) setCancel(jobID string, cancel func()) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[jobID]
	if !ok {
		return fmt.Errorf("job_id not found")
	}
	if job.done {
		return fmt.Errorf("job already finished")
	}
	job.cancel = cancel
	return nil
}

func (m *managedProgressJobManager) clearCancel(jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.jobs[jobID]; ok {
		job.cancel = nil
	}
}

func (m *managedProgressJobManager) stop(jobID string) error {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("job_id not found")
	}
	if job.done {
		m.mu.Unlock()
		return fmt.Errorf("job already finished")
	}
	cancel := job.cancel
	if job.pauseCh != nil {
		close(job.pauseCh)
		job.pauseCh = nil
	}
	job.paused = false
	if cancel == nil {
		m.mu.Unlock()
		return fmt.Errorf("job is not cancellable")
	}
	job.cancel = nil
	m.mu.Unlock()

	cancel()
	return nil
}

func (m *managedProgressJobManager) pause(jobID string) (*JobEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job_id not found")
	}
	if job.done {
		return nil, fmt.Errorf("job already finished")
	}
	if job.paused {
		if job.lastProgress == nil {
			return nil, nil
		}
		cloned := *job.lastProgress
		return &cloned, nil
	}
	job.paused = true
	job.pauseCh = make(chan struct{})

	event := JobEvent{
		Kind:      JobEventKindProgress,
		JobID:     job.id,
		TaskType:  job.taskType,
		StartedAt: job.startedAt,
		Stage:     "paused",
		Message:   "任务已暂停",
		At:        time.Now().Unix(),
	}
	if job.lastProgress != nil {
		event = *job.lastProgress
		event.Stage = "paused"
		event.Message = "任务已暂停"
		event.Done = false
		event.At = time.Now().Unix()
	}
	cloned := event
	return &cloned, nil
}

func (m *managedProgressJobManager) resume(jobID string) (*JobEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job_id not found")
	}
	if job.done {
		return nil, fmt.Errorf("job already finished")
	}
	if !job.paused {
		if job.lastProgress == nil {
			return nil, nil
		}
		cloned := *job.lastProgress
		return &cloned, nil
	}
	if job.pauseCh != nil {
		close(job.pauseCh)
		job.pauseCh = nil
	}
	job.paused = false

	event := JobEvent{
		Kind:      JobEventKindProgress,
		JobID:     job.id,
		TaskType:  job.taskType,
		StartedAt: job.startedAt,
		Stage:     "resumed",
		Message:   "任务已继续",
		At:        time.Now().Unix(),
	}
	if job.lastProgress != nil {
		event = *job.lastProgress
		event.Stage = "resumed"
		event.Message = "任务已继续"
		event.Done = false
		event.At = time.Now().Unix()
	}
	cloned := event
	return &cloned, nil
}

func (m *managedProgressJobManager) waitIfPaused(ctx context.Context, jobID string) error {
	for {
		m.mu.Lock()
		job, ok := m.jobs[jobID]
		if !ok || job.done || !job.paused || job.pauseCh == nil {
			m.mu.Unlock()
			return nil
		}
		pauseCh := job.pauseCh
		m.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pauseCh:
		}
	}
}

func (m *managedProgressJobManager) listRunning() []JobSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := make([]JobSnapshot, 0, len(m.jobs))
	for _, job := range m.jobs {
		if job.done {
			continue
		}
		list = append(list, snapshotManagedJob(job))
	}

	sort.SliceStable(list, func(i, j int) bool {
		if list[i].StartedAt != list[j].StartedAt {
			return list[i].StartedAt > list[j].StartedAt
		}
		return list[i].JobID > list[j].JobID
	})
	return list
}

func (m *managedProgressJobManager) listAll() []JobSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := make([]JobSnapshot, 0, len(m.jobs))
	for _, job := range m.jobs {
		list = append(list, snapshotManagedJob(job))
	}

	sort.SliceStable(list, func(i, j int) bool {
		if list[i].StartedAt != list[j].StartedAt {
			return list[i].StartedAt > list[j].StartedAt
		}
		return list[i].JobID > list[j].JobID
	})
	return list
}

func (m *managedProgressJobManager) snapshot(jobID string) (JobSnapshot, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return JobSnapshot{}, false
	}
	return snapshotManagedJob(job), true
}

func (m *managedProgressJobManager) history(jobID string, afterID int64) ([]JobEvent, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, false
	}
	return cloneManagedJobEventsAfter(job.history, afterID), true
}

func prepareManagedJobEvent(job *managedJob, event JobEvent) JobEvent {
	if job == nil {
		return event
	}
	if event.Kind == "" {
		event.Kind = JobEventKindProgress
	}
	if event.JobID == "" {
		event.JobID = job.id
	}
	if event.TaskType == "" {
		event.TaskType = job.taskType
	}
	if event.StartedAt == 0 {
		event.StartedAt = job.startedAt
	}
	if event.At == 0 {
		event.At = time.Now().Unix()
	}
	return event
}

func applyManagedJobPhaseEvent(job *managedJob, event JobEvent) JobEvent {
	if job == nil {
		return event
	}
	if event.CurrentPhaseKey == "" && event.PhaseKey != "" {
		event.CurrentPhaseKey = event.PhaseKey
	}
	if event.CurrentPhaseKey != "" {
		job.currentPhaseKey = event.CurrentPhaseKey
	} else if job.currentPhaseKey != "" {
		event.CurrentPhaseKey = job.currentPhaseKey
	}
	if event.PhaseKey != "" {
		if job.phaseStats == nil {
			job.phaseStats = make(map[string]JobPhaseStat)
		}
		job.phaseStats[event.PhaseKey] = JobPhaseStat{
			HandledCount: event.PhaseHandledCount,
			TotalCount:   event.PhaseTotalCount,
			SuccessCount: event.PhaseSuccessCount,
			FailedCount:  event.PhaseFailedCount,
		}
	}
	event.PhaseStats = cloneManagedPhaseStats(job.phaseStats)
	return event
}

func assignManagedJobEventID(m *managedProgressJobManager, event JobEvent) JobEvent {
	if m == nil {
		return event
	}
	if event.ID == 0 {
		m.eventSeq++
		event.ID = m.eventSeq
	}
	return event
}

func cloneManagedPhaseStats(in map[string]JobPhaseStat) map[string]JobPhaseStat {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]JobPhaseStat, len(in))
	for key, item := range in {
		out[key] = item
	}
	return out
}

func recordManagedJobEvent(job *managedJob, event JobEvent) {
	if job == nil {
		return
	}
	job.history = append(job.history, event)
	if len(job.history) > managedJobHistorySize {
		job.history = append([]JobEvent(nil), job.history[len(job.history)-managedJobHistorySize:]...)
	}
	if event.Kind == JobEventKindProgress {
		cloned := event
		job.lastProgress = &cloned
	}
}

func recordManagedGlobalEvent(m *managedProgressJobManager, event JobEvent) {
	if m == nil {
		return
	}
	m.globalHistory = append(m.globalHistory, event)
	if len(m.globalHistory) > managedJobHistorySize {
		m.globalHistory = append([]JobEvent(nil), m.globalHistory[len(m.globalHistory)-managedJobHistorySize:]...)
	}
}

func snapshotManagedJobSubscribers(job *managedJob) []chan JobEvent {
	if job == nil || job.done {
		return nil
	}
	subs := make([]chan JobEvent, 0, len(job.subscribers))
	for ch := range job.subscribers {
		subs = append(subs, ch)
	}
	return subs
}

func snapshotManagedGlobalSubscribers(m *managedProgressJobManager) []chan JobEvent {
	if m == nil {
		return nil
	}
	subs := make([]chan JobEvent, 0, len(m.globalSubscribers))
	for ch := range m.globalSubscribers {
		subs = append(subs, ch)
	}
	return subs
}

func cloneManagedJobEventsAfter(list []JobEvent, afterID int64) []JobEvent {
	if len(list) == 0 {
		return nil
	}
	if afterID <= 0 {
		return append([]JobEvent(nil), list...)
	}
	filtered := make([]JobEvent, 0, len(list))
	for _, item := range list {
		if item.ID <= afterID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func pushManagedProgressEvent(ch chan JobEvent, event JobEvent) {
	if ch == nil {
		return
	}
	select {
	case ch <- event:
	default:
	}
}

func snapshotManagedJob(job *managedJob) JobSnapshot {
	item := JobSnapshot{
		JobID:           job.id,
		TaskType:        job.taskType,
		StartedAt:       job.startedAt,
		Done:            job.done,
		Paused:          job.paused,
		CurrentPhaseKey: job.currentPhaseKey,
		PhaseStats:      cloneManagedPhaseStats(job.phaseStats),
	}

	event := job.lastProgress
	if job.done && job.finalEvent != nil {
		event = job.finalEvent
	}
	if event != nil {
		item.Stage = event.Stage
		item.Message = event.Message
		item.HandledCount = event.HandledCount
		item.SuccessCount = event.SuccessCount
		item.FailedCount = event.FailedCount
		item.QueuedCount = event.QueuedCount
		item.At = event.At
		if item.CurrentPhaseKey == "" {
			item.CurrentPhaseKey = event.CurrentPhaseKey
		}
		if len(item.PhaseStats) == 0 && len(event.PhaseStats) > 0 {
			item.PhaseStats = cloneManagedPhaseStats(event.PhaseStats)
		}
	}

	return item
}
