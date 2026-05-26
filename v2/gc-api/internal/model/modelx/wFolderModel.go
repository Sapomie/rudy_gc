package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WFolderModel = (*customWFolderModel)(nil)

type (
	// WFolderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWFolderModel.
	WFolderModel interface {
		wFolderModel
	}

	customWFolderModel struct {
		*defaultWFolderModel
	}
)

// NewWFolderModel returns a model for the database table.
func NewWFolderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WFolderModel {
	return &customWFolderModel{
		defaultWFolderModel: newWFolderModel(conn, c, opts...),
	}
}
