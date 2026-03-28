package fetchqueue

import (
	"context"
	"fmt"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

func (s *Service) BackfillFetchTasks(ctx context.Context, pageSize int64) (*fetchsite.BackfillResult, error) {
	return s.siteSvc.BackfillFetchTasks(ctx, pageSize)
}

func (s *Service) RunPendingJavbusFetchTasks(ctx context.Context, limit int64) (*fetchsite.RunFetchTasksResult, error) {
	tasks, err := s.siteSvc.ListPendingJavbusFetchTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	reportInfoLog(ctx, fmt.Sprintf("JavBus 待抓取任务已加载: %d 条", len(tasks)))

	result := &fetchsite.RunFetchTasksResult{Queued: len(tasks)}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_javbus_queue_ready",
		Message:           fmt.Sprintf("待抓取 JavBus %d 条", len(tasks)),
		QueuedCount:       len(tasks),
		CurrentPhaseKey:   "fetch_javbus",
		PhaseKey:          "fetch_javbus",
		PhaseHandledCount: 0,
		PhaseTotalCount:   len(tasks),
		PhaseSuccessCount: 0,
		PhaseFailedCount:  0,
	})

	hasPreviousRemoteRequest := false
	for i, task := range tasks {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
		if task == nil || task.MovieJavID == "" || task.MovieCode == "" {
			continue
		}

		if hasPreviousRemoteRequest {
			reportInfoLog(ctx, fmt.Sprintf("JavBus 请求前等待: %s", task.MovieCode))
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
				Message:           fmt.Sprintf("JavBus 抓取失败 %s：%v", task.MovieCode, runErr),
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
				Message:           fmt.Sprintf("JavBus 抓取完成 %s (%d/%d)", task.MovieCode, i+1, len(tasks)),
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
	reportInfoLog(ctx, fmt.Sprintf("JavBus 抓取批次结束: success=%d | failed=%d", result.Success, result.Failed))

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_javbus_all_done",
		Message:           fmt.Sprintf("JavBus 抓取完成：成功=%d，失败=%d", result.Success, result.Failed),
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

func (s *Service) RunPendingSukebeiFetchTasks(ctx context.Context, limit int64) (*fetchsite.RunFetchTasksResult, error) {
	tasks, err := s.siteSvc.ListPendingSukebeiFetchTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	reportInfoLog(ctx, fmt.Sprintf("Sukebei 待抓取任务已加载: %d 条", len(tasks)))

	result := &fetchsite.RunFetchTasksResult{Queued: len(tasks)}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_sukebei_queue_ready",
		Message:           fmt.Sprintf("待抓取 Sukebei %d 条", len(tasks)),
		QueuedCount:       len(tasks),
		CurrentPhaseKey:   "fetch_sukebei",
		PhaseKey:          "fetch_sukebei",
		PhaseHandledCount: 0,
		PhaseTotalCount:   len(tasks),
		PhaseSuccessCount: 0,
		PhaseFailedCount:  0,
	})

	hasPreviousRemoteRequest := false
	for i, task := range tasks {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
		if task == nil || task.MovieJavID == "" || task.MovieCode == "" {
			continue
		}

		if hasPreviousRemoteRequest {
			reportInfoLog(ctx, fmt.Sprintf("Sukebei 请求前等待: %s", task.MovieCode))
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
				Message:           fmt.Sprintf("Sukebei 抓取失败 %s：%v", task.MovieCode, runErr),
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
				Message:           fmt.Sprintf("Sukebei 抓取完成 %s (%d/%d)", task.MovieCode, i+1, len(tasks)),
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
	reportInfoLog(ctx, fmt.Sprintf("Sukebei 抓取批次结束: success=%d | failed=%d", result.Success, result.Failed))

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_sukebei_all_done",
		Message:           fmt.Sprintf("Sukebei 抓取完成：成功=%d，失败=%d", result.Success, result.Failed),
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

func (s *Service) runJavbusTask(ctx context.Context, task *fetchsite.JavbusFetchTask) error {
	if task == nil || task.Row == nil {
		return fmt.Errorf("invalid javbus fetch task")
	}
	reportInfoLog(ctx, fmt.Sprintf("开始抓取 JavBus: %s", task.MovieCode))
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
	reportInfoLog(ctx, fmt.Sprintf("开始抓取 Sukebei: %s", task.MovieCode))
	if err := s.siteSvc.MarkSukebeiRunning(ctx, task.Row); err != nil {
		return err
	}
	_, err := s.sukebeiSvc.FetchMovieTorrents(ctx, task.MovieJavID, task.MovieCode)
	return err
}
