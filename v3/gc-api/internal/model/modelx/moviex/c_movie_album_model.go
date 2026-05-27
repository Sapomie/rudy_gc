package moviex

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CMovieAlbumModel = (*customCMovieAlbumModel)(nil)

type (
	CMovieAlbumModel interface {
		cMovieAlbumModel
		ListAll(ctx context.Context) ([]*CMovieAlbum, error)
	}

	customCMovieAlbumModel struct {
		*defaultCMovieAlbumModel
	}
)

func NewCMovieAlbumModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CMovieAlbumModel {
	return &customCMovieAlbumModel{
		defaultCMovieAlbumModel: newCMovieAlbumModel(conn, c, opts...),
	}
}

func (m *customCMovieAlbumModel) ListAll(ctx context.Context) ([]*CMovieAlbum, error) {
	query := "SELECT " + cMovieAlbumRows + " FROM " + strings.Trim(m.table, "`") + " ORDER BY `id` ASC"
	var rows []*CMovieAlbum
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CMovieAlbum{}, nil
		}
		return nil, err
	}
	return rows, nil
}
