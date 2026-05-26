package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TSukebeiTorrentModel = (*customTSukebeiTorrentModel)(nil)

type (
	// TSukebeiTorrentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTSukebeiTorrentModel.
	TSukebeiTorrentModel interface {
		tSukebeiTorrentModel
	}

	customTSukebeiTorrentModel struct {
		*defaultTSukebeiTorrentModel
	}
)

// NewTSukebeiTorrentModel returns a model for the database table.
func NewTSukebeiTorrentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TSukebeiTorrentModel {
	return &customTSukebeiTorrentModel{
		defaultTSukebeiTorrentModel: newTSukebeiTorrentModel(conn, c, opts...),
	}
}
