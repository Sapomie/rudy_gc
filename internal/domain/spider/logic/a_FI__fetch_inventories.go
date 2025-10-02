// internal/spiderx/logic/a_FI__fetch_inventories.go
package logic

import (
	"context"
	"rudy_gc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CrawlLogic struct {
	ctx  context.Context
	deps *svc.Deps
}

func NewCrawlLogic(ctx context.Context, deps *svc.Deps) *CrawlLogic {
	return &CrawlLogic{ctx: ctx, deps: deps}
}

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

func (l *CrawlLogic) SyncDailyBestinv() error {
	logx.WithContext(l.ctx).Info("SyncDailyBestinv: begin")
	// TODO: 同步远端/历史榜单；幂等合并；必要时触发详情补抓
	logx.WithContext(l.ctx).Info("SyncDailyBestinv: done")
	return nil
}
