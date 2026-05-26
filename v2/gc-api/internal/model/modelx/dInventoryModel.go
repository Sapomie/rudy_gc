package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DInventoryModel = (*customDInventoryModel)(nil)

type (
	// DInventoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDInventoryModel.
	DInventoryModel interface {
		dInventoryModel
	}

	customDInventoryModel struct {
		*defaultDInventoryModel
	}
)

// NewDInventoryModel returns a model for the database table.
func NewDInventoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) DInventoryModel {
	return &customDInventoryModel{
		defaultDInventoryModel: newDInventoryModel(conn, c, opts...),
	}
}
