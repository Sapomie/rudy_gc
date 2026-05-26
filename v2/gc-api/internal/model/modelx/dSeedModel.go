package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DSeedModel = (*customDSeedModel)(nil)

type (
	// DSeedModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDSeedModel.
	DSeedModel interface {
		dSeedModel
	}

	customDSeedModel struct {
		*defaultDSeedModel
	}
)

// NewDSeedModel returns a model for the database table.
func NewDSeedModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) DSeedModel {
	return &customDSeedModel{
		defaultDSeedModel: newDSeedModel(conn, c, opts...),
	}
}
