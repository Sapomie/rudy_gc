package infra

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/repo"
	"rudy_gc/internal/types"
)

// 确保实现接口
var _ repo.InventoryRepo = (*InventoryRepoSqlx)(nil)

// InventoryRepoSqlx 用 goctl 生成的 DInventoryModel 实现仓储
type InventoryRepoSqlx struct {
	m spiderx.DInventoryModel
}

func NewInventoryRepoSqlx(m spiderx.DInventoryModel) *InventoryRepoSqlx {
	return &InventoryRepoSqlx{m: m}
}

func (r *InventoryRepoSqlx) Upsert(ctx context.Context, inv *types.Inventory) error {
	// 1) 先按 name 查
	exist, err := r.m.FindOneByName(ctx, inv.Name)
	if err == nil && exist != nil {
		// 2) 存在 -> Update
		exist.NeedScan = inv.NeedScan
		exist.Keyword = inv.Keyword
		exist.Parent = inv.Parent
		exist.Page = inv.Page
		exist.Content = inv.Content
		exist.Category = inv.Category
		exist.LastQueryTime = inv.LastQueryTime
		exist.UpdatedOn = inv.UpdatedOn
		return r.m.Update(ctx, exist)
	}

	// 3) 不存在或 ErrNotFound -> Insert
	if err != nil {
		if !errors.Is(err, spiderx.ErrNotFound) {
			return fmt.Errorf("FindOneByName(%s) error: %w", inv.Name, err)
		}
	}

	row := &spiderx.DInventory{
		Name:          inv.Name,
		NeedScan:      inv.NeedScan,
		Keyword:       inv.Keyword,
		Parent:        inv.Parent,
		Page:          inv.Page,
		Content:       inv.Content,
		Category:      inv.Category,
		LastQueryTime: inv.LastQueryTime,
		CreatedOn:     inv.CreatedOn,
		UpdatedOn:     inv.UpdatedOn,
	}
	_, ierr := r.m.Insert(ctx, row)
	return ierr
}

func (r *InventoryRepoSqlx) FindOneByName(ctx context.Context, name string) (*types.Inventory, error) {
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &types.Inventory{
		Id:            row.Id,
		Name:          row.Name,
		NeedScan:      row.NeedScan,
		Keyword:       row.Keyword,
		Parent:        row.Parent,
		Page:          row.Page,
		Content:       row.Content,
		Category:      row.Category,
		LastQueryTime: row.LastQueryTime,
		CreatedOn:     row.CreatedOn,
		UpdatedOn:     row.UpdatedOn,
	}, nil
}
