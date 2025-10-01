package spider_repo

import (
	"context"

	"rudy_gc/internal/types"
)

// InventoryRepo 定义 inventory 的仓储接口
type InventoryRepo interface {
	// Upsert: 以 Name 作为幂等键，不存在则 Insert，存在则 Update（覆盖 Content/时间等）
	Upsert(ctx context.Context, inv *types.Inventory) error

	// FindOneByName: 便于调试/校验
	FindOneByName(ctx context.Context, name string) (*types.Inventory, error)

	// ListNeedScanIDs: 查询 NeedScan=YES 的若干条 id
	ListNeedScanIDs(ctx context.Context, limit int) ([]int64, error)

	// FindOne: 按 id 查询完整记录
	FindOne(ctx context.Context, id int64) (*types.Inventory, error)

	// MarkScanned: 将 NeedScan 改为 NO，并更新 LastQueryTime/UpdatedOn
	MarkScanned(ctx context.Context, id int64, ts int64) error
}
