package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GScStatModel = (*customGScStatModel)(nil)

type (
	// GScStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGScStatModel.
	GScStatModel interface {
		gScStatModel
	}

	customGScStatModel struct {
		*defaultGScStatModel
	}
)

// NewGScStatModel returns a model for the database table.
func NewGScStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GScStatModel {
	return &customGScStatModel{
		defaultGScStatModel: newGScStatModel(conn, c, opts...),
	}
}
