package logic

import (
	"errors"
	"fmt"
	consts "rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *CrawlLogic) FetchBestinv(typ int64, pageEnd int64) error {
	date := time.Now().Add(-8 * time.Hour).Format(time.DateOnly) // 与老项目保持相同的日期口径
	logx.WithContext(l.ctx).Infof("开始抓取 Bestinv：typ=%d, date=%s, 至多 %d 页", typ, date, pageEnd)

	for page := int64(1); page <= pageEnd; page++ {
		if err := l.fetchBestinvByRated(typ, date, page); err != nil {
			// 已有的 retry 流程在空列表时会返回 ErrBlankPage，这里直接提前结束
			if errors.Is(err, ErrBlankPage) {
				logx.WithContext(l.ctx).Infof("第 %d 页为空白页，提前停止", page)
				break
			}
			return err
		}
		time.Sleep(getRandomSleepDuration())
	}
	logx.WithContext(l.ctx).Info("抓取 Bestinv 完成")
	return nil
}

func (l *CrawlLogic) fetchBestinvByRated(typ int64, date string, page int64) error {
	queryBy := getQueryByBestRatedType(typ)
	if queryBy == "" {
		return fmt.Errorf("不支持的 Bestinv 类型: %d", typ)
	}

	queryWithPage := fmt.Sprintf("/%s&page=%d", queryBy, page)
	fullURL := fmt.Sprintf("https://%s/cn%s", l.deps.Config.Fetcher.JavAddress, queryWithPage)

	content, err := l.fetchInventoryContentWithRetry(fullURL)
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

	if err := l.deps.BestinvRepo.Upsert(l.ctx, best); err != nil {
		return fmt.Errorf("写入 Bestinv 失败(page=%d): %w", page, err)
	}

	logx.WithContext(l.ctx).Infof("已抓取 Bestinv 第 %d 页：%s", page, best.Name)
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
