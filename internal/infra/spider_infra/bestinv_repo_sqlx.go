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

func (r *BestinvRepoSqlx) Upsert(ctx context.Context, b *types.Bestinv) error {
	// 先按 Name（唯一键）查询
	row, err := r.m.FindOneByName(ctx, b.Name)
	switch {
	case err == nil && row != nil:
		// 更新：保留原 CreatedOn，只刷新其它字段
		row.NeedScan = b.NeedScan
		row.NeedRankCheck = b.NeedRankCheck
		row.Category = b.Category
		row.Page = b.Page
		row.DayNumber = b.DayNumber
		row.Content = b.Content
		row.LastQueryTime = b.LastQueryTime
		row.Date = b.Date
		// row.CreatedOn 保留旧值
		row.UpdatedOn = b.UpdatedOn

		if uerr := r.m.Update(ctx, row); uerr != nil {
			return fmt.Errorf("update d_bestinv(%s) failed: %w", b.Name, uerr)
		}
		return nil

	case err != nil && !errors.Is(err, spiderx.ErrNotFound):
		// 真实错误
		return fmt.Errorf("find d_bestinv by name failed: %w", err)

	default:
		// 不存在则插入
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
