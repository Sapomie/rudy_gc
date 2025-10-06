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

		ListGenreIDsByMovie(ctx context.Context, movieId int64) ([]int64, error)
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

func (m *customAmrMovieGenreModel) ListGenreIDsByMovie(ctx context.Context, movieId int64) ([]int64, error) {
	q, args, err := squirrel.
		Select("`genre_id`").
		From(m.tableName()).
		Where("`movie_id` = ?", movieId).
		OrderBy("`genre_id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var ids []int64
	// 列表查询不走缓存，风格与 movie_cast 保持一致
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}
