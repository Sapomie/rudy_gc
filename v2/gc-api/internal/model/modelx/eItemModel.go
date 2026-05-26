package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EItemModel = (*customEItemModel)(nil)

type (
	// EItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEItemModel.
	EItemModel interface {
		eItemModel
	}

	customEItemModel struct {
		*defaultEItemModel
	}
)

// NewEItemModel returns a model for the database table.
func NewEItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) EItemModel {
	return &customEItemModel{
		defaultEItemModel: newEItemModel(conn, c, opts...),
	}
}
