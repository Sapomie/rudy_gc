package spider_infra

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/repo/spider_repo"
	"rudy_gc/internal/types"
)

var _ spider_repo.InventoryRepo = (*InventoryRepoSqlx)(nil)

type InventoryRepoSqlx struct {
	m spiderx.DInventoryModel
}

func NewInventoryRepoSqlx(m spiderx.DInventoryModel) *InventoryRepoSqlx {
	return &InventoryRepoSqlx{m: m}
}

func (r *InventoryRepoSqlx) Upsert(ctx context.Context, inv *types.Inventory) error {
	row, err := r.m.FindOneByName(ctx, inv.Name)
	if err == nil && row != nil {
		row.NeedScan = inv.NeedScan
		row.Keyword = inv.Keyword
		row.Parent = inv.Parent
		row.Page = inv.Page
		row.Content = inv.Content
		row.Category = inv.Category
		row.LastQueryTime = inv.LastQueryTime
		row.UpdatedOn = inv.UpdatedOn
		return r.m.Update(ctx, row)
	}
	if err != nil && !errors.Is(err, spiderx.ErrNotFound) {
		return fmt.Errorf("FindOneByName(%s): %w", inv.Name, err)
	}

	_, ierr := r.m.Insert(ctx, &spiderx.DInventory{
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
	})
	return ierr
}

func (r *InventoryRepoSqlx) FindOneByName(ctx context.Context, name string) (*types.Inventory, error) {
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return invRowToType(row), nil
}

func (r *InventoryRepoSqlx) FindOne(ctx context.Context, id int64) (*types.Inventory, error) {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return invRowToType(row), nil
}

func (r *InventoryRepoSqlx) ListNeedScanIDs(ctx context.Context, limit int) ([]int64, error) {
	return r.m.ListNeedScanIDs(ctx, int64(limit))
}

func (r *InventoryRepoSqlx) MarkScanned(ctx context.Context, id int64, ts int64) error {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	row.NeedScan = types.InventoryNoNeedScan
	row.UpdatedOn = ts
	return r.m.Update(ctx, row)
}

func invRowToType(row *spiderx.DInventory) *types.Inventory {
	if row == nil {
		return nil
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
	}
}
