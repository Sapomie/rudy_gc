package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmrMovieCastModel = (*customAmrMovieCastModel)(nil)

type (
	// AmrMovieCastModel 可扩展的接口
	AmrMovieCastModel interface {
		amrMovieCastModel

		ListCastIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error)
	}

	customAmrMovieCastModel struct {
		*defaultAmrMovieCastModel
	}
)

// NewAmrMovieCastModel returns a model for the database table.
func NewAmrMovieCastModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmrMovieCastModel {
	return &customAmrMovieCastModel{
		defaultAmrMovieCastModel: newAmrMovieCastModel(conn, c, opts...),
	}
}

// ListCastIDsByMovieJavId 查询某个 movie_jav_id 下的所有 cast_id
func (m *customAmrMovieCastModel) ListCastIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error) {
	query, args, err := squirrel.
		Select("cast_id").
		From(m.tableName()).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		OrderBy("cast_id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}
