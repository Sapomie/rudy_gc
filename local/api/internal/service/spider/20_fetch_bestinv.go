package spider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	consts "rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/types"
	"time"
)

func (l *Service) FetchBestinv(ctx context.Context, typ int64, pageEnd int64) error {
	date := time.Now().Add(-8 * time.Hour).Format(time.DateOnly) // 与老项目保持相同的日期口径
	l.deps.Log.WithContext(ctx).Infof("开始抓取 Bestinv：typ=%d, date=%s, 至多 %d 页", typ, date, pageEnd)
	l.reportPhaseProgress(ctx, "bestinv", "bestinv_prepare", "开始抓取榜单", 0, int(pageEnd), 0, 0)

	seenChecksums := make(map[string]struct{})
	for page := int64(1); page <= pageEnd; page++ {
		if err := l.waitIfPaused(ctx); err != nil {
			return err
		}
		content, err := l.fetchBestinvContentByRated(ctx, typ, page)
		if err != nil {
			// 已有的 retry 流程在空列表时会返回 ErrBlankPage，这里直接提前结束
			if errors.Is(err, ErrBlankPage) {
				l.deps.Log.WithContext(ctx).Infof("第 %d 页为空白页，提前停止", page)
				break
			}
			return err
		}
		sum := sha256.Sum256([]byte(content))
		checksum := hex.EncodeToString(sum[:])
		if _, exists := seenChecksums[checksum]; exists {
			l.deps.Log.WithContext(ctx).Infof("第 %d 页与已有榜单页内容重复，榜单分页抓取正常结束", page)
			break
		}
		if err := l.saveBestinv(ctx, typ, date, page, content); err != nil {
			return err
		}
		seenChecksums[checksum] = struct{}{}
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

func (l *Service) fetchBestinvContentByRated(ctx context.Context, typ int64, page int64) (string, error) {
	queryBy := getQueryByBestRatedType(typ)
	if queryBy == "" {
		return "", fmt.Errorf("不支持的 Bestinv 类型: %d", typ)
	}

	queryWithPage := fmt.Sprintf("/%s&page=%d", queryBy, page)
	fullURL, err := l.javlibURL(queryWithPage)
	if err != nil {
		return "", err
	}

	content, err := l.fetchInventoryContentWithRetryWithOptions(ctx, fullURL, inventoryFetchOptions{
		successMessage: func(_ int, _ string) string {
			return fmt.Sprintf("成功获取第%d页内容", page)
		},
	})
	if err != nil {
		return "", err // 这里包含 ErrBlankPage；上层会识别并提前停止
	}
	return content, nil
}

func (l *Service) saveBestinv(ctx context.Context, typ int64, date string, page int64, content string) error {
	queryBy := getQueryByBestRatedType(typ)
	if queryBy == "" {
		return fmt.Errorf("不支持的 Bestinv 类型: %d", typ)
	}
	queryWithPage := fmt.Sprintf("/%s&page=%d", queryBy, page)
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
