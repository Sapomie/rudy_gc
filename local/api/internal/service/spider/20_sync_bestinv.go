package spider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/local/api/internal/consts"
	dailybestsync "rudy_gc/local/api/internal/service/dailybest"
	"rudy_gc/local/api/internal/types"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
)

func (l *Service) SyncBestinv(ctx context.Context, targets []dailybesttypes.SyncTarget) (*dailybestsync.Result, error) {
	if l.dailyBestSync == nil {
		return nil, fmt.Errorf("DailyBest API 同步服务未初始化")
	}
	return l.dailyBestSync.Sync(ctx, targets, l.importDailyBestRanks)
}

func (l *Service) importDailyBestRanks(ctx context.Context, batch dailybestsync.Batch) error {
	now := time.Now().Unix()
	searchBy := fmt.Sprintf("dailybest-api:%d", batch.Manifest.FetchRunID)
	for _, rank := range batch.Ranks {
		javID := strings.TrimSpace(rank.MovieJavID)
		if javID == "" {
			return fmt.Errorf("DailyBest fetch_run_id=%d 存在空影片标识", batch.Manifest.FetchRunID)
		}
		prefix, _ := splitPrefixAndNumber(javID)
		if isRedundantPrefix(prefix) {
			continue
		}
		item := &types.Item{
			Name: javID, JavId: javID, Prefix: prefix,
			SearchType: consts.ItemSearchByBestMonth, SearchBy: searchBy,
			HasDetail: consts.ItemDetailNone, HasDownloadCover: consts.ItemCoverNone,
			HasChinese: consts.ItemChineseNone, DetailNeedScan: consts.ItemDetailStatusNeedScan,
			DetailBirthTime: 0, LastQueryDetailTime: 0, CreatedOn: now, UpdatedOn: now,
		}
		if _, err := l.deps.ItemRepo.TryInsert(ctx, item); err != nil {
			return fmt.Errorf("保存 DailyBest fetch_run_id=%d 影片 %s 失败: %w", batch.Manifest.FetchRunID, javID, err)
		}
	}
	return nil
}
