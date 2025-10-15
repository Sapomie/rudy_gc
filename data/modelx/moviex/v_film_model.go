package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VFilmModel = (*customVFilmModel)(nil)

type (
	VFilmModel interface {
		vFilmModel

		FindAll(ctx context.Context, removedStatus int64) ([]*VFilm, error)
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customVFilmModel struct {
		*defaultVFilmModel
	}
)

func NewVFilmModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VFilmModel {
	return &customVFilmModel{
		defaultVFilmModel: newVFilmModel(conn, c, opts...),
	}
}

func (m *customVFilmModel) FindAll(ctx context.Context, removedStatus int64) ([]*VFilm, error) {
	builder := squirrel.Select(vFilmRows).From(m.table)
	if removedStatus > 0 {
		builder = builder.Where(squirrel.Eq{"is_removed": removedStatus})
	}
	q, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var list []*VFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*VFilm{}, nil
		}
		return nil, err
	}
	return list, nil
}

func (m *customVFilmModel) TableName() string { return m.table }

func (m *customVFilmModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customVFilmModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}
