package repo

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
}
