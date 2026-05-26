package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GListModel = (*customGListModel)(nil)

type (
	// GListModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGListModel.
	GListModel interface {
		gListModel
	}

	customGListModel struct {
		*defaultGListModel
	}
)

// NewGListModel returns a model for the database table.
func NewGListModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GListModel {
	return &customGListModel{
		defaultGListModel: newGListModel(conn, c, opts...),
	}
}
