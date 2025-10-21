// internal/spiderx/logic/a_FI__fetch_inventories.go
package logic

import "context"

// ------------------------ 对外暴露的方法（与旧项目保持一致命名） ------------------------

func (l *CrawlLogic) FetchAndParseInventoryBySeedActive(ctx context.Context) error {
	l.deps.Log.WithContext(ctx).Info("FetchAndParseInventoryBySeedActive: begin")
	// 1) 抓取库存页原文至 raw_inventory
	if err := l.FetchInventoriesBySeedActive(ctx); err != nil {
		l.deps.Log.WithContext(ctx).Errorf("FetchInventoriesBySeedActive: %v", err)
		return err
	}

	// 2) 解析 raw_inventory -> AItem（HasDetail=NoDetail）并推进断点
	if err := l.ParseInventory(ctx); err != nil {
		l.deps.Log.WithContext(ctx).Errorf("ParseInventory: %v", err)
		return err
	}
	l.deps.Log.WithContext(ctx).Info("FetchAndParseInventoryBySeedActive: done")
	return nil
}

func (l *CrawlLogic) FetchAndParseInventoryBySeedName(ctx context.Context, seedName string) error {
	l.deps.Log.WithContext(ctx).Info("FetchAndParseInventoryBySeedName: begin")
	// 1) 抓取库存页原文至 raw_inventory
	if err := l.FetchInventoriesBySeedName(ctx, seedName); err != nil {
		l.deps.Log.WithContext(ctx).Errorf("FetchAndParseInventoryBySeedName: %v", err)
		return err
	}

	// 2) 解析 raw_inventory -> AItem（HasDetail=NoDetail）并推进断点
	if err := l.ParseInventory(ctx); err != nil {
		l.deps.Log.WithContext(ctx).Errorf("ParseInventory: %v", err)
		return err
	}
	l.deps.Log.WithContext(ctx).Info("FetchAndParseInventoryBySeedName: done")
	return nil
}
