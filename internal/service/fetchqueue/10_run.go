package fetchqueue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
	"rudy_gc/internal/types"
)

func (s *Service) BackfillFetchTasks(ctx context.Context, pageSize int64) (*fetchsite.BackfillResult, error) {
	return s.siteSvc.BackfillFetchTasks(ctx, pageSize)
}

func (s *Service) RunPendingJavbusFetchTasks(ctx context.Context, limit int64) (*fetchsite.RunFetchTasksResult, error) {
	tasks, err := s.siteSvc.ListPendingJavbusFetchTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	return s.runJavbusTasks(ctx, tasks, "JavBus 待抓取任务已加载")
}

func (s *Service) RunPendingSukebeiFetchTasks(ctx context.Context, limit int64) (*fetchsite.RunFetchTasksResult, error) {
	tasks, err := s.siteSvc.ListPendingSukebeiFetchTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	return s.runSukebeiTasks(ctx, tasks, "Sukebei 待抓取任务已加载")
}

func (s *Service) RunJavbusFetchTasksByMovies(
	ctx context.Context,
	movies []*types.MovieType,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) (*fetchsite.RunFetchTasksResult, error) {
	tasks, skippedFetch, skippedSuccess, err := s.BuildFilteredJavbusFetchTasksByMovies(ctx, movies, lastFetchDurationDays, lastSuccessDurationDays)
	if err != nil {
		return nil, err
	}
	reportRecentDurationSkipLog(ctx, "JavBus", skippedFetch, skippedSuccess, lastFetchDurationDays, lastSuccessDurationDays)
	return s.RunPreparedJavbusFetchTasks(ctx, tasks, "JavBus 筛选任务已加载")
}

func (s *Service) RunSukebeiFetchTasksByMovies(
	ctx context.Context,
	movies []*types.MovieType,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) (*fetchsite.RunFetchTasksResult, error) {
	tasks, skippedFetch, skippedSuccess, err := s.BuildFilteredSukebeiFetchTasksByMovies(ctx, movies, lastFetchDurationDays, lastSuccessDurationDays)
	if err != nil {
		return nil, err
	}
	reportRecentDurationSkipLog(ctx, "Sukebei", skippedFetch, skippedSuccess, lastFetchDurationDays, lastSuccessDurationDays)
	return s.RunPreparedSukebeiFetchTasks(ctx, tasks, "Sukebei 筛选任务已加载")
}

func (s *Service) BuildFilteredJavbusFetchTasksByMovies(
	ctx context.Context,
	movies []*types.MovieType,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) ([]*fetchsite.JavbusFetchTask, int, int, error) {
	javbusTasks, err := s.siteSvc.BuildJavbusFetchTasksByMovies(ctx, movies)
	if err != nil {
		return nil, 0, 0, err
	}
	javbusTasks, skippedFetch, skippedSuccess := filterJavbusTasksByRecentDuration(javbusTasks, lastFetchDurationDays, lastSuccessDurationDays)
	return javbusTasks, skippedFetch, skippedSuccess, nil
}

func (s *Service) BuildFilteredSukebeiFetchTasksByMovies(
	ctx context.Context,
	movies []*types.MovieType,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) ([]*fetchsite.SukebeiFetchTask, int, int, error) {
	tasks, err := s.siteSvc.BuildSukebeiFetchTasksByMovies(ctx, movies)
	if err != nil {
		return nil, 0, 0, err
	}
	tasks, skippedFetch, skippedSuccess := filterSukebeiTasksByRecentDuration(tasks, lastFetchDurationDays, lastSuccessDurationDays)
	return tasks, skippedFetch, skippedSuccess, nil
}

func (s *Service) RunPreparedJavbusFetchTasks(
	ctx context.Context,
	tasks []*fetchsite.JavbusFetchTask,
	loadLabel string,
) (*fetchsite.RunFetchTasksResult, error) {
	return s.runJavbusTasks(ctx, tasks, loadLabel)
}

func (s *Service) RunPreparedSukebeiFetchTasks(
	ctx context.Context,
	tasks []*fetchsite.SukebeiFetchTask,
	loadLabel string,
) (*fetchsite.RunFetchTasksResult, error) {
	return s.runSukebeiTasks(ctx, tasks, loadLabel)
}

func (s *Service) RunSingleJavbusFetchTask(ctx context.Context, movieJavID, movieCode string) (*fetchsite.RunFetchTasksResult, error) {
	task, err := s.siteSvc.FindJavbusFetchTask(ctx, movieJavID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &fetchsite.RunFetchTasksResult{}, nil
	}
	if movieCode != "" {
		task.MovieCode = movieCode
	}
	result := &fetchsite.RunFetchTasksResult{Queued: 1, Handled: 1}
	if err := s.runJavbusTask(ctx, task); err != nil {
		result.Failed = 1
		return result, err
	}
	result.Success = 1
	return result, nil
}

func (s *Service) RunSingleSukebeiFetchTask(ctx context.Context, movieJavID, movieCode string) (*fetchsite.RunFetchTasksResult, error) {
	task, err := s.siteSvc.FindSukebeiFetchTask(ctx, movieJavID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &fetchsite.RunFetchTasksResult{}, nil
	}
	if movieCode != "" {
		task.MovieCode = movieCode
	}
	result := &fetchsite.RunFetchTasksResult{Queued: 1, Handled: 1}
	if err := s.runSukebeiTask(ctx, task); err != nil {
		result.Failed = 1
		return result, err
	}
	result.Success = 1
	return result, nil
}

func (s *Service) runJavbusTasks(ctx context.Context, tasks []*fetchsite.JavbusFetchTask, loadLabel string) (*fetchsite.RunFetchTasksResult, error) {
	result := &fetchsite.RunFetchTasksResult{Queued: len(tasks)}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_javbus_queue_ready",
		Message:           fmt.Sprintf("%s：%d 条", loadLabel, len(tasks)),
		QueuedCount:       len(tasks),
		CurrentPhaseKey:   "fetch_javbus",
		PhaseKey:          "fetch_javbus",
		PhaseHandledCount: 0,
		PhaseTotalCount:   len(tasks),
		PhaseSuccessCount: 0,
		PhaseFailedCount:  0,
	})

	hasPreviousRemoteRequest := false
	for _, task := range tasks {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
		if task == nil || task.MovieJavID == "" || task.MovieCode == "" {
			continue
		}

		if hasPreviousRemoteRequest {
			if err := s.siteSvc.SleepRequest(ctx, fetchsite.FetchSiteCodeJavbus); err != nil {
				return nil, err
			}
		}

		result.Handled++
		runErr := s.runJavbusTask(ctx, task)
		if runErr != nil {
			reportErrorLog(ctx, fmt.Sprintf("JavBus 抓取失败: %s | err=%v", task.MovieCode, runErr))
			result.Failed++
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:             "fetch_javbus_failed",
				Message:           fmt.Sprintf("JavBus 已完成 %d/%d", result.Handled, len(tasks)),
				HandledCount:      result.Handled,
				SuccessCount:      result.Success,
				FailedCount:       result.Failed,
				QueuedCount:       len(tasks) - result.Handled,
				CurrentPhaseKey:   "fetch_javbus",
				PhaseKey:          "fetch_javbus",
				PhaseHandledCount: result.Handled,
				PhaseTotalCount:   len(tasks),
				PhaseSuccessCount: result.Success,
				PhaseFailedCount:  result.Failed,
			})
		} else {
			reportInfoLog(ctx, fmt.Sprintf("JavBus 抓取成功: %s", task.MovieCode))
			result.Success++
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:             "fetch_javbus_done",
				Message:           fmt.Sprintf("JavBus 已完成 %d/%d", result.Handled, len(tasks)),
				HandledCount:      result.Handled,
				SuccessCount:      result.Success,
				FailedCount:       result.Failed,
				QueuedCount:       len(tasks) - result.Handled,
				CurrentPhaseKey:   "fetch_javbus",
				PhaseKey:          "fetch_javbus",
				PhaseHandledCount: result.Handled,
				PhaseTotalCount:   len(tasks),
				PhaseSuccessCount: result.Success,
				PhaseFailedCount:  result.Failed,
			})
		}
		hasPreviousRemoteRequest = true
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_javbus_all_done",
		Message:           "JavBus 抓取完成",
		HandledCount:      result.Handled,
		SuccessCount:      result.Success,
		FailedCount:       result.Failed,
		QueuedCount:       0,
		CurrentPhaseKey:   "fetch_javbus",
		PhaseKey:          "fetch_javbus",
		PhaseHandledCount: result.Handled,
		PhaseTotalCount:   len(tasks),
		PhaseSuccessCount: result.Success,
		PhaseFailedCount:  result.Failed,
	})
	return result, nil
}

func (s *Service) runSukebeiTasks(ctx context.Context, tasks []*fetchsite.SukebeiFetchTask, loadLabel string) (*fetchsite.RunFetchTasksResult, error) {
	result := &fetchsite.RunFetchTasksResult{Queued: len(tasks)}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_sukebei_queue_ready",
		Message:           fmt.Sprintf("%s：%d 条", loadLabel, len(tasks)),
		QueuedCount:       len(tasks),
		CurrentPhaseKey:   "fetch_sukebei",
		PhaseKey:          "fetch_sukebei",
		PhaseHandledCount: 0,
		PhaseTotalCount:   len(tasks),
		PhaseSuccessCount: 0,
		PhaseFailedCount:  0,
	})

	hasPreviousRemoteRequest := false
	for _, task := range tasks {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
		if task == nil || task.MovieJavID == "" || task.MovieCode == "" {
			continue
		}

		if hasPreviousRemoteRequest {
			if err := s.siteSvc.SleepRequest(ctx, fetchsite.FetchSiteCodeSukebei); err != nil {
				return nil, err
			}
		}

		result.Handled++
		runErr := s.runSukebeiTask(ctx, task)
		if runErr != nil {
			reportErrorLog(ctx, fmt.Sprintf("Sukebei 抓取失败: %s | err=%v", task.MovieCode, runErr))
			result.Failed++
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:             "fetch_sukebei_failed",
				Message:           fmt.Sprintf("Sukebei 已完成 %d/%d", result.Handled, len(tasks)),
				HandledCount:      result.Handled,
				SuccessCount:      result.Success,
				FailedCount:       result.Failed,
				QueuedCount:       len(tasks) - result.Handled,
				CurrentPhaseKey:   "fetch_sukebei",
				PhaseKey:          "fetch_sukebei",
				PhaseHandledCount: result.Handled,
				PhaseTotalCount:   len(tasks),
				PhaseSuccessCount: result.Success,
				PhaseFailedCount:  result.Failed,
			})
		} else {
			reportInfoLog(ctx, fmt.Sprintf("Sukebei 抓取成功: %s", task.MovieCode))
			result.Success++
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:             "fetch_sukebei_done",
				Message:           fmt.Sprintf("Sukebei 已完成 %d/%d", result.Handled, len(tasks)),
				HandledCount:      result.Handled,
				SuccessCount:      result.Success,
				FailedCount:       result.Failed,
				QueuedCount:       len(tasks) - result.Handled,
				CurrentPhaseKey:   "fetch_sukebei",
				PhaseKey:          "fetch_sukebei",
				PhaseHandledCount: result.Handled,
				PhaseTotalCount:   len(tasks),
				PhaseSuccessCount: result.Success,
				PhaseFailedCount:  result.Failed,
			})
		}
		hasPreviousRemoteRequest = true
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_sukebei_all_done",
		Message:           "Sukebei 抓取完成",
		HandledCount:      result.Handled,
		SuccessCount:      result.Success,
		FailedCount:       result.Failed,
		QueuedCount:       0,
		CurrentPhaseKey:   "fetch_sukebei",
		PhaseKey:          "fetch_sukebei",
		PhaseHandledCount: result.Handled,
		PhaseTotalCount:   len(tasks),
		PhaseSuccessCount: result.Success,
		PhaseFailedCount:  result.Failed,
	})
	return result, nil
}

func (s *Service) runJavbusTask(ctx context.Context, task *fetchsite.JavbusFetchTask) error {
	if task == nil || task.Row == nil {
		return fmt.Errorf("invalid javbus fetch task")
	}
	if err := s.siteSvc.MarkJavbusRunning(ctx, task.Row); err != nil {
		return err
	}
	_, err := s.javbusSvc.FetchMovieMagnets(ctx, task.MovieJavID, task.MovieCode)
	return err
}

func (s *Service) runSukebeiTask(ctx context.Context, task *fetchsite.SukebeiFetchTask) error {
	if task == nil || task.Row == nil {
		return fmt.Errorf("invalid sukebei fetch task")
	}
	if err := s.siteSvc.MarkSukebeiRunning(ctx, task.Row); err != nil {
		return err
	}
	_, err := s.sukebeiSvc.FetchMovieTorrents(ctx, task.MovieJavID, task.MovieCode)
	return err
}

func filterJavbusTasksByRecentDuration(
	tasks []*fetchsite.JavbusFetchTask,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) ([]*fetchsite.JavbusFetchTask, int, int) {
	now := time.Now().Unix()
	out := make([]*fetchsite.JavbusFetchTask, 0, len(tasks))
	skippedByFetch := 0
	skippedBySuccess := 0
	for _, task := range tasks {
		if task == nil {
			continue
		}
		lastFetchTime := int64(0)
		lastSuccessTime := int64(0)
		if task.Row != nil {
			lastFetchTime = task.Row.LastFetchTime
			lastSuccessTime = task.Row.LastSuccessTime
		}
		if shouldSkipByRecentDuration(now, lastFetchTime, lastFetchDurationDays) {
			skippedByFetch++
			continue
		}
		if shouldSkipByRecentDuration(now, lastSuccessTime, lastSuccessDurationDays) {
			skippedBySuccess++
			continue
		}
		out = append(out, task)
	}
	return out, skippedByFetch, skippedBySuccess
}

func filterSukebeiTasksByRecentDuration(
	tasks []*fetchsite.SukebeiFetchTask,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) ([]*fetchsite.SukebeiFetchTask, int, int) {
	now := time.Now().Unix()
	out := make([]*fetchsite.SukebeiFetchTask, 0, len(tasks))
	skippedByFetch := 0
	skippedBySuccess := 0
	for _, task := range tasks {
		if task == nil {
			continue
		}
		lastFetchTime := int64(0)
		lastSuccessTime := int64(0)
		if task.Row != nil {
			lastFetchTime = task.Row.LastFetchTime
			lastSuccessTime = task.Row.LastSuccessTime
		}
		if shouldSkipByRecentDuration(now, lastFetchTime, lastFetchDurationDays) {
			skippedByFetch++
			continue
		}
		if shouldSkipByRecentDuration(now, lastSuccessTime, lastSuccessDurationDays) {
			skippedBySuccess++
			continue
		}
		out = append(out, task)
	}
	return out, skippedByFetch, skippedBySuccess
}

func shouldSkipByRecentDuration(now, targetTime, durationDays int64) bool {
	if durationDays <= 0 || targetTime <= 0 {
		return false
	}
	return now-targetTime < durationDays*24*3600
}

func reportRecentDurationSkipLog(
	ctx context.Context,
	siteName string,
	skippedByFetch int,
	skippedBySuccess int,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) {
	parts := make([]string, 0, 2)
	if lastFetchDurationDays > 0 {
		parts = append(parts, fmt.Sprintf("最近抓取<%d天：%d条", lastFetchDurationDays, skippedByFetch))
	}
	if lastSuccessDurationDays > 0 {
		parts = append(parts, fmt.Sprintf("最近成功<%d天：%d条", lastSuccessDurationDays, skippedBySuccess))
	}
	if len(parts) == 0 {
		return
	}
	reportInfoLog(ctx, fmt.Sprintf("%s 过滤跳过：%s", siteName, strings.Join(parts, "，")))
}
