package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmCastModel = (*customAmCastModel)(nil)

type (
	// AmCastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmCastModel.
	AmCastModel interface {
		amCastModel
	}

	customAmCastModel struct {
		*defaultAmCastModel
	}
)

// NewAmCastModel returns a model for the database table.
func NewAmCastModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmCastModel {
	return &customAmCastModel{
		defaultAmCastModel: newAmCastModel(conn, c, opts...),
	}
}
