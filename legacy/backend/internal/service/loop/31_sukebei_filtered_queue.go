package loop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

func (l *FetchLoopService) runSukebeiFilteredQueue(ctx context.Context, tasks []*fetchsite.SukebeiFetchTask) error {
	validTasks := make([]*fetchsite.SukebeiFetchTask, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		task.MovieJavID = strings.TrimSpace(task.MovieJavID)
		task.MovieName = strings.TrimSpace(task.MovieName)
		if task.MovieJavID == "" || task.MovieName == "" {
			continue
		}
		validTasks = append(validTasks, task)
	}

	pending := buildSukebeiFilteredPendingItems(validTasks)
	running := []taskctx.QueueItem{}
	done := []taskctx.QueueItem{}
	successCount := 0
	failedCount := 0
	total := len(validTasks)

	reportSukebeiFilteredQueueLog(ctx, "info", fmt.Sprintf("筛选抓取队列已加载：%d 条", total))
	reportSukebeiFilteredQueueProgress(ctx, "fetch_sukebei_filtered_queue_ready", fmt.Sprintf("筛选抓取队列已加载：%d 条", total), pending, running, done, successCount, failedCount)

	for index, task := range validTasks {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return err
		}
		if index > 0 {
			if err := l.fetchSiteSvc.SleepRequest(ctx, fetchsite.FetchSiteCodeSukebei); err != nil {
				return err
			}
		}

		current := taskctx.QueueItem{
			MovieJavID: task.MovieJavID,
			MovieName:  task.MovieName,
			Seq:        int64(index + 1),
			State:      "running",
		}
		if len(pending) > 0 {
			pending = pending[1:]
		}
		running = []taskctx.QueueItem{current}
		reportSukebeiFilteredQueueLog(ctx, "info", fmt.Sprintf("开始抓取 %d/%d：%s", index+1, total, task.MovieName))
		reportSukebeiFilteredQueueProgress(ctx, "fetch_sukebei_filtered_running", fmt.Sprintf("开始抓取 %d/%d：%s", index+1, total, task.MovieName), pending, running, done, successCount, failedCount)

		_, err := l.fetchQueue.RunSingleSukebeiFetchTask(ctx, task.MovieJavID, task.MovieName)
		running = []taskctx.QueueItem{}

		doneItem := taskctx.QueueItem{
			MovieJavID: task.MovieJavID,
			MovieName:  task.MovieName,
			Seq:        int64(index + 1),
		}
		if err != nil {
			failedCount++
			doneItem.State = "failed"
			doneItem.Error = err.Error()
			doneItem.CompletedAt = time.Now().Unix()
			reportSukebeiFilteredQueueLog(ctx, "error", fmt.Sprintf("抓取失败 %d/%d：%s | err=%v", index+1, total, task.MovieName, err))
			done = append(done, doneItem)
			reportSukebeiFilteredQueueProgress(ctx, "fetch_sukebei_filtered_failed", fmt.Sprintf("抓取失败 %d/%d：%s", index+1, total, task.MovieName), pending, running, done, successCount, failedCount)
			continue
		}

		successCount++
		doneItem.State = "success"
		doneItem.CompletedAt = time.Now().Unix()
		reportSukebeiFilteredQueueLog(ctx, "info", fmt.Sprintf("抓取成功 %d/%d：%s", index+1, total, task.MovieName))
		done = append(done, doneItem)
		reportSukebeiFilteredQueueProgress(ctx, "fetch_sukebei_filtered_done", fmt.Sprintf("抓取成功 %d/%d：%s", index+1, total, task.MovieName), pending, running, done, successCount, failedCount)
	}

	reportSukebeiFilteredQueueLog(ctx, "info", fmt.Sprintf("筛选抓取完成：total=%d success=%d failed=%d", total, successCount, failedCount))
	reportSukebeiFilteredQueueProgress(ctx, "fetch_sukebei_filtered_all_done", "筛选抓取任务完成", pending, running, done, successCount, failedCount)
	return nil
}

func buildSukebeiFilteredPendingItems(tasks []*fetchsite.SukebeiFetchTask) []taskctx.QueueItem {
	items := make([]taskctx.QueueItem, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		items = append(items, taskctx.QueueItem{
			MovieJavID: strings.TrimSpace(task.MovieJavID),
			MovieName:  strings.TrimSpace(task.MovieName),
			Seq:        int64(len(items) + 1),
			State:      "pending",
		})
	}
	return items
}

func cloneSukebeiFilteredQueueItems(items []taskctx.QueueItem) []taskctx.QueueItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]taskctx.QueueItem, len(items))
	copy(out, items)
	return out
}

func reportSukebeiFilteredQueueProgress(
	ctx context.Context,
	stage string,
	message string,
	pending []taskctx.QueueItem,
	running []taskctx.QueueItem,
	done []taskctx.QueueItem,
	successCount int,
	failedCount int,
) {
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             stage,
		Message:           message,
		HandledCount:      len(done),
		SuccessCount:      successCount,
		FailedCount:       failedCount,
		QueuedCount:       len(pending) + len(running),
		PendingItems:      cloneSukebeiFilteredQueueItems(pending),
		RunningItems:      cloneSukebeiFilteredQueueItems(running),
		DoneItems:         cloneSukebeiFilteredQueueItems(done),
		CurrentPhaseKey:   "fetch_sukebei",
		PhaseKey:          "fetch_sukebei",
		PhaseHandledCount: len(done),
		PhaseTotalCount:   len(pending) + len(running) + len(done),
		PhaseSuccessCount: successCount,
		PhaseFailedCount:  failedCount,
	})
}

func reportSukebeiFilteredQueueLog(ctx context.Context, level string, message string) {
	taskctx.ReportLog(ctx, taskctx.Log{
		Level:   level,
		Message: message,
		Line:    message,
		At:      time.Now().Unix(),
	})
}
