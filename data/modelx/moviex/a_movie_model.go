package moviex

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AMovieModel = (*customAMovieModel)(nil)

type (
	// AMovieModel 接口：继承 goctl 生成的 + 扩展方法
	AMovieModel interface {
		aMovieModel

		ListPage(ctx context.Context, limit, offset int64) ([]*AMovie, error)
		CountAll(ctx context.Context) (int64, error)
	}

	customAMovieModel struct {
		*defaultAMovieModel
	}
)

// NewAMovieModel returns a model for the database table.
func NewAMovieModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AMovieModel {
	return &customAMovieModel{
		defaultAMovieModel: newAMovieModel(conn, c, opts...),
	}
}

// ListPage 分页查询（按 id DESC）
func (m *customAMovieModel) ListPage(ctx context.Context, limit, offset int64) ([]*AMovie, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query, args, err := squirrel.
		Select(aMovieRows). // goctl 已生成的行列表
		From(m.table).      // defaultAMovieModel 的 table 字段
		OrderBy("id DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*AMovie
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// CountAll 返回总数
func (m *customAMovieModel) CountAll(ctx context.Context) (int64, error) {
	query, args, err := squirrel.
		Select("COUNT(*)").
		From(m.table).
		ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}
