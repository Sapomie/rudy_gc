package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GScStatModel = (*customGScStatModel)(nil)

type (
	GScStatModel interface {
		gScStatModel
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customGScStatModel struct {
		*defaultGScStatModel
	}
)

func NewGScStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GScStatModel {
	return &customGScStatModel{
		defaultGScStatModel: newGScStatModel(conn, c, opts...),
	}
}

func (m *customGScStatModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customGScStatModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customGScStatModel) TableName() string {
	return m.table
}
