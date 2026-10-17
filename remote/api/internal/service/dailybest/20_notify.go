package dailybest

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"rudy_gc/remote/api/internal/notify"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) notifyFetchResult(
	ctx context.Context,
	runID int64,
	businessDate string,
	category int64,
	triggerType string,
	finishedAt int64,
	collectErr error,
) {
	if s.notifier == nil {
		return
	}
	var message notify.Message
	var notificationType string
	switch {
	case collectErr != nil && triggerType == dailybesttypes.TriggerScheduled:
		message = s.buildFailureMessage(runID, businessDate, category, finishedAt, collectErr)
		notificationType = "失败"
	case collectErr == nil && triggerType == dailybesttypes.TriggerRecovery:
		message = s.buildRecoveryMessage(runID, businessDate, category, finishedAt)
		notificationType = "恢复"
	default:
		return
	}
	if err := s.notifier.Notify(ctx, message); err != nil {
		log.Printf("DailyBest 抓取任务 %d 发送%s通知异常: %v", runID, notificationType, err)
	}
}

func (s *Service) buildFailureMessage(
	runID int64,
	businessDate string,
	category int64,
	failedAt int64,
	fetchErr error,
) notify.Message {
	return notify.Message{
		Subject: fmt.Sprintf("[Rudy-GC] 每日榜单抓取失败 - %s", businessDate),
		Body: fmt.Sprintf(
			"Rudy-GC 远端每日榜单抓取失败\n\n任务 ID：%d\n业务日期：%s\n榜单分类：%s\n失败时间：%s\n错误信息：%s\n详情页：%s\n",
			runID,
			businessDate,
			categoryName(category),
			s.formatNotificationTime(failedAt),
			fetchErr.Error(),
			s.fetchRunDetailURL(runID),
		),
	}
}

func (s *Service) buildRecoveryMessage(
	runID int64,
	businessDate string,
	category int64,
	recoveredAt int64,
) notify.Message {
	return notify.Message{
		Subject: fmt.Sprintf("[Rudy-GC] 每日榜单抓取已恢复 - %s", businessDate),
		Body: fmt.Sprintf(
			"Rudy-GC 远端每日榜单抓取已恢复\n\n任务 ID：%d\n业务日期：%s\n榜单分类：%s\n恢复时间：%s\n详情页：%s\n",
			runID,
			businessDate,
			categoryName(category),
			s.formatNotificationTime(recoveredAt),
			s.fetchRunDetailURL(runID),
		),
	}
}

func (s *Service) formatNotificationTime(timestamp int64) string {
	return time.Unix(timestamp, 0).In(s.loc).Format("2006-01-02 15:04:05 MST")
}

func (s *Service) fetchRunDetailURL(runID int64) string {
	return strings.TrimRight(strings.TrimSpace(s.detailBaseURL), "/") +
		"/dailybestfetchrundetail/" + strconv.FormatInt(runID, 10)
}

func categoryName(category int64) string {
	if category == contracts.DailyBestCategoryMonth {
		return "月榜（1）"
	}
	return strconv.FormatInt(category, 10)
}
