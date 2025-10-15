// data/modelx/moviex/amr_movie_cast_model_ext.go
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

		// 便于上层复用（和 go-zero 生成模型保持一致）
		TableName() string
		QueryRowNoCacheCtx(ctx context.Context, v any, q string, args ...any) error
		QueryRowsNoCacheCtx(ctx context.Context, v any, q string, args ...any) error

		// 业务扩展
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

// TableName 直接返回底层表名
func (m *customAmrMovieCastModel) TableName() string {
	return m.table
}

// QueryRowNoCacheCtx 统一使用 CachedConn 的无缓存查询
func (m *customAmrMovieCastModel) QueryRowNoCacheCtx(ctx context.Context, v any, q string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, v, q, args...)
}

// QueryRowsNoCacheCtx 统一使用 CachedConn 的无缓存查询
func (m *customAmrMovieCastModel) QueryRowsNoCacheCtx(ctx context.Context, v any, q string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, v, q, args...)
}

// ListCastIDsByMovieJavId 查询某个 movie_jav_id 下的所有 cast_id（不走缓存）
func (m *customAmrMovieCastModel) ListCastIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error) {
	query, args, err := squirrel.
		Select("`cast_id`").
		From(m.TableName()).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		OrderBy("`cast_id` ASC").
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
