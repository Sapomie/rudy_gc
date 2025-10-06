package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CCafoModel = (*customCCafoModel)(nil)

type (
	// CCafoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCCafoModel.
	CCafoModel interface {
		cCafoModel
	}

	customCCafoModel struct {
		*defaultCCafoModel
	}
)

// NewCCafoModel returns a model for the database table.
func NewCCafoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CCafoModel {
	return &customCCafoModel{
		defaultCCafoModel: newCCafoModel(conn, c, opts...),
	}
}
