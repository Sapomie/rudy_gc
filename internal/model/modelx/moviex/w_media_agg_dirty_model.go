package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaAggDirtyModel = (*customWMediaAggDirtyModel)(nil)

type (
	// WMediaAggDirtyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaAggDirtyModel.
	WMediaAggDirtyModel interface {
		wMediaAggDirtyModel
		ListAll(ctx context.Context, limit int) ([]*WMediaAggDirty, error)
		TouchDay(ctx context.Context, bucketDay int64, scopeKey string, now int64) error
	}

	customWMediaAggDirtyModel struct {
		*defaultWMediaAggDirtyModel
	}
)

// NewWMediaAggDirtyModel returns a model for the database table.
func NewWMediaAggDirtyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaAggDirtyModel {
	return &customWMediaAggDirtyModel{
		defaultWMediaAggDirtyModel: newWMediaAggDirtyModel(conn, c, opts...),
	}
}

func (m *customWMediaAggDirtyModel) ListAll(ctx context.Context, limit int) ([]*WMediaAggDirty, error) {
	query := "select " + wMediaAggDirtyRows + " from " + m.table + " order by `bucket_day` asc"
	args := make([]any, 0, 1)
	if limit > 0 {
		query += " limit ?"
		args = append(args, limit)
	}
	var rows []*WMediaAggDirty
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*WMediaAggDirty{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaAggDirtyModel) TouchDay(ctx context.Context, bucketDay int64, scopeKey string, now int64) error {
	row, err := m.FindOneByBucketDay(ctx, bucketDay)
	if err == nil && row != nil {
		row.ScopeKey = scopeKey
		row.UpdatedOn = now
		return m.Update(ctx, row)
	}
	if err != nil && err != ErrNotFound {
		return err
	}
	_, err = m.Insert(ctx, &WMediaAggDirty{
		BucketDay: bucketDay,
		ScopeKey:  scopeKey,
		CreatedOn: now,
		UpdatedOn: now,
	})
	return err
}
