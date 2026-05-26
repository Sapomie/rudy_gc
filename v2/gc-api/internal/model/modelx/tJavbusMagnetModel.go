package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TJavbusMagnetModel = (*customTJavbusMagnetModel)(nil)

type (
	// TJavbusMagnetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTJavbusMagnetModel.
	TJavbusMagnetModel interface {
		tJavbusMagnetModel
	}

	customTJavbusMagnetModel struct {
		*defaultTJavbusMagnetModel
	}
)

// NewTJavbusMagnetModel returns a model for the database table.
func NewTJavbusMagnetModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TJavbusMagnetModel {
	return &customTJavbusMagnetModel{
		defaultTJavbusMagnetModel: newTJavbusMagnetModel(conn, c, opts...),
	}
}
