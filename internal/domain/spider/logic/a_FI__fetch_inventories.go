// internal/spiderx/logic/a_FI__fetch_inventories.go
package logic

import (
	"context"
	"rudy_gc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ------------------------ 对外暴露的方法（与旧项目保持一致命名） ------------------------

func (l *CrawlLogic) CrawlActiveSeeds() error {
	logx.WithContext(l.ctx).Info("CrawlActiveSeeds: begin")
	// 1) 抓取库存页原文至 raw_inventory
	if err := l.FetchInventoriesBySeedActive(); err != nil {
		logx.WithContext(l.ctx).Errorf("FetchInventoriesBySeedActive: %v", err)
		return err
	}

	// 2) 解析 raw_inventory -> AItem（HasDetail=NoDetail）并推进断点
	if err := l.ProcessInventory(); err != nil {
		logx.WithContext(l.ctx).Errorf("ProcessInventory: %v", err)
		return err
	}
	logx.WithContext(l.ctx).Info("CrawlActiveSeeds: done")
	return nil
}
