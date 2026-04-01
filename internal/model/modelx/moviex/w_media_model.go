package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaModel = (*customWMediaModel)(nil)

type (
	// WMediaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaModel.
	WMediaModel interface {
		wMediaModel
		ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*WMedia, error)
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

func (m *customWMediaModel) ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*WMedia, error) {
	if len(movieJavIds) == 0 {
		return []*WMedia{}, nil
	}

	query, args, err := squirrel.
		Select(wMediaRows).
		From(m.table).
		Where(squirrel.Eq{"movie_jav_id": movieJavIds}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMedia
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMedia{}, nil
		}
		return nil, err
	}
	return rows, nil
}
