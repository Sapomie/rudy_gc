package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmrMovieGenreModel = (*customAmrMovieGenreModel)(nil)

type (
	AmrMovieGenreModel interface {
		amrMovieGenreModel

		ListGenreIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error)
	}

	customAmrMovieGenreModel struct {
		*defaultAmrMovieGenreModel
	}
)

// NewAmrMovieGenreModel returns a model for the database table.
func NewAmrMovieGenreModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmrMovieGenreModel {
	return &customAmrMovieGenreModel{
		defaultAmrMovieGenreModel: newAmrMovieGenreModel(conn, c, opts...),
	}
}

// ListGenreIDsByMovieJavId 查询某个 movie_jav_id 下的所有 genre_id（不走缓存）
func (m *customAmrMovieGenreModel) ListGenreIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error) {
	q, args, err := squirrel.
		Select("`genre_id`").
		From(m.tableName()).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		OrderBy("`genre_id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}
