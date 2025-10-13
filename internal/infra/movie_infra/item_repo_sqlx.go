package movie_infra

import (
	"context"
	"fmt"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/spider_repo"
	"time"

	"rudy_gc/internal/types"
)

var _ spider_repo.ItemRepo = (*ItemRepoSqlx)(nil)

type ItemRepoSqlx struct {
	m moviex.EItemModel
}

const defaultItemLimit int64 = 1_000_000

func NewItemRepoSqlx(m moviex.EItemModel) spider_repo.ItemRepo {
	return &ItemRepoSqlx{m: m}
}

func (r *ItemRepoSqlx) TryInsert(ctx context.Context, it *types.Item) (bool, error) {
	old, err := r.m.FindOneByJavId(ctx, it.JavId)
	if err == nil && old != nil {
		return false, nil
	}
	_, ierr := r.m.Insert(ctx, &moviex.EItem{
		Id:                  it.Id,
		Name:                it.Name,
		JavId:               it.JavId,
		Prefix:              it.Prefix,
		SearchType:          it.SearchType,
		CoverUrl:            it.CoverUrl,
		SearchBy:            it.SearchBy,
		HasDetail:           it.HasDetail,
		HasDownloadCover:    it.HasDownloadCover,
		HasChinese:          it.HasChinese,
		DetailNeedScan:      it.DetailNeedScan,
		DetailBirthTime:     it.DetailBirthTime,
		LastQueryDetailTime: it.LastQueryDetailTime,
		CreatedOn:           it.CreatedOn,
		UpdatedOn:           it.UpdatedOn,
	})
	return ierr == nil, ierr
}

func (r *ItemRepoSqlx) UpdateDetailMeta(
	ctx context.Context,
	id int64,
	needScan int64,
	birthTime int64,
	detailUpdateTime int64,
	updatedOn int64,
	hasDetail int64,
) error {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	// DetailNeedScan 直接覆盖为最新
	row.DetailNeedScan = needScan

	// 首次写入 DetailBirthTime（仅当原值为 0 且传入大于 0 才设置）
	if row.DetailBirthTime == 0 && birthTime > 0 {
		row.DetailBirthTime = birthTime
	}

	// 本次抓取/解析对应的更新时间（>0 才更新）
	if detailUpdateTime > 0 {
		row.LastQueryDetailTime = detailUpdateTime
	}

	// 同步 HasDetail 与 UpdatedOn
	row.HasDetail = hasDetail
	row.UpdatedOn = updatedOn

	return r.m.Update(ctx, row)
}

func (r *ItemRepoSqlx) FindOneByJavId(ctx context.Context, javId string) (*types.Item, error) {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		return nil, err
	}
	return itemRowToType(row), nil
}

// internal/infra/spider_infra/item_repo_sqlx.go
func (r *ItemRepoSqlx) UpdatePartialByJavId(ctx context.Context, javId string, patch types.ItemPatch) error {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		return err // 保留 go-zero 的 ErrNotFound 语义
	}

	changed := false
	if patch.HasDownloadCover != nil && row.HasDownloadCover != *patch.HasDownloadCover {
		row.HasDownloadCover = *patch.HasDownloadCover
		changed = true
	}
	if patch.HasChinese != nil && row.HasChinese != *patch.HasChinese {
		row.HasChinese = *patch.HasChinese
		changed = true
	}
	if patch.HasDetail != nil && row.HasDetail != *patch.HasDetail {
		row.HasDetail = *patch.HasDetail
		changed = true
	}
	if patch.DetailNeedScan != nil && row.DetailNeedScan != *patch.DetailNeedScan {
		row.DetailNeedScan = *patch.DetailNeedScan
		changed = true
	}
	if patch.DetailBirthTime != nil && row.DetailBirthTime != *patch.DetailBirthTime {
		row.DetailBirthTime = *patch.DetailBirthTime
		changed = true
	}
	if patch.LastQueryDetailTime != nil && row.LastQueryDetailTime != *patch.LastQueryDetailTime {
		row.LastQueryDetailTime = *patch.LastQueryDetailTime
		changed = true
	}

	// UpdatedOn：不传则自动填 now（只有在确实有变更时才更新）
	if changed {
		if patch.UpdatedOn != nil {
			row.UpdatedOn = *patch.UpdatedOn
		} else {
			row.UpdatedOn = time.Now().Unix()
		}
		return r.m.Update(ctx, row) // go-zero 自动清缓存
	}
	return nil
}

func (r *ItemRepoSqlx) FindByDetailNeedScan(ctx context.Context, needScan int64) ([]*types.Item, error) {
	return r.listBy(ctx, "ListByDetailNeedScan", r.m.ListByDetailNeedScan, needScan)
}

func (r *ItemRepoSqlx) FindByDownloadCoverStatus(ctx context.Context, downloadCoverStatus int64) ([]*types.Item, error) {
	return r.listBy(ctx, "ListByDownloadCoverStatus", r.m.ListByDownloadCoverStatus, downloadCoverStatus)
}

func (r *ItemRepoSqlx) FindByTranslateStatus(ctx context.Context, translateStatus int64) ([]*types.Item, error) {
	return r.listBy(ctx, "ListByTranslateStatus", r.m.ListByTranslateStatus, translateStatus)
}

func (r *ItemRepoSqlx) FindByDetailStatus(ctx context.Context, status int64) ([]*types.Item, error) {
	return r.listBy(ctx, "ListByDetailStatus", r.m.ListByDetailStatus, status)
}

// 把 modelx 行转成 types
func mapItems(rows []*moviex.EItem) []*types.Item {
	if len(rows) == 0 {
		return nil
	}
	out := make([]*types.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, itemRowToType(row))
	}
	return out
}

// 通用封装：调用指定的列表函数 + 统一错误包装 + 转型
func (r *ItemRepoSqlx) listBy(
	ctx context.Context,
	label string,
	fn func(context.Context, int64, int64) ([]*moviex.EItem, error),
	val int64,
) ([]*types.Item, error) {
	rows, err := fn(ctx, val, defaultItemLimit)
	if err != nil {
		return nil, fmt.Errorf("%s(%d): %w", label, val, err)
	}
	return mapItems(rows), nil
}

// ===== 你的对外方法（变得很简洁）=====

func itemRowToType(row *moviex.EItem) *types.Item {
	if row == nil {
		return nil
	}
	return &types.Item{
		Id:                  row.Id,
		Name:                row.Name,
		JavId:               row.JavId,
		Prefix:              row.Prefix,
		SearchType:          row.SearchType,
		CoverUrl:            row.CoverUrl,
		SearchBy:            row.SearchBy,
		HasDetail:           row.HasDetail,
		HasDownloadCover:    row.HasDownloadCover,
		HasChinese:          row.HasChinese,
		DetailNeedScan:      row.DetailNeedScan,
		DetailBirthTime:     row.DetailBirthTime,
		LastQueryDetailTime: row.LastQueryDetailTime,
		CreatedOn:           row.CreatedOn,
		UpdatedOn:           row.UpdatedOn,
	}
}
