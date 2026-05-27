// internal/model/modelx/moviex/e_item_model_ext.go
package moviex

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/types"
)

type EItemModel interface {
	eItemModel
	ListByDetailStatus(ctx context.Context, hasDetail int64, limit int64) ([]*EItem, error)
	ListByDetailNeedScan(ctx context.Context, needScan int64, limit int64) ([]*EItem, error)
	ListByDownloadCoverStatus(ctx context.Context, hasDownloadCover int64, limit int64) ([]*EItem, error)
	ListByTranslateStatus(ctx context.Context, hasDownloadCover int64, limit int64) ([]*EItem, error)
	ListOldestByLastQueryDetailTime(ctx context.Context, limit int64) ([]*EItem, error)
	ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.ItemListFilter) ([]*types.Item, error)
	CountAll(ctx context.Context, filter types.ItemListFilter) (int64, error)
	TableName() string
	ExecNoCacheCtx(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type customEItemModel struct {
	*defaultEItemModel
}

func NewEItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) EItemModel {
	return &customEItemModel{
		defaultEItemModel: newEItemModel(conn, c, opts...),
	}
}
func (m *customEItemModel) TableName() string {
	return m.table
}

// ===== 对外方法用通用方法实现 =====

func (m *customEItemModel) ListByDetailStatus(ctx context.Context, hasDetail int64, limit int64) ([]*EItem, error) {
	return m.listByIntFlag(ctx, "has_detail", hasDetail, limit)
}

func (m *customEItemModel) ListByDetailNeedScan(ctx context.Context, needScan int64, limit int64) ([]*EItem, error) {
	return m.listByIntFlag(ctx, "detail_need_scan", needScan, limit)
}

func (m *customEItemModel) ListByDownloadCoverStatus(ctx context.Context, hasDownloadCover int64, limit int64) ([]*EItem, error) {
	return m.listByIntFlag(ctx, "has_download_cover", hasDownloadCover, limit)
}

func (m *customEItemModel) ListByTranslateStatus(ctx context.Context, translateStatus int64, limit int64) ([]*EItem, error) {
	return m.listByIntFlag(ctx, "has_chinese", translateStatus, limit)
}

// 私有通用方法：按某个 int 列精确匹配并限制条数
func (m *customEItemModel) listByIntFlag(ctx context.Context, col string, val, limit int64) ([]*EItem, error) {
	if limit <= 0 {
		limit = 10000
	}
	builder := squirrel.
		Select(eItemRows).
		From(m.tableName()).
		Where(squirrel.Eq{col: val}).
		OrderBy("`id` ASC").
		Limit(uint64(limit))

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*EItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		// go-zero 的 sqlx.ErrNotFound 表示 0 行，按空切片返回即可
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*EItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customEItemModel) ListOldestByLastQueryDetailTime(ctx context.Context, limit int64) ([]*EItem, error) {
	if limit <= 0 {
		limit = 1
	}

	builder := squirrel.
		Select(eItemRows).
		From(m.tableName()).
		OrderBy("`last_query_detail_time` ASC").
		Limit(uint64(limit))

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*EItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*EItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customEItemModel) ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.ItemListFilter) ([]*types.Item, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if strings.TrimSpace(orderBy) == "" {
		orderBy = "`last_query_detail_time` DESC, `id` DESC"
	}

	builder := squirrel.
		Select(eItemRows).
		From(m.tableName())

	builder = applyEItemListFilter(builder, filter)

	sqlStr, args, err := builder.
		OrderBy(orderBy).
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*EItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*types.Item{}, nil
		}
		return nil, err
	}

	return mapEItemRows(rows), nil
}

func (m *customEItemModel) CountAll(ctx context.Context, filter types.ItemListFilter) (int64, error) {
	builder := squirrel.
		Select("COUNT(*)").
		From(m.tableName())

	builder = applyEItemListFilter(builder, filter)

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, sqlStr, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func applyEItemListFilter(builder squirrel.SelectBuilder, filter types.ItemListFilter) squirrel.SelectBuilder {
	if javID := strings.TrimSpace(filter.JavID); javID != "" {
		builder = builder.Where("`jav_id` LIKE ?", "%"+javID+"%")
	}
	if name := strings.TrimSpace(filter.Name); name != "" {
		builder = builder.Where("`name` LIKE ?", "%"+name+"%")
	}
	if filter.HasDetailSet {
		builder = builder.Where(squirrel.Eq{"has_detail": filter.HasDetail})
	}
	if filter.HasDownloadCoverSet {
		builder = builder.Where(squirrel.Eq{"has_download_cover": filter.HasDownloadCover})
	}
	if filter.HasChineseSet {
		builder = builder.Where(squirrel.Eq{"has_chinese": filter.HasChinese})
	}
	if filter.DetailNeedScanSet {
		builder = builder.Where(squirrel.Eq{"detail_need_scan": filter.DetailNeedScan})
	}
	if filter.HasDetailBirthTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"detail_birth_time": filter.DetailBirthTimeFrom})
	}
	if filter.HasDetailBirthTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"detail_birth_time": filter.DetailBirthTimeTo})
	}
	if filter.HasLastQueryDetailTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"last_query_detail_time": filter.LastQueryDetailTimeFrom})
	}
	if filter.HasLastQueryDetailTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"last_query_detail_time": filter.LastQueryDetailTimeTo})
	}
	return builder
}

func mapEItemRows(rows []*EItem) []*types.Item {
	if len(rows) == 0 {
		return []*types.Item{}
	}

	out := make([]*types.Item, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &types.Item{
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
		})
	}
	return out
}
