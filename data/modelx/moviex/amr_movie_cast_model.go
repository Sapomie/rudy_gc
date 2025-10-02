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
	// AmrMovieCastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmrMovieCastModel.
	AmrMovieCastModel interface {
		amrMovieCastModel

		ListCastIDsByMovie(ctx context.Context, movieId int64) ([]int64, error)
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

func (m *customAmrMovieCastModel) ListCastIDsByMovie(ctx context.Context, movieId int64) ([]int64, error) {
	q, args, err := squirrel.
		Select("`cast_id`").
		From(m.tableName()).
		Where("`movie_id` = ?", movieId).
		OrderBy("`cast_id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var ids []int64
	// 列表查询不走缓存
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}
