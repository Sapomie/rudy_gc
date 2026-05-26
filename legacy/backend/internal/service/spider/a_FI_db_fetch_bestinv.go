package spider

import (
	"context"
	"errors"
	"fmt"
	consts "rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"time"
)

func (l *CrawlLogic) FetchBestinv(ctx context.Context, typ int64, pageEnd int64) error {
	date := time.Now().Add(-8 * time.Hour).Format(time.DateOnly) // 与老项目保持相同的日期口径
	l.deps.Log.WithContext(ctx).Infof("开始抓取 Bestinv：typ=%d, date=%s, 至多 %d 页", typ, date, pageEnd)
	l.reportPhaseProgress(ctx, "bestinv", "bestinv_prepare", "开始抓取榜单", 0, int(pageEnd), 0, 0)

	for page := int64(1); page <= pageEnd; page++ {
		if err := l.waitIfPaused(ctx); err != nil {
			return err
		}
		if err := l.fetchBestinvByRated(ctx, typ, date, page); err != nil {
			// 已有的 retry 流程在空列表时会返回 ErrBlankPage，这里直接提前结束
			if errors.Is(err, ErrBlankPage) {
				l.deps.Log.WithContext(ctx).Infof("第 %d 页为空白页，提前停止", page)
				break
			}
			return err
		}
		l.reportPhaseProgress(
			ctx,
			"bestinv",
			"bestinv_page_done",
			fmt.Sprintf("已抓取榜单第 %d 页", page),
			int(page),
			int(pageEnd),
			int(page),
			0,
		)
		if err := l.sleepWithContext(ctx, getRandomSleepDuration()); err != nil {
			return err
		}
	}
	l.deps.Log.WithContext(ctx).Info("抓取 Bestinv 完成")
	return nil
}

func (l *CrawlLogic) fetchBestinvByRated(ctx context.Context, typ int64, date string, page int64) error {
	queryBy := getQueryByBestRatedType(typ)
	if queryBy == "" {
		return fmt.Errorf("不支持的 Bestinv 类型: %d", typ)
	}

	queryWithPage := fmt.Sprintf("/%s&page=%d", queryBy, page)
	fullURL := fmt.Sprintf("https://%s/cn%s", l.deps.Config.Fetcher.JavAddress, queryWithPage)

	content, err := l.fetchInventoryContentWithRetryWithOptions(ctx, fullURL, inventoryFetchOptions{
		successMessage: func(_ int, _ string) string {
			return fmt.Sprintf("成功获取第%d页内容", page)
		},
	})
	if err != nil {
		return err // 这里包含 ErrBlankPage；上层会识别并提前停止
	}

	now := time.Now().Unix()
	day := consts.GetRankDayNumber(date)

	best := &types.Bestinv{
		Name:          fmt.Sprintf("%s&date=%s", queryWithPage, date),
		NeedScan:      consts.BestinvNeedScan,
		NeedRankCheck: consts.BestinvNeedRankCheck,
		Category:      typ,
		Page:          page,
		DayNumber:     day,
		Content:       content,
		LastQueryTime: now,
		Date:          date,
		CreatedOn:     now,
		UpdatedOn:     now,
	}

	if err := l.deps.BestinvRepo.Upsert(ctx, best); err != nil {
		return fmt.Errorf("写入 Bestinv 失败(page=%d): %w", page, err)
	}
	return nil
}

// 与旧项目等价的路由片段
func getQueryByBestRatedType(typ int64) string {
	switch typ {
	case consts.BestCategoryAllTime:
		return consts.SearchByBestRatedAllTIme
	case consts.BestCategoryMonth:
		return consts.SearchByBestRatedMonth
	default:
		return ""
	}
}
