package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CMovieAlbumModel = (*customCMovieAlbumModel)(nil)

type (
	// CMovieAlbumModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCMovieAlbumModel.
	CMovieAlbumModel interface {
		cMovieAlbumModel
	}

	customCMovieAlbumModel struct {
		*defaultCMovieAlbumModel
	}
)

// NewCMovieAlbumModel returns a model for the database table.
func NewCMovieAlbumModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CMovieAlbumModel {
	return &customCMovieAlbumModel{
		defaultCMovieAlbumModel: newCMovieAlbumModel(conn, c, opts...),
	}
}
