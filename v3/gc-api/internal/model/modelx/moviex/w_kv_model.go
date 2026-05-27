package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WKvModel = (*customWKvModel)(nil)

type (
	WKvModel interface {
		wKvModel
		ListAll(ctx context.Context) ([]*WKv, error)
		TableName() string
	}

	customWKvModel struct {
		*defaultWKvModel
	}
)

func NewWKvModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WKvModel {
	return &customWKvModel{
		defaultWKvModel: newWKvModel(conn, c, opts...),
	}
}

func (m *customWKvModel) TableName() string {
	return m.table
}

func (m *customWKvModel) ListAll(ctx context.Context) ([]*WKv, error) {
	query := "select " + wKvRows + " from " + m.table + " order by `id` asc"
	var rows []*WKv
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if err == sqlx.ErrNotFound {
			return []*WKv{}, nil
		}
		return nil, err
	}
	return rows, nil
}
