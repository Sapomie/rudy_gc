package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BmMinfoModel = (*customBmMinfoModel)(nil)

type (
	// BmMinfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmMinfoModel.
	BmMinfoModel interface {
		bmMinfoModel
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customBmMinfoModel struct {
		*defaultBmMinfoModel
	}
)

// NewBmMinfoModel returns a model for the database table.
func NewBmMinfoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmMinfoModel {
	return &customBmMinfoModel{
		defaultBmMinfoModel: newBmMinfoModel(conn, c, opts...),
	}
}

// TableName 返回 bm_minfo 表名（供外部构建 SQL）
func (m *customBmMinfoModel) TableName() string {
	return m.table
}

// QueryRowsNoCacheCtx 直接执行无缓存的多行查询
func (m *customBmMinfoModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}
