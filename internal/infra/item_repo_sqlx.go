// internal/infra/item_repo_sqlx.go
package infra

import (
	"context"
	"fmt"

	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/repo"
	"rudy_gc/internal/types"
)

var _ repo.ItemRepo = (*ItemRepoSqlx)(nil)

type ItemRepoSqlx struct {
	m spiderx.EItemModel
}

func NewItemRepoSqlx(m spiderx.EItemModel) repo.ItemRepo {
	return &ItemRepoSqlx{m: m}
}

func (r *ItemRepoSqlx) TryInsert(ctx context.Context, it *types.Item) (bool, error) {
	old, err := r.m.FindOneByJavId(ctx, it.JavId)
	if err == nil && old != nil {
		return false, nil
	}
	_, ierr := r.m.Insert(ctx, &spiderx.EItem{
		Id:               it.Id,
		Name:             it.Name,
		JavId:            it.JavId,
		Prefix:           it.Prefix,
		SearchType:       it.SearchType,
		CoverUrl:         it.CoverUrl,
		SearchBy:         it.SearchBy,
		HasDetail:        it.HasDetail,
		HasDownloadCover: it.HasDownloadCover,
		HasChinese:       it.HasChinese,
		DetailNeedScan:   it.DetailNeedScan,
		DetailBirthTime:  it.DetailBirthTime,
		DetailUpdateTime: it.DetailUpdateTime,
		CreatedOn:        it.CreatedOn,
		UpdatedOn:        it.UpdatedOn,
	})
	return ierr == nil, ierr
}

func (r *ItemRepoSqlx) FindByDetailStatus(ctx context.Context, status int64) ([]*types.Item, error) {
	rows, err := r.m.ListByDetailStatus(ctx, status, 10000)
	if err != nil {
		return nil, fmt.Errorf("ListByDetailStatus(%d): %w", status, err)
	}
	out := make([]*types.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, itemRowToType(row))
	}
	return out, nil
}

func (r *ItemRepoSqlx) FindByDetailNeedScan(ctx context.Context, needScan int64) ([]*types.Item, error) {
	rows, err := r.m.ListByDetailNeedScan(ctx, needScan, 10000)
	if err != nil {
		return nil, fmt.Errorf("ListByDetailNeedScan(%d): %w", needScan, err)
	}
	out := make([]*types.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, itemRowToType(row))
	}
	return out, nil
}

func (r *ItemRepoSqlx) UpdateDetailMeta(
	ctx context.Context,
	id int64,
	detailNeedScan int64,
	birthTime int64,
	updateTime int64,
	hasDetail int64,
	updatedOn int64,
) error {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	row.DetailNeedScan = detailNeedScan
	if row.DetailBirthTime == 0 && birthTime > 0 {
		row.DetailBirthTime = birthTime
	}
	if updateTime > 0 {
		row.DetailUpdateTime = updateTime
	}
	row.HasDetail = hasDetail
	row.UpdatedOn = updatedOn

	return r.m.Update(ctx, row)
}

func itemRowToType(row *spiderx.EItem) *types.Item {
	if row == nil {
		return nil
	}
	return &types.Item{
		Id:               row.Id,
		Name:             row.Name,
		JavId:            row.JavId,
		Prefix:           row.Prefix,
		SearchType:       row.SearchType,
		CoverUrl:         row.CoverUrl,
		SearchBy:         row.SearchBy,
		HasDetail:        row.HasDetail,
		HasDownloadCover: row.HasDownloadCover,
		HasChinese:       row.HasChinese,
		DetailNeedScan:   row.DetailNeedScan,
		DetailBirthTime:  row.DetailBirthTime,
		DetailUpdateTime: row.DetailUpdateTime,
		CreatedOn:        row.CreatedOn,
		UpdatedOn:        row.UpdatedOn,
	}
}
