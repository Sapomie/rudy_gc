// data/modelx/moviex/e_item_model_ext.go
package moviex

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type EItemModel interface {
	eItemModel
	ListByDetailStatus(ctx context.Context, hasDetail int64, limit int64) ([]*EItem, error)
	ListByDetailNeedScan(ctx context.Context, needScan int64, limit int64) ([]*EItem, error)
	ListByDownloadCoverStatus(ctx context.Context, hasDownloadCover int64, limit int64) ([]*EItem, error)
	ListByTranslateStatus(ctx context.Context, hasDownloadCover int64, limit int64) ([]*EItem, error)
	ListOldestByLastQueryDetailTime(ctx context.Context, limit int64) ([]*EItem, error)
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
