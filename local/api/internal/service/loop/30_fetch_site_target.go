package loop

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"rudy_gc/local/api/internal/service/fetchqueue"
	"rudy_gc/local/api/internal/service/fetchsite"
	"rudy_gc/local/api/internal/taskctx"
)

type fetchSiteRunState struct {
	SelectedCount     int64
	CurrentSource     string
	CurrentMovieJavID string
	CurrentMovieName  string
	CompletedCount    int64
	SuccessCount      int64
	FailedCount       int64
	JavbusCompleted   int64
	JavbusSuccess     int64
	JavbusFailed      int64
	SukebeiCompleted  int64
	SukebeiSuccess    int64
	SukebeiFailed     int64
}

func (l *FetchLoopService) startFetchSiteTargets(req StartTaskRequest) (string, error) {
	taskType := strings.TrimSpace(req.TaskType)
	targets, err := normalizeFetchSiteTargets(taskType, req.FetchSiteTargets)
	if err != nil {
		return "", err
	}
	return l.startManagedTask(taskType, taskRuntimePolicy{}, func(ctx context.Context) error {
		return l.runFetchSiteTargets(ctx, targets)
	})
}

func normalizeFetchSiteTargets(taskType string, targets []FetchSiteTarget) ([]FetchSiteTarget, error) {
	if !isFetchSitePreviewTaskType(taskType) {
		return nil, fmt.Errorf("unsupported task_type: %s", taskType)
	}
	normalized := make([]FetchSiteTarget, 0, len(targets))
	seen := make(map[string]struct{}, len(targets))
	for index, target := range targets {
		target.Source = strings.ToLower(strings.TrimSpace(target.Source))
		target.MovieJavID = strings.TrimSpace(target.MovieJavID)
		target.MovieName = strings.TrimSpace(target.MovieName)
		if target.Sequence <= 0 {
			target.Sequence = int64(index + 1)
		}
		if !fetchSiteTargetSourceAllowed(taskType, target.Source) {
			return nil, fmt.Errorf("task_type %s does not support source %s", taskType, target.Source)
		}
		if target.MovieJavID == "" || target.MovieName == "" {
			return nil, fmt.Errorf("fetch site target movie_jav_id and movie_name are required")
		}
		key := target.Source + "\x00" + target.MovieJavID
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate fetch site target: %s/%s", target.Source, target.MovieJavID)
		}
		seen[key] = struct{}{}
		normalized = append(normalized, target)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("当前预览没有可抓取目标")
	}
	return normalized, nil
}

func fetchSiteTargetSourceAllowed(taskType string, source string) bool {
	switch taskType {
	case TaskSpiderFetchJavbus:
		return source == fetchSiteSourceJavbus
	case TaskSpiderFetchSukebei:
		return source == fetchSiteSourceSukebei
	case TaskSpiderFetchSiteBoth:
		return source == fetchSiteSourceJavbus || source == fetchSiteSourceSukebei
	default:
		return false
	}
}

func (l *FetchLoopService) runFetchSiteTargets(ctx context.Context, targets []FetchSiteTarget) error {
	javbusTasks, sukebeiTasks, err := l.prepareFetchSiteTasks(ctx, targets)
	if err != nil {
		return err
	}
	run := &fetchSiteRunState{SelectedCount: int64(len(targets))}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:           "fetch_site_targets_start",
		Message:         fmt.Sprintf("开始抓取预览清单，抓取项=%d，JavBus=%d，Sukebei=%d", len(targets), len(javbusTasks), len(sukebeiTasks)),
		QueuedCount:     len(targets),
		CurrentPhaseKey: "fetch_site",
		Stats:           fetchSiteRunStats(run),
		StatsText:       fetchSiteRunStatsText(run),
	})

	var runMu sync.Mutex
	hooks := &fetchqueue.FetchTaskHooks{
		Snapshot: func() (map[string]int64, map[string]string) {
			runMu.Lock()
			defer runMu.Unlock()
			return fetchSiteRunStats(run), fetchSiteRunStatsText(run)
		},
		Before: func(event fetchqueue.FetchTaskHookEvent) error {
			runMu.Lock()
			defer runMu.Unlock()
			run.CurrentSource = event.Source
			run.CurrentMovieJavID = event.MovieJavID
			run.CurrentMovieName = event.MovieName
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:           "fetch_site_item_start",
				Message:         fmt.Sprintf("开始抓取 %s：%s", fetchSiteSourceLabel(event.Source), event.MovieName),
				HandledCount:    int(run.CompletedCount),
				QueuedCount:     int(run.SelectedCount - run.CompletedCount),
				CurrentPhaseKey: fetchSiteSourcePhaseKey(event.Source),
				Stats:           fetchSiteRunStats(run),
				StatsText:       fetchSiteRunStatsText(run),
			})
			return nil
		},
		After: func(event fetchqueue.FetchTaskHookEvent, taskErr error) error {
			runMu.Lock()
			defer runMu.Unlock()
			run.CompletedCount++
			if taskErr != nil {
				run.FailedCount++
			} else {
				run.SuccessCount++
			}
			if event.Source == fetchSiteSourceJavbus {
				run.JavbusCompleted++
				if taskErr != nil {
					run.JavbusFailed++
				} else {
					run.JavbusSuccess++
				}
			} else {
				run.SukebeiCompleted++
				if taskErr != nil {
					run.SukebeiFailed++
				} else {
					run.SukebeiSuccess++
				}
			}
			message := fmt.Sprintf("%s 抓取完成：%s", fetchSiteSourceLabel(event.Source), event.MovieName)
			stage := "fetch_site_item_done"
			if taskErr != nil {
				message = fmt.Sprintf("%s 抓取失败：%s，错误=%v", fetchSiteSourceLabel(event.Source), event.MovieName, taskErr)
				stage = "fetch_site_item_failed"
			}
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:           stage,
				Message:         message,
				HandledCount:    int(run.CompletedCount),
				SuccessCount:    int(run.SuccessCount),
				FailedCount:     int(run.FailedCount),
				QueuedCount:     int(run.SelectedCount - run.CompletedCount),
				CurrentPhaseKey: fetchSiteSourcePhaseKey(event.Source),
				Stats:           fetchSiteRunStats(run),
				StatsText:       fetchSiteRunStatsText(run),
			})
			return nil
		},
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultCh := make(chan error, 2)
	workerCount := 0
	if len(javbusTasks) > 0 {
		workerCount++
		go func() {
			_, runErr := l.fetchQueue.RunPreparedJavbusFetchTasksWithHooks(runCtx, javbusTasks, "JavBus 预览任务已加载", hooks)
			resultCh <- runErr
		}()
	}
	if len(sukebeiTasks) > 0 {
		workerCount++
		go func() {
			_, runErr := l.fetchQueue.RunPreparedSukebeiFetchTasksWithHooks(runCtx, sukebeiTasks, "Sukebei 预览任务已加载", hooks)
			resultCh <- runErr
		}()
	}
	for range workerCount {
		if runErr := <-resultCh; runErr != nil && err == nil {
			err = runErr
			cancel()
		}
	}
	if err != nil {
		return err
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "fetch_site_targets_done",
		Message:      fmt.Sprintf("抓取预览清单完成：处理=%d，成功=%d，失败=%d", run.CompletedCount, run.SuccessCount, run.FailedCount),
		HandledCount: int(run.CompletedCount),
		SuccessCount: int(run.SuccessCount),
		FailedCount:  int(run.FailedCount),
		QueuedCount:  0,
		Stats:        fetchSiteRunStats(run),
		StatsText:    fetchSiteRunStatsText(run),
	})
	return nil
}

func (l *FetchLoopService) prepareFetchSiteTasks(ctx context.Context, targets []FetchSiteTarget) ([]*fetchsite.JavbusFetchTask, []*fetchsite.SukebeiFetchTask, error) {
	javbusTasks := make([]*fetchsite.JavbusFetchTask, 0)
	sukebeiTasks := make([]*fetchsite.SukebeiFetchTask, 0)
	for _, target := range targets {
		switch target.Source {
		case fetchSiteSourceJavbus:
			task, err := l.fetchSiteSvc.FindJavbusFetchTask(ctx, target.MovieJavID)
			if err != nil {
				return nil, nil, err
			}
			if task == nil {
				return nil, nil, fmt.Errorf("JavBus 抓取目标不存在: %s", target.MovieJavID)
			}
			task.MovieName = target.MovieName
			javbusTasks = append(javbusTasks, task)
		case fetchSiteSourceSukebei:
			task, err := l.fetchSiteSvc.FindSukebeiFetchTask(ctx, target.MovieJavID)
			if err != nil {
				return nil, nil, err
			}
			if task == nil {
				return nil, nil, fmt.Errorf("Sukebei 抓取目标不存在: %s", target.MovieJavID)
			}
			task.MovieName = target.MovieName
			sukebeiTasks = append(sukebeiTasks, task)
		default:
			return nil, nil, fmt.Errorf("unsupported fetch site source: %s", target.Source)
		}
	}
	return javbusTasks, sukebeiTasks, nil
}

func fetchSiteRunStats(run *fetchSiteRunState) map[string]int64 {
	return map[string]int64{
		"selected_items":    run.SelectedCount,
		"completed_items":   run.CompletedCount,
		"success_items":     run.SuccessCount,
		"failed_items":      run.FailedCount,
		"javbus_completed":  run.JavbusCompleted,
		"javbus_success":    run.JavbusSuccess,
		"javbus_failed":     run.JavbusFailed,
		"sukebei_completed": run.SukebeiCompleted,
		"sukebei_success":   run.SukebeiSuccess,
		"sukebei_failed":    run.SukebeiFailed,
	}
}

func fetchSiteRunStatsText(run *fetchSiteRunState) map[string]string {
	return map[string]string{
		"current_source":       fetchSiteSourceLabel(run.CurrentSource),
		"current_movie_jav_id": run.CurrentMovieJavID,
		"current_movie_name":   run.CurrentMovieName,
	}
}

func fetchSiteSourceLabel(source string) string {
	switch strings.TrimSpace(source) {
	case fetchSiteSourceJavbus:
		return "JavBus"
	case fetchSiteSourceSukebei:
		return "Sukebei"
	default:
		return strings.TrimSpace(source)
	}
}

func fetchSiteSourcePhaseKey(source string) string {
	switch strings.TrimSpace(source) {
	case fetchSiteSourceJavbus:
		return "fetch_javbus"
	case fetchSiteSourceSukebei:
		return "fetch_sukebei"
	default:
		return strings.TrimSpace(source)
	}
}
