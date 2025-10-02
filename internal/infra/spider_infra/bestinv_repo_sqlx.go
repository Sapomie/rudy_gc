package spider_infra

import (
	"context"
	"errors"
	"fmt"

	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/repo/spider_repo"
	"rudy_gc/internal/types"
)

var _ spider_repo.BestinvRepo = (*BestinvRepoSqlx)(nil)

type BestinvRepoSqlx struct {
	m spiderx.DBestinvModel
}

func NewBestinvRepoSqlx(m spiderx.DBestinvModel) spider_repo.BestinvRepo {
	return &BestinvRepoSqlx{m: m}
}

// ===== 已有：Upsert（保持不变）=====
func (r *BestinvRepoSqlx) Upsert(ctx context.Context, b *types.Bestinv) error {
	row, err := r.m.FindOneByName(ctx, b.Name)
	switch {
	case err == nil && row != nil:
		// 更新：保留 CreatedOn
		row.NeedScan = b.NeedScan
		row.NeedRankCheck = b.NeedRankCheck
		row.Category = b.Category
		row.Page = b.Page
		row.DayNumber = b.DayNumber
		row.Content = b.Content
		row.LastQueryTime = b.LastQueryTime
		row.Date = b.Date
		row.UpdatedOn = b.UpdatedOn
		if uerr := r.m.Update(ctx, row); uerr != nil {
			return fmt.Errorf("update d_bestinv(%s) failed: %w", b.Name, uerr)
		}
		return nil

	case err != nil && !errors.Is(err, spiderx.ErrNotFound):
		return fmt.Errorf("find d_bestinv by name failed: %w", err)

	default:
		toIns := &spiderx.DBestinv{
			Name:          b.Name,
			NeedScan:      b.NeedScan,
			NeedRankCheck: b.NeedRankCheck,
			Category:      b.Category,
			Page:          b.Page,
			DayNumber:     b.DayNumber,
			Content:       b.Content,
			LastQueryTime: b.LastQueryTime,
			Date:          b.Date,
			CreatedOn:     b.CreatedOn,
			UpdatedOn:     b.UpdatedOn,
		}
		if _, ierr := r.m.Insert(ctx, toIns); ierr != nil {
			return fmt.Errorf("insert d_bestinv(%s) failed: %w", b.Name, ierr)
		}
		return nil
	}
}

func (r *BestinvRepoSqlx) ListNeedScanIDs(ctx context.Context, limit int) ([]int64, error) {
	ids, err := r.m.ListNeedScanIDs(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("d_bestinv ListNeedScanIDs: %w", err)
	}
	return ids, nil
}

func (r *BestinvRepoSqlx) FindOne(ctx context.Context, id int64) (*types.Bestinv, error) {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.Bestinv{
		Id:            row.Id,
		Name:          row.Name,
		NeedScan:      row.NeedScan,
		NeedRankCheck: row.NeedRankCheck,
		Category:      row.Category,
		Page:          row.Page,
		DayNumber:     row.DayNumber,
		Content:       row.Content,
		LastQueryTime: row.LastQueryTime,
		Date:          row.Date,
		CreatedOn:     row.CreatedOn,
		UpdatedOn:     row.UpdatedOn,
	}, nil
}

// ===== 新增：标记已扫描 =====
// 说明：只把 NeedScan 改为“已扫描”的值（沿用你 types 常量），并更新 UpdatedOn。
func (r *BestinvRepoSqlx) MarkScanned(ctx context.Context, id int64, ts int64) error {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	row.NeedScan = types.BestinvNoNeedScan // 使用你在 internal/types 里定义的常量
	row.UpdatedOn = ts
	return r.m.Update(ctx, row)
}
