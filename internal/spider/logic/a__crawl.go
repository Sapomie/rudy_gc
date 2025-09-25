// internal/spiderx/logic/a__crawl.go
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

func (l *CrawlLogic) CrawlActiveQueries() error {
	logx.WithContext(l.ctx).Info("CrawlActiveQueries: begin")

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

	logx.WithContext(l.ctx).Info("CrawlActiveQueries: done")
	return nil
}

func (l *CrawlLogic) CrawlDailyBestinv() error {
	logx.WithContext(l.ctx).Info("CrawlDailyBestinv: begin")
	// TODO: 拉取每日排行原文 -> raw_bestinv；解析 -> c_rank / 其它结构化表；设置保留周期/清理策略
	logx.WithContext(l.ctx).Info("CrawlDailyBestinv: done")
	return nil
}

func (l *CrawlLogic) SyncDailyBestinv() error {
	logx.WithContext(l.ctx).Info("SyncDailyBestinv: begin")
	// TODO: 同步远端/历史榜单；幂等合并；必要时触发详情补抓
	logx.WithContext(l.ctx).Info("SyncDailyBestinv: done")
	return nil
}

// ProcessInventory 解析 raw_inventory -> AItem（HasDetail=NoDetail）并做幂等入库与过滤
func (l *CrawlLogic) ProcessInventory() error {
	// TODO:
	// - 查询 NeedScan=YES 的 raw_inventory
	// - goquery/选择器解析每张卡片：name/javId/cover/prefix/……
	// - 过滤冗余前缀/蓝光标记；TryInsert 幂等入 AItem
	// - 成功后将该 raw_inventory 标记为已扫描（NeedScan=NO；最好同一事务）
	return nil
}
