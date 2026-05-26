package loop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

func (l *FetchLoopService) runJavbusFilteredQueue(ctx context.Context, tasks []*fetchsite.JavbusFetchTask) error {
	validTasks := make([]*fetchsite.JavbusFetchTask, 0, len(tasks))
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

	pending := buildJavbusFilteredPendingItems(validTasks)
	running := []taskctx.QueueItem{}
	done := []taskctx.QueueItem{}
	successCount := 0
	failedCount := 0
	total := len(validTasks)

	reportJavbusFilteredQueueLog(ctx, "info", fmt.Sprintf("筛选抓取队列已加载：%d 条", total))
	reportJavbusFilteredQueueProgress(ctx, "fetch_javbus_filtered_queue_ready", fmt.Sprintf("筛选抓取队列已加载：%d 条", total), pending, running, done, successCount, failedCount)

	for index, task := range validTasks {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return err
		}
		if index > 0 {
			if err := l.fetchSiteSvc.SleepRequest(ctx, fetchsite.FetchSiteCodeJavbus); err != nil {
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
		reportJavbusFilteredQueueLog(ctx, "info", fmt.Sprintf("开始抓取 %d/%d：%s", index+1, total, task.MovieName))
		reportJavbusFilteredQueueProgress(ctx, "fetch_javbus_filtered_running", fmt.Sprintf("开始抓取 %d/%d：%s", index+1, total, task.MovieName), pending, running, done, successCount, failedCount)

		_, err := l.fetchQueue.RunSingleJavbusFetchTask(ctx, task.MovieJavID, task.MovieName)
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
			reportJavbusFilteredQueueLog(ctx, "error", fmt.Sprintf("抓取失败 %d/%d：%s | err=%v", index+1, total, task.MovieName, err))
			done = append(done, doneItem)
			reportJavbusFilteredQueueProgress(ctx, "fetch_javbus_filtered_failed", fmt.Sprintf("抓取失败 %d/%d：%s", index+1, total, task.MovieName), pending, running, done, successCount, failedCount)
			continue
		}

		successCount++
		doneItem.State = "success"
		doneItem.CompletedAt = time.Now().Unix()
		reportJavbusFilteredQueueLog(ctx, "info", fmt.Sprintf("抓取成功 %d/%d：%s", index+1, total, task.MovieName))
		done = append(done, doneItem)
		reportJavbusFilteredQueueProgress(ctx, "fetch_javbus_filtered_done", fmt.Sprintf("抓取成功 %d/%d：%s", index+1, total, task.MovieName), pending, running, done, successCount, failedCount)
	}

	reportJavbusFilteredQueueLog(ctx, "info", fmt.Sprintf("筛选抓取完成：total=%d success=%d failed=%d", total, successCount, failedCount))
	reportJavbusFilteredQueueProgress(ctx, "fetch_javbus_filtered_all_done", "筛选抓取任务完成", pending, running, done, successCount, failedCount)
	return nil
}

func buildJavbusFilteredPendingItems(tasks []*fetchsite.JavbusFetchTask) []taskctx.QueueItem {
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

func cloneJavbusFilteredQueueItems(items []taskctx.QueueItem) []taskctx.QueueItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]taskctx.QueueItem, len(items))
	copy(out, items)
	return out
}

func reportJavbusFilteredQueueProgress(
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
		PendingItems:      cloneJavbusFilteredQueueItems(pending),
		RunningItems:      cloneJavbusFilteredQueueItems(running),
		DoneItems:         cloneJavbusFilteredQueueItems(done),
		CurrentPhaseKey:   "fetch_javbus",
		PhaseKey:          "fetch_javbus",
		PhaseHandledCount: len(done),
		PhaseTotalCount:   len(pending) + len(running) + len(done),
		PhaseSuccessCount: successCount,
		PhaseFailedCount:  failedCount,
	})
}

func reportJavbusFilteredQueueLog(ctx context.Context, level string, message string) {
	taskctx.ReportLog(ctx, taskctx.Log{
		Level:   level,
		Message: message,
		Line:    message,
		At:      time.Now().Unix(),
	})
}
