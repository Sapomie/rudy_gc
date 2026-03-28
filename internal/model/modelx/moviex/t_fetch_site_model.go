package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TFetchSiteModel = (*customTFetchSiteModel)(nil)

type (
	// TFetchSiteModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTFetchSiteModel.
	TFetchSiteModel interface {
		tFetchSiteModel
	}

	customTFetchSiteModel struct {
		*defaultTFetchSiteModel
	}
)

// NewTFetchSiteModel returns a model for the database table.
func NewTFetchSiteModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TFetchSiteModel {
	return &customTFetchSiteModel{
		defaultTFetchSiteModel: newTFetchSiteModel(conn, c, opts...),
	}
}
