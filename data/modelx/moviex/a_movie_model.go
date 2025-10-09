package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AMovieModel = (*customAMovieModel)(nil)

type (
	// 继承 goctl 生成接口 + 扩展方法
	AMovieModel interface {
		aMovieModel

		ListPageWithTotal(ctx context.Context, offset, limit int64, orderKey string) ([]*AMovie, int64, error)
		CountAll(ctx context.Context) (int64, error)
		FindMoviesByName(ctx context.Context, name string) ([]*AMovie, error)

		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customAMovieModel struct {
		*defaultAMovieModel
	}
)

func NewAMovieModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AMovieModel {
	return &customAMovieModel{
		defaultAMovieModel: newAMovieModel(conn, c, opts...),
	}
}

func (m *customAMovieModel) TableName() string {
	return m.table
}

func (m *customAMovieModel) CountAll(ctx context.Context) (int64, error) {
	q, args, err := squirrel.Select("COUNT(*)").From(m.table).ToSql()
	if err != nil {
		return 0, err
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, q, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customAMovieModel) FindMoviesByName(ctx context.Context, name string) ([]*AMovie, error) {
	q, args, err := squirrel.
		Select(aMovieRows).
		From(m.table).
		Where("`name` = ?", name).
		ToSql()
	if err != nil {
		return nil, err
	}

	var list []*AMovie
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*AMovie{}, nil
		}
		return nil, err
	}
	return list, nil
}

func (m *customAMovieModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAMovieModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}
