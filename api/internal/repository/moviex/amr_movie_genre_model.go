package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 暴露接口：补充 TableName / QueryRowNoCacheCtx / QueryRowsNoCacheCtx
type AmrMovieGenreModel interface {
	amrMovieGenreModel

	ListGenreIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error)

	// 供 repo 侧复用的通用方法（与其它 model 一致）
	TableName() string
	QueryRowNoCacheCtx(ctx context.Context, v any, q string, args ...any) error
	QueryRowsNoCacheCtx(ctx context.Context, v any, q string, args ...any) error
}

type customAmrMovieGenreModel struct {
	*defaultAmrMovieGenreModel
}

// New
func NewAmrMovieGenreModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmrMovieGenreModel {
	return &customAmrMovieGenreModel{
		defaultAmrMovieGenreModel: newAmrMovieGenreModel(conn, c, opts...),
	}
}

/* ---------------- 通用补充方法 ---------------- */

// TableName 返回表名
func (m *customAmrMovieGenreModel) TableName() string {
	return m.table
}

// QueryRowNoCacheCtx：走 CachedConn 的无缓存查询
func (m *customAmrMovieGenreModel) QueryRowNoCacheCtx(ctx context.Context, v any, q string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, v, q, args...)
}

// QueryRowsNoCacheCtx：走 CachedConn 的无缓存查询
func (m *customAmrMovieGenreModel) QueryRowsNoCacheCtx(ctx context.Context, v any, q string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, v, q, args...)
}

/* ---------------- 业务扩展 ---------------- */

// ListGenreIDsByMovieJavId 查询某个 movie_jav_id 下的所有 genre_id（不走缓存）
func (m *customAmrMovieGenreModel) ListGenreIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error) {
	q, args, err := squirrel.
		Select("`genre_id`").
		From(m.TableName()).
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
