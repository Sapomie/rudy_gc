package spider

import (
	"context"
	"fmt"
	"rudy_gc/internal/model/modelx/moviex"
	"time"

	"rudy_gc/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ItemRepoSqlx struct {
	m moviex.EItemModel
}

const defaultItemLimit int64 = 1_000_000

// 按 LastQueryDetailTime 最早排序，取最前面的 num 条
func (r *ItemRepoSqlx) FindOldestByLastQueryDetailTime(ctx context.Context, num int64) ([]*types.Item, error) {
	if num <= 0 {
		num = 1
	}

	rows, err := r.m.ListOldestByLastQueryDetailTime(ctx, num)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sqlx.ErrNotFound
	}

	return mapItems(rows), nil
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

func (r *ItemRepoSqlx) FindByTranslateStatuses(ctx context.Context, statuses []int64) ([]*types.Item, error) {
	if len(statuses) == 0 {
		return []*types.Item{}, nil
	}

	seen := make(map[int64]struct{}, len(statuses))
	dedupStatuses := make([]int64, 0, len(statuses))
	for _, status := range statuses {
		if _, ok := seen[status]; ok {
			continue
		}
		seen[status] = struct{}{}
		dedupStatuses = append(dedupStatuses, status)
	}

	out := make([]*types.Item, 0)
	seenJavIDs := make(map[string]struct{})
	for _, status := range dedupStatuses {
		rows, err := r.listBy(ctx, "ListByTranslateStatus", r.m.ListByTranslateStatus, status)
		if err != nil {
			return nil, err
		}
		for _, item := range rows {
			if item == nil {
				continue
			}
			javID := item.JavId
			if javID == "" {
				continue
			}
			if _, ok := seenJavIDs[javID]; ok {
				continue
			}
			seenJavIDs[javID] = struct{}{}
			out = append(out, item)
		}
	}
	return out, nil
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
