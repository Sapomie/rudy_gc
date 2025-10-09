package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VFilmModel = (*customVFilmModel)(nil)

type (
	VFilmModel interface {
		vFilmModel

		FindAll(ctx context.Context) ([]*VFilm, error)
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customVFilmModel struct {
		*defaultVFilmModel
	}
)

func NewVFilmModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VFilmModel {
	return &customVFilmModel{
		defaultVFilmModel: newVFilmModel(conn, c, opts...),
	}
}

// FindAll 返回所有 v_film 记录（不走缓存）
func (m *customVFilmModel) FindAll(ctx context.Context) ([]*VFilm, error) {
	q, args, err := squirrel.Select(vFilmRows).From(m.table).ToSql()
	if err != nil {
		return nil, err
	}

	var list []*VFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*VFilm{}, nil
		}
		return nil, err
	}
	return list, nil
}

// TableName 返回 v_film 表名（供外部构建 SQL）
func (m *customVFilmModel) TableName() string {
	return m.table
}

// QueryRowsNoCacheCtx 执行无缓存的多行查询
func (m *customVFilmModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}
