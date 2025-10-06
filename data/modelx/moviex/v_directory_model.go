package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VDirectoryModel = (*customVDirectoryModel)(nil)

type (
	// VDirectoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVDirectoryModel.
	VDirectoryModel interface {
		vDirectoryModel
	}

	customVDirectoryModel struct {
		*defaultVDirectoryModel
	}
)

// NewVDirectoryModel returns a model for the database table.
func NewVDirectoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VDirectoryModel {
	return &customVDirectoryModel{
		defaultVDirectoryModel: newVDirectoryModel(conn, c, opts...),
	}
}
