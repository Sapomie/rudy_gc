package moviex

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAlbumModel = (*customTAlbumModel)(nil)

type (
	// TAlbumModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAlbumModel.
	TAlbumModel interface {
		tAlbumModel
		ListAll(ctx context.Context) ([]*TAlbum, error)
	}

	customTAlbumModel struct {
		*defaultTAlbumModel
	}
)

// NewTAlbumModel returns a model for the database table.
func NewTAlbumModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAlbumModel {
	return &customTAlbumModel{
		defaultTAlbumModel: newTAlbumModel(conn, c, opts...),
	}
}

func (m *customTAlbumModel) ListAll(ctx context.Context) ([]*TAlbum, error) {
	query := "SELECT " + tAlbumRows + " FROM " + strings.Trim(m.table, "`") + " ORDER BY `id` ASC"
	var rows []*TAlbum
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TAlbum{}, nil
		}
		return nil, err
	}
	return rows, nil
}
