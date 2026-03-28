package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CCafoModel = (*customCCafoModel)(nil)

type (
	// CCafoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCCafoModel.
	CCafoModel interface {
		cCafoModel
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListAll(ctx context.Context) ([]*CCafo, error)
	}

	customCCafoModel struct {
		*defaultCCafoModel
	}
)

// NewCCafoModel returns a model for the database table.
func NewCCafoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CCafoModel {
	return &customCCafoModel{
		defaultCCafoModel: newCCafoModel(conn, c, opts...),
	}
}

func (m *customCCafoModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customCCafoModel) ListAll(ctx context.Context) ([]*CCafo, error) {
	var rows []*CCafo
	query := "select " + cCafoRows + " from " + m.table + " order by id asc"
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CCafo{}, nil
		}
		return nil, err
	}
	return rows, nil
}
