package logic

import (
	"fmt"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"time"
)

func (l *CrawlLogic) ParseBestinv() error {
	log := l.deps.Log.WithContext(l.ctx)

	ids, err := l.deps.BestinvRepo.ListNeedScanIDs(l.ctx, 1000) // 防止一次性全扫过多
	if err != nil {
		return fmt.Errorf("查询待扫描的 Bestinv 失败: %w", err)
	}
	if len(ids) == 0 {
		log.Info("ParseBestinv: 没有需要扫描的 Bestinv")
		return nil
	}
	log.Infof("ParseBestinv: 共有 %d 个 Bestinv 需要扫描", len(ids))

	for _, id := range ids {
		b, err := l.deps.BestinvRepo.FindOne(l.ctx, id)
		if err != nil {
			return fmt.Errorf("读取 Bestinv 失败 id=%d: %w", id, err)
		}
		if err := l.makeAndInsertItemsByBestinv(b); err != nil {
			return err
		}
		// 标记已扫描
		if err := l.deps.BestinvRepo.MarkScanned(l.ctx, id, time.Now().Unix()); err != nil {
			return fmt.Errorf("标记 Bestinv 已扫描失败 id=%d: %w", id, err)
		}
	}

	log.Info("ParseBestinv: 完成")
	return nil

}

func (l *CrawlLogic) makeAndInsertItemsByBestinv(b *types.Bestinv) error {
	if b == nil {
		return fmt.Errorf("nil bestinv")
	}
	return l.makeAndInsertItems(b.Content, b.Name, getBestinvSearchType(b.Category))
}

func getBestinvSearchType(category int64) int64 {
	switch category {
	case consts.BestCategoryMonth:
		return consts.ItemSearchByBestMonth
	case consts.BestCategoryAllTime:
		return consts.ItemSearchByBestAllTime
	default:
		return 0
	}
}
