package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TmAlbumItemModel = (*customTmAlbumItemModel)(nil)

type (
	// TmAlbumItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTmAlbumItemModel.
	TmAlbumItemModel interface {
		tmAlbumItemModel
	}

	customTmAlbumItemModel struct {
		*defaultTmAlbumItemModel
	}
)

// NewTmAlbumItemModel returns a model for the database table.
func NewTmAlbumItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TmAlbumItemModel {
	return &customTmAlbumItemModel{
		defaultTmAlbumItemModel: newTmAlbumItemModel(conn, c, opts...),
	}
}
