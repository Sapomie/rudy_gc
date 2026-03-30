package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaModel = (*customWMediaModel)(nil)

type (
	// WMediaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaModel.
	WMediaModel interface {
		wMediaModel
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customWMediaModel struct {
		*defaultWMediaModel
	}
)

// NewWMediaModel returns a model for the database table.
func NewWMediaModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaModel {
	return &customWMediaModel{
		defaultWMediaModel: newWMediaModel(conn, c, opts...),
	}
}

func (m *customWMediaModel) TableName() string { return m.table }

func (m *customWMediaModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}
