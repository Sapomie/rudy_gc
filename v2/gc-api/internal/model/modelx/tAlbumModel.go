package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAlbumModel = (*customTAlbumModel)(nil)

type (
	// TAlbumModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAlbumModel.
	TAlbumModel interface {
		tAlbumModel
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
