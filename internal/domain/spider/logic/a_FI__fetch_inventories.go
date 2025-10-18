// internal/spiderx/logic/a_FI__fetch_inventories.go
package logic

// ------------------------ 对外暴露的方法（与旧项目保持一致命名） ------------------------

func (l *CrawlLogic) FetchAndParseInventoryBySeed() error {
	l.deps.Log.WithContext(l.ctx).Info("FetchAndParseInventoryBySeed: begin")
	// 1) 抓取库存页原文至 raw_inventory
	if err := l.FetchInventoriesBySeedActive(); err != nil {
		l.deps.Log.WithContext(l.ctx).Errorf("FetchInventoriesBySeedActive: %v", err)
		return err
	}

	// 2) 解析 raw_inventory -> AItem（HasDetail=NoDetail）并推进断点
	if err := l.ParseInventory(); err != nil {
		l.deps.Log.WithContext(l.ctx).Errorf("ParseInventory: %v", err)
		return err
	}
	l.deps.Log.WithContext(l.ctx).Info("FetchAndParseInventoryBySeed: done")
	return nil
}
