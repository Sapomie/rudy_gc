package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DDetailModel = (*customDDetailModel)(nil)

type (
	// DDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDDetailModel.
	DDetailModel interface {
		dDetailModel
	}

	customDDetailModel struct {
		*defaultDDetailModel
	}
)

// NewDDetailModel returns a model for the database table.
func NewDDetailModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) DDetailModel {
	return &customDDetailModel{
		defaultDDetailModel: newDDetailModel(conn, c, opts...),
	}
}
