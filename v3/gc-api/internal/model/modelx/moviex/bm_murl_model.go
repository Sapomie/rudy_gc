package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BmMurlModel = (*customBmMurlModel)(nil)

type (
	// BmMurlModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmMurlModel.
	BmMurlModel interface {
		bmMurlModel
	}

	customBmMurlModel struct {
		*defaultBmMurlModel
	}
)

// NewBmMurlModel returns a model for the database table.
func NewBmMurlModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmMurlModel {
	return &customBmMurlModel{
		defaultBmMurlModel: newBmMurlModel(conn, c, opts...),
	}
}
