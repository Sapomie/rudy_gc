package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CMovieAlbumItemModel = (*customCMovieAlbumItemModel)(nil)

type (
	// CMovieAlbumItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCMovieAlbumItemModel.
	CMovieAlbumItemModel interface {
		cMovieAlbumItemModel
	}

	customCMovieAlbumItemModel struct {
		*defaultCMovieAlbumItemModel
	}
)

// NewCMovieAlbumItemModel returns a model for the database table.
func NewCMovieAlbumItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CMovieAlbumItemModel {
	return &customCMovieAlbumItemModel{
		defaultCMovieAlbumItemModel: newCMovieAlbumItemModel(conn, c, opts...),
	}
}
