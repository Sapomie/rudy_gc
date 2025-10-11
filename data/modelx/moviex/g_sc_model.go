package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GScModel = (*customGScModel)(nil)

type (
	// GScModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGScModel.
	GScModel interface {
		gScModel

		FindAll(ctx context.Context) ([]*GSc, error)
		ListTopNByScTime(ctx context.Context, n uint64) ([]*GSc, error)
	}

	customGScModel struct {
		*defaultGScModel
	}
)

// NewGScModel returns a model for the database table.
func NewGScModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GScModel {
	return &customGScModel{
		defaultGScModel: newGScModel(conn, c, opts...),
	}
}

func (m *customGScModel) ListTopNByScTime(ctx context.Context, n uint64) ([]*GSc, error) {
	sqlStr, args, err := squirrel.
		Select(gScRows).
		From(m.table).
		OrderBy("sc_time DESC").
		Limit(n).
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGScModel) FindAll(ctx context.Context) ([]*GSc, error) {
	sqlStr, args, err := squirrel.
		Select(gScRows).
		From(m.table).
		OrderBy("sc_time DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}
