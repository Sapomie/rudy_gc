// data/modelx/moviex/e_item_model_ext.go
package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type EItemModel interface {
	eItemModel
	ListByDetailStatus(ctx context.Context, hasDetail int64, limit int64) ([]*EItem, error)
	ListByDetailNeedScan(ctx context.Context, needScan int64, limit int64) ([]*EItem, error)
}

type customEItemModel struct {
	*defaultEItemModel
}

func NewEItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) EItemModel {
	return &customEItemModel{
		defaultEItemModel: newEItemModel(conn, c, opts...),
	}
}

func (m *customEItemModel) ListByDetailStatus(ctx context.Context, hasDetail int64, limit int64) ([]*EItem, error) {
	if limit <= 0 {
		limit = 10000
	}
	builder := squirrel.
		Select(eItemRows).
		From(m.tableName()).
		Where(squirrel.Eq{"has_detail": hasDetail}).
		OrderBy("`id` ASC").
		Limit(uint64(limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*EItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customEItemModel) ListByDetailNeedScan(ctx context.Context, needScan int64, limit int64) ([]*EItem, error) {
	if limit <= 0 {
		limit = 10000
	}
	builder := squirrel.
		Select(eItemRows).
		From(m.tableName()).
		Where(squirrel.Eq{"detail_need_scan": needScan}).
		OrderBy("`id` ASC").
		Limit(uint64(limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*EItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rows, nil
}
