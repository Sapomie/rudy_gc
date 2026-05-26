package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TSehuatangMagnetModel = (*customTSehuatangMagnetModel)(nil)

type (
	// TSehuatangMagnetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTSehuatangMagnetModel.
	TSehuatangMagnetModel interface {
		tSehuatangMagnetModel
	}

	customTSehuatangMagnetModel struct {
		*defaultTSehuatangMagnetModel
	}
)

// NewTSehuatangMagnetModel returns a model for the database table.
func NewTSehuatangMagnetModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TSehuatangMagnetModel {
	return &customTSehuatangMagnetModel{
		defaultTSehuatangMagnetModel: newTSehuatangMagnetModel(conn, c, opts...),
	}
}
