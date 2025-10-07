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

		CountAll(ctx context.Context) (int64, error)
		FindMoviesByName(ctx context.Context, name string) ([]*AMovie, error)
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
