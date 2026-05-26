package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmPrefixModel = (*customAmPrefixModel)(nil)

type (
	// AmPrefixModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmPrefixModel.
	AmPrefixModel interface {
		amPrefixModel
	}

	customAmPrefixModel struct {
		*defaultAmPrefixModel
	}
)

// NewAmPrefixModel returns a model for the database table.
func NewAmPrefixModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmPrefixModel {
	return &customAmPrefixModel{
		defaultAmPrefixModel: newAmPrefixModel(conn, c, opts...),
	}
}
