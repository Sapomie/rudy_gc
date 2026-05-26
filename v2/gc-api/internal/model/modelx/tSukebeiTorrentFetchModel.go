package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TSukebeiTorrentFetchModel = (*customTSukebeiTorrentFetchModel)(nil)

type (
	// TSukebeiTorrentFetchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTSukebeiTorrentFetchModel.
	TSukebeiTorrentFetchModel interface {
		tSukebeiTorrentFetchModel
	}

	customTSukebeiTorrentFetchModel struct {
		*defaultTSukebeiTorrentFetchModel
	}
)

// NewTSukebeiTorrentFetchModel returns a model for the database table.
func NewTSukebeiTorrentFetchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TSukebeiTorrentFetchModel {
	return &customTSukebeiTorrentFetchModel{
		defaultTSukebeiTorrentFetchModel: newTSukebeiTorrentFetchModel(conn, c, opts...),
	}
}
