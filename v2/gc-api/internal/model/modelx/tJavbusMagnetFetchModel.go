package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TJavbusMagnetFetchModel = (*customTJavbusMagnetFetchModel)(nil)

type (
	// TJavbusMagnetFetchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTJavbusMagnetFetchModel.
	TJavbusMagnetFetchModel interface {
		tJavbusMagnetFetchModel
	}

	customTJavbusMagnetFetchModel struct {
		*defaultTJavbusMagnetFetchModel
	}
)

// NewTJavbusMagnetFetchModel returns a model for the database table.
func NewTJavbusMagnetFetchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TJavbusMagnetFetchModel {
	return &customTJavbusMagnetFetchModel{
		defaultTJavbusMagnetFetchModel: newTJavbusMagnetFetchModel(conn, c, opts...),
	}
}
