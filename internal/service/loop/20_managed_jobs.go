package loop

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type JobProgress struct {
	JobID        string `json:"job_id"`
	TaskType     string `json:"task_type"`
	Stage        string `json:"stage"`
	Message      string `json:"message"`
	HandledCount int    `json:"handled_count"`
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
	QueuedCount  int    `json:"queued_count"`
	StartedAt    int64  `json:"started_at"`
	At           int64  `json:"at"`
	Done         bool   `json:"done"`
}

type JobSnapshot struct {
	JobID        string `json:"job_id"`
	TaskType     string `json:"task_type"`
	StartedAt    int64  `json:"started_at"`
	Done         bool   `json:"done"`
	Paused       bool   `json:"paused"`
	Stage        string `json:"stage"`
	Message      string `json:"message"`
	HandledCount int    `json:"handled_count"`
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
	QueuedCount  int    `json:"queued_count"`
	At           int64  `json:"at"`
}

type managedJob struct {
	id          string
	taskType    string
	startedAt   int64
	lastEvent   *JobProgress
	done        bool
	paused      bool
	pauseCh     chan struct{}
	finalEvent  *JobProgress
	cancel      func()
	subscribers map[chan JobProgress]struct{}
}

type managedProgressEventPusher func(chan JobProgress, JobProgress)

type managedProgressJobManager struct {
	prefix          string
	subscribeBuffer int
	push            managedProgressEventPusher

	seed uint64
	mu   sync.Mutex
	jobs map[string]*managedJob
}

func newManagedProgressJobManager(prefix string, subscribeBuffer int, push managedProgressEventPusher) *managedProgressJobManager {
	if subscribeBuffer <= 0 {
		subscribeBuffer = 64
	}
	if push == nil {
		push = pushManagedProgressEvent
	}
	return &managedProgressJobManager{
		prefix:          prefix,
		subscribeBuffer: subscribeBuffer,
		push:            push,
		jobs:            make(map[string]*managedJob),
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
		subscribers: make(map[chan JobProgress]struct{}),
	}
	m.mu.Unlock()
	return id
}

func (m *managedProgressJobManager) subscribe(jobID string) (<-chan JobProgress, func(), error) {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return nil, nil, fmt.Errorf("job_id not found")
	}

	ch := make(chan JobProgress, m.subscribeBuffer)
	if job.done {
		finalEvent := job.finalEvent
		m.mu.Unlock()
		if finalEvent != nil {
			ch <- *finalEvent
		}
		return ch, func() {}, nil
	}

	job.subscribers[ch] = struct{}{}
	lastEvent := job.lastEvent
	m.mu.Unlock()

	if lastEvent != nil {
		ch <- *lastEvent
	}

	cancel := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if current, ok := m.jobs[jobID]; ok {
			delete(current.subscribers, ch)
		}
	}
	return ch, cancel, nil
}

func (m *managedProgressJobManager) publish(jobID string, event JobProgress) {
	m.mu.Lock()
	if job, ok := m.jobs[jobID]; ok && !job.done {
		cloned := event
		job.lastEvent = &cloned
	}
	m.mu.Unlock()

	subs := m.snapshotSubscribers(jobID)
	for _, ch := range subs {
		m.push(ch, event)
	}
}

func (m *managedProgressJobManager) finish(jobID string, event JobProgress) {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return
	}

	job.done = true
	job.finalEvent = &event
	job.cancel = nil
	if job.pauseCh != nil {
		close(job.pauseCh)
		job.pauseCh = nil
	}
	job.paused = false

	subs := make([]chan JobProgress, 0, len(job.subscribers))
	for ch := range job.subscribers {
		subs = append(subs, ch)
	}
	job.subscribers = make(map[chan JobProgress]struct{})
	m.mu.Unlock()

	for _, ch := range subs {
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

func (m *managedProgressJobManager) pause(jobID string) (*JobProgress, error) {
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
		if job.lastEvent == nil {
			return nil, nil
		}
		cloned := *job.lastEvent
		return &cloned, nil
	}
	job.paused = true
	job.pauseCh = make(chan struct{})

	event := JobProgress{
		JobID:     job.id,
		TaskType:  job.taskType,
		StartedAt: job.startedAt,
		Stage:     "paused",
		Message:   "任务已暂停",
		At:        time.Now().Unix(),
	}
	if job.lastEvent != nil {
		event = *job.lastEvent
		event.Stage = "paused"
		event.Message = "任务已暂停"
		event.Done = false
		event.At = time.Now().Unix()
	}
	job.lastEvent = &event
	cloned := event
	return &cloned, nil
}

func (m *managedProgressJobManager) resume(jobID string) (*JobProgress, error) {
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
		if job.lastEvent == nil {
			return nil, nil
		}
		cloned := *job.lastEvent
		return &cloned, nil
	}
	if job.pauseCh != nil {
		close(job.pauseCh)
		job.pauseCh = nil
	}
	job.paused = false

	event := JobProgress{
		JobID:     job.id,
		TaskType:  job.taskType,
		StartedAt: job.startedAt,
		Stage:     "resumed",
		Message:   "任务已继续",
		At:        time.Now().Unix(),
	}
	if job.lastEvent != nil {
		event = *job.lastEvent
		event.Stage = "resumed"
		event.Message = "任务已继续"
		event.Done = false
		event.At = time.Now().Unix()
	}
	job.lastEvent = &event
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

func (m *managedProgressJobManager) snapshotSubscribers(jobID string) []chan JobProgress {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[jobID]
	if !ok || job.done {
		return nil
	}
	subs := make([]chan JobProgress, 0, len(job.subscribers))
	for ch := range job.subscribers {
		subs = append(subs, ch)
	}
	return subs
}

func pushManagedProgressEvent(ch chan JobProgress, event JobProgress) {
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
		JobID:     job.id,
		TaskType:  job.taskType,
		StartedAt: job.startedAt,
		Done:      job.done,
		Paused:    job.paused,
	}

	event := job.lastEvent
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
	}

	return item
}
