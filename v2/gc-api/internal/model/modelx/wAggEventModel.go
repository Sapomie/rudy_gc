package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WAggEventModel = (*customWAggEventModel)(nil)

type (
	// WAggEventModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWAggEventModel.
	WAggEventModel interface {
		wAggEventModel
	}

	customWAggEventModel struct {
		*defaultWAggEventModel
	}
)

// NewWAggEventModel returns a model for the database table.
func NewWAggEventModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WAggEventModel {
	return &customWAggEventModel{
		defaultWAggEventModel: newWAggEventModel(conn, c, opts...),
	}
}
