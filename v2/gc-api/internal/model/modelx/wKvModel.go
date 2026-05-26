package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WKvModel = (*customWKvModel)(nil)

type (
	// WKvModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWKvModel.
	WKvModel interface {
		wKvModel
	}

	customWKvModel struct {
		*defaultWKvModel
	}
)

// NewWKvModel returns a model for the database table.
func NewWKvModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WKvModel {
	return &customWKvModel{
		defaultWKvModel: newWKvModel(conn, c, opts...),
	}
}
