package loop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/taskctx"
)

const (
	TaskSpiderDailyBest          = "spider_daily_best"
	TaskSpiderDailyBestSync      = "spider_daily_best_sync"
	TaskSpiderSeeds              = "spider_seeds"
	TaskSpiderSeedByName         = "spider_seed_by_name"
	TaskSpiderRefreshOldest      = "spider_refresh_oldest_detail"
	TaskSpiderRebuildCastRank    = "spider_rebuild_cast_rank"
	TaskSpiderRebuildActorRank   = "spider_rebuild_actor_rank"
	TaskSpiderBackfillPerson     = "spider_backfill_person"
	TaskSpiderBackfillRankPeriod = "spider_backfill_rank_period"
	TaskFilmRename               = "film_rename"
	TaskFilmProcess              = "film_process"
	TaskScRebuildStats           = "sc_rebuild_stats"
	TaskScMove                   = "sc_move"
	TaskScAdd                    = "sc_add"
)

const (
	taskGroupFetchPriority = "fetch_priority"
	taskGroupRefreshOldest = "refresh_oldest"
)

type StartTaskRequest struct {
	TaskType       string `json:"task_type"`
	Name           string `json:"name"`
	ActorName      string `json:"actor_name"`
	Number         int64  `json:"number"`
	ScName         string `json:"sc_name"`
	Dir            string `json:"dir"`
	ComeMovieJavID string `json:"come_movie_jav_id"`
	MovieCast      string `json:"movie_cast"`
	Duration       int64  `json:"duration"`
	Fg             string `json:"fg"`
	Vessel         string `json:"vessel"`
	Remarks        string `json:"remarks"`
}

type DetailLoopSnapshot struct {
	Running     bool   `json:"running"`
	Paused      bool   `json:"paused"`
	Buffered    int    `json:"buffered"`
	StartedAt   int64  `json:"started_at"`
	LastEventAt int64  `json:"last_event_at"`
	LastLine    string `json:"last_line"`
}

func (l *FetchLoopService) Start(ctx context.Context) {
	l.rootMu.Lock()
	defer l.rootMu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	l.rootCtx = ctx
	if l.started {
		return
	}
	l.started = true
	l.startDetailLoopLocked(ctx, l.detailBaseWindow, l.detailMaxBatch)
}

func (l *FetchLoopService) Shutdown() {
	l.stopDetailLoop()
}

func (l *FetchLoopService) ListJobs() []JobSnapshot {
	return l.jobs.listRunning()
}

func (l *FetchLoopService) ListAllJobs() []JobSnapshot {
	return l.jobs.listAll()
}

func (l *FetchLoopService) GetJob(jobID string) (JobSnapshot, error) {
	snapshot, ok := l.jobs.snapshot(strings.TrimSpace(jobID))
	if !ok {
		return JobSnapshot{}, fmt.Errorf("job_id not found")
	}
	return snapshot, nil
}

func (l *FetchLoopService) GetJobEvents(jobID string, afterID int64) ([]JobEvent, error) {
	events, ok := l.jobs.history(strings.TrimSpace(jobID), afterID)
	if !ok {
		return nil, fmt.Errorf("job_id not found")
	}
	return events, nil
}

func (l *FetchLoopService) GetDetailLoopSnapshot() DetailLoopSnapshot {
	l.detailMu.Lock()
	snapshot := DetailLoopSnapshot{
		Running:   l.detailCancel != nil,
		Paused:    l.detailPaused,
		Buffered:  len(l.deps.DetailJobs),
		StartedAt: l.detailStartedAt,
	}
	l.detailMu.Unlock()

	if event, ok := l.detailLogs.latest(); ok {
		snapshot.LastEventAt = event.At
		snapshot.LastLine = event.Line
	}
	return snapshot
}

func (l *FetchLoopService) SubscribeJob(jobID string, afterID int64) ([]JobEvent, <-chan JobEvent, func(), error) {
	return l.jobs.subscribe(jobID, afterID)
}

func (l *FetchLoopService) SubscribeAllJobs(afterID int64) ([]JobEvent, <-chan JobEvent, func()) {
	return l.jobs.subscribeAll(afterID)
}

func (l *FetchLoopService) SubscribeDetailLoop() ([]DetailLoopEvent, <-chan DetailLoopEvent, func()) {
	return l.detailLogs.subscribe()
}

func (l *FetchLoopService) StartManagedDetailLoop() DetailLoopSnapshot {
	l.rootMu.Lock()
	rootCtx := l.rootCtx
	l.rootMu.Unlock()
	l.StartDetailLoop(rootCtx, l.detailBaseWindow, l.detailMaxBatch)
	return l.GetDetailLoopSnapshot()
}

func (l *FetchLoopService) StopManagedDetailLoop() DetailLoopSnapshot {
	l.StopDetailLoop()
	return l.GetDetailLoopSnapshot()
}

func (l *FetchLoopService) StopJob(jobID string) error {
	return l.jobs.stop(strings.TrimSpace(jobID))
}

func (l *FetchLoopService) PauseJob(jobID string) error {
	jobID = strings.TrimSpace(jobID)
	event, err := l.jobs.pause(jobID)
	if err != nil {
		return err
	}
	if event != nil {
		l.jobs.publish(jobID, *event)
	}
	return nil
}

func (l *FetchLoopService) ResumeJob(jobID string) error {
	jobID = strings.TrimSpace(jobID)
	if snapshot, ok := l.jobs.snapshot(jobID); ok {
		if snapshot.TaskType == TaskSpiderRefreshOldest && l.isExclusiveGroupRunning(taskGroupFetchPriority) {
			return fmt.Errorf("高优先抓取任务运行中，刷新最久详情暂不可继续")
		}
	}
	event, err := l.jobs.resume(jobID)
	if err != nil {
		return err
	}
	if event != nil {
		l.jobs.publish(jobID, *event)
	}
	return nil
}

func (l *FetchLoopService) StartTask(req StartTaskRequest) (string, error) {
	switch strings.TrimSpace(req.TaskType) {
	case TaskSpiderDailyBest:
		return l.StartDailyBest()
	case TaskSpiderDailyBestSync:
		return l.StartDailyBestSync()
	case TaskSpiderSeeds:
		return l.StartSeeds()
	case TaskSpiderSeedByName:
		name := strings.TrimSpace(req.Name)
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		return l.StartSeedByName(name)
	case TaskSpiderRefreshOldest:
		if req.Number <= 0 {
			return "", fmt.Errorf("number is required")
		}
		return l.StartRefreshOldestDetail(req.Number)
	case TaskSpiderRebuildCastRank:
		return l.StartRebuildCastRank()
	case TaskSpiderRebuildActorRank:
		actorName := strings.TrimSpace(req.ActorName)
		if actorName == "" {
			return "", fmt.Errorf("actor_name is required")
		}
		return l.StartRebuildActorRank(actorName)
	case TaskSpiderBackfillPerson:
		return l.StartBackfillPerson()
	case TaskSpiderBackfillRankPeriod:
		return l.StartBackfillRankPeriod()
	case TaskFilmRename:
		return l.StartFilmRename()
	case TaskFilmProcess:
		return l.StartFilmProcess()
	case TaskScRebuildStats:
		return l.StartScRebuildStats()
	case TaskScMove:
		scName := strings.TrimSpace(req.ScName)
		if scName == "" {
			return "", fmt.Errorf("sc_name is required")
		}
		return l.StartScMove(scName)
	case TaskScAdd:
		dir := strings.TrimSpace(req.Dir)
		if dir == "" {
			return "", fmt.Errorf("dir is required")
		}
		return l.StartScAdd(sc.AddScInput{
			Dir:            dir,
			ComeMovieJavId: strings.TrimSpace(req.ComeMovieJavID),
			MovieCast:      strings.TrimSpace(req.MovieCast),
			Duration:       req.Duration,
			Fg:             strings.TrimSpace(req.Fg),
			Vessel:         strings.TrimSpace(req.Vessel),
			Remarks:        strings.TrimSpace(req.Remarks),
		})
	default:
		return "", fmt.Errorf("unsupported task_type: %s", req.TaskType)
	}
}

func (l *FetchLoopService) StartDailyBest() (string, error) {
	return l.startManagedTask(TaskSpiderDailyBest, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:           "pipeline_pre",
			Message:         "开始抓取每日榜",
			CurrentPhaseKey: "bestinv",
		})
		return l.crawlLogic.CrawlDailyBestProcession(ctx, false)
	})
}

func (l *FetchLoopService) StartDailyBestSync() (string, error) {
	return l.startManagedTask(TaskSpiderDailyBestSync, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:           "pipeline_pre",
			Message:         "开始同步每日榜",
			CurrentPhaseKey: "bestinv",
		})
		return l.crawlLogic.CrawlDailyBestProcession(ctx, true)
	})
}

func (l *FetchLoopService) StartSeeds() (string, error) {
	return l.startManagedTask(TaskSpiderSeeds, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始抓取活跃种子"})
		return l.crawlLogic.CrawlBySeedsActiveProcession(ctx)
	})
}

func (l *FetchLoopService) StartSeedByName(name string) (string, error) {
	return l.startManagedTask(TaskSpiderSeedByName, taskRuntimePolicy{
		ExclusiveGroup:        taskGroupFetchPriority,
		PauseDetailLoop:       true,
		PreemptsRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: fmt.Sprintf("开始按名称抓取：%s", name)})
		return l.crawlLogic.CrawlBySeedName(ctx, name)
	})
}

func (l *FetchLoopService) StartRefreshOldestDetail(number int64) (string, error) {
	return l.startManagedTask(TaskSpiderRefreshOldest, taskRuntimePolicy{
		ExclusiveGroup:         taskGroupRefreshOldest,
		PauseDetailLoop:        true,
		RegistersRefreshOldest: true,
	}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: fmt.Sprintf("开始刷新最久未更新详情，数量=%d", number)})
		_, err := l.crawlLogic.RefreshOldestDetail(ctx, number)
		return err
	})
}

func (l *FetchLoopService) StartRebuildCastRank() (string, error) {
	return l.startManagedTask(TaskSpiderRebuildCastRank, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始回填演员 rank"})
		return l.crawlLogic.RebuildAllCastRankStats(ctx)
	})
}

func (l *FetchLoopService) StartRebuildActorRank(actorName string) (string, error) {
	return l.startManagedTask(TaskSpiderRebuildActorRank, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: fmt.Sprintf("开始回填演员 rank：%s", actorName)})
		return l.crawlLogic.RebuildCastRankStatsByName(ctx, actorName)
	})
}

func (l *FetchLoopService) StartBackfillPerson() (string, error) {
	return l.startManagedTask(TaskSpiderBackfillPerson, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始回填 person"})
		return l.crawlLogic.BackfillPersonData(ctx)
	})
}

func (l *FetchLoopService) StartBackfillRankPeriod() (string, error) {
	return l.startManagedTask(TaskSpiderBackfillRankPeriod, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "pipeline_pre", Message: "开始回填周期排行"})
		return l.movieSvc.RebuildAllRankPeriods(ctx)
	})
}

func (l *FetchLoopService) StartFilmRename() (string, error) {
	return l.startManagedTask(TaskFilmRename, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "film_prepare", Message: "开始重命名影片"})
		return l.filmSvc.RenameFilm(ctx)
	})
}

func (l *FetchLoopService) StartFilmProcess() (string, error) {
	return l.startManagedTask(TaskFilmProcess, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "film_prepare", Message: "开始处理影片"})
		return l.filmSvc.ProcessFilm(ctx)
	})
}

func (l *FetchLoopService) StartScRebuildStats() (string, error) {
	return l.startManagedTask(TaskScRebuildStats, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "sc_prepare", Message: "开始回填 SC 统计"})
		return l.scSvc.RebuildAllScStats(ctx)
	})
}

func (l *FetchLoopService) StartScMove(scName string) (string, error) {
	return l.startManagedTask(TaskScMove, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "sc_prepare", Message: fmt.Sprintf("开始移动 SC 影片：%s", scName)})
		return l.scSvc.MoveScFilm(ctx, scName)
	})
}

func (l *FetchLoopService) StartScAdd(in sc.AddScInput) (string, error) {
	return l.startManagedTask(TaskScAdd, taskRuntimePolicy{}, func(ctx context.Context) error {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "sc_prepare", Message: fmt.Sprintf("开始新增 SC：%s", in.Dir)})
		return l.scSvc.AddSc(ctx, in)
	})
}

func (l *FetchLoopService) startManagedTask(taskType string, policy taskRuntimePolicy, run func(context.Context) error) (string, error) {
	if run == nil {
		return "", fmt.Errorf("nil task runner")
	}
	jobID := l.jobs.create(taskType)
	if err := l.beginTaskRuntime(jobID, taskType, policy); err != nil {
		l.jobs.finish(jobID, JobEvent{
			Kind:     JobEventKindProgress,
			JobID:    jobID,
			TaskType: taskType,
			Stage:    "failed",
			Message:  err.Error(),
			Done:     true,
			At:       time.Now().Unix(),
		})
		return "", err
	}

	ctx, cancel := context.WithCancel(l.currentRootContext())
	if err := l.jobs.setCancel(jobID, cancel); err != nil {
		cancel()
		l.finishTaskRuntime(jobID, policy)
		return "", err
	}

	go func() {
		defer cancel()
		defer l.jobs.clearCancel(jobID)
		defer l.finishTaskRuntime(jobID, policy)

		ctx = taskctx.WithPauseWaiter(ctx, func(ctx context.Context) error {
			return l.jobs.waitIfPaused(ctx, jobID)
		})
		ctx = taskctx.WithProgressReporter(ctx, func(progress taskctx.Progress) {
			l.publishTaskProgress(jobID, taskType, progress)
		})
		ctx = taskctx.WithLogReporter(ctx, func(logEvent taskctx.Log) {
			l.publishTaskLog(jobID, taskType, logEvent)
		})

		l.publishTaskProgress(jobID, taskType, taskctx.Progress{
			Stage:   "job_started",
			Message: "任务已启动",
		})

		l.applyTaskStartPolicy(jobID, policy)

		if err := l.jobs.waitIfPaused(ctx, jobID); err != nil {
			l.jobs.finish(jobID, JobEvent{
				Kind:     JobEventKindProgress,
				JobID:    jobID,
				TaskType: taskType,
				Stage:    "failed",
				Message:  err.Error(),
				Done:     true,
				At:       time.Now().Unix(),
			})
			return
		}

		detailPauseApplied := false
		detailPauseStateChanged := false
		if policy.PauseDetailLoop && l.IsDetailLoopRunning() {
			detailPauseApplied, detailPauseStateChanged = l.pauseDetailLoop(jobID)
		}
		if detailPauseStateChanged {
			l.publishTaskProgress(jobID, taskType, taskctx.Progress{
				Stage:   "detail_paused",
				Message: "详情抓取已暂停",
			})
		}

		err := run(ctx)

		if detailPauseApplied {
			_, detailResumed := l.resumeDetailLoop(jobID)
			if detailResumed {
				l.publishTaskProgress(jobID, taskType, taskctx.Progress{
					Stage:   "detail_resumed",
					Message: "详情抓取已恢复",
				})
			}
		}
		if policy.PreemptsRefreshOldest {
			l.resumeRefreshOldestIfAutoPaused()
		}

		if err != nil {
			l.jobs.finish(jobID, JobEvent{
				Kind:     JobEventKindProgress,
				JobID:    jobID,
				TaskType: taskType,
				Stage:    "failed",
				Message:  err.Error(),
				Done:     true,
				At:       time.Now().Unix(),
			})
			return
		}

		l.jobs.finish(jobID, JobEvent{
			Kind:     JobEventKindProgress,
			JobID:    jobID,
			TaskType: taskType,
			Stage:    "done",
			Message:  "任务完成",
			Done:     true,
			At:       time.Now().Unix(),
		})
	}()

	return jobID, nil
}
