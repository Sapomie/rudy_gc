// data/modelx/moviex/g_list_model_ext.go
package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GListModel = (*customGListModel)(nil)

type (
	// 在 goctl 生成的 gListModel 基础上扩展
	GListModel interface {
		gListModel

		// 便于外层使用的扩展
		FindAll(ctx context.Context) ([]*GList, error)
		ListByFilters(ctx context.Context, scName string, isCome *int64, offset, limit int64) ([]*GList, error)

		// 暴露只读能力，供 infra 构造 SQL（避免直接依赖未导出字段）
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, v any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, v any, query string, args ...any) error
		ListByMovieJavIds(ctx context.Context, javIds []string) ([]*GList, error)
		ListByMovieJavId(ctx context.Context, javId string) ([]*GList, error)
	}

	customGListModel struct {
		*defaultGListModel
	}
)

func NewGListModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GListModel {
	return &customGListModel{
		defaultGListModel: newGListModel(conn, c, opts...),
	}
}

// ---- 只读辅助 ----
func (m *customGListModel) TableName() string { return m.table }

func (m *customGListModel) QueryRowsNoCacheCtx(ctx context.Context, v any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, v, query, args...)
}

func (m *customGListModel) QueryRowNoCacheCtx(ctx context.Context, v any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, v, query, args...)
}

// ---- 业务扩展 ----
func (m *customGListModel) FindAll(ctx context.Context) ([]*GList, error) {
	sqlStr, args, err := squirrel.Select(gListRows).From(m.table).ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GList
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GList{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGListModel) ListByFilters(ctx context.Context, scName string, isCome *int64, offset, limit int64) ([]*GList, error) {
	w := squirrel.And{}
	if scName != "" {
		w = append(w, squirrel.Like{"sc_name": "%" + scName + "%"})
	}
	if isCome != nil {
		w = append(w, squirrel.Eq{"is_come": *isCome})
	}
	sb := squirrel.Select(gListRows).
		From(m.table).
		Where(w).
		OrderBy("updated_on DESC").
		Offset(uint64(offset)).
		Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GList
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GList{}, nil
		}
		return nil, err
	}
	return rows, nil
}

// ✅ 新增：按 movie_jav_id 批量查询
func (m *customGListModel) ListByMovieJavIds(ctx context.Context, javIds []string) ([]*GList, error) {
	if len(javIds) == 0 {
		return []*GList{}, nil
	}
	sqlStr, args, err := squirrel.
		Select(gListRows).
		From(m.table).
		Where(squirrel.Eq{"movie_jav_id": javIds}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GList
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GList{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGListModel) ListByMovieJavId(ctx context.Context, javId string) ([]*GList, error) {
	sqlStr, args, err := squirrel.
		Select(gListRows).
		From(m.table).
		Where(squirrel.Eq{"movie_jav_id": javId}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GList
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GList{}, nil
		}
		return nil, err
	}
	return rows, nil
}
