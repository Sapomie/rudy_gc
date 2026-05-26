package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CRankModel = (*customCRankModel)(nil)

type (
	// CRankModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCRankModel.
	CRankModel interface {
		cRankModel
	}

	customCRankModel struct {
		*defaultCRankModel
	}
)

// NewCRankModel returns a model for the database table.
func NewCRankModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CRankModel {
	return &customCRankModel{
		defaultCRankModel: newCRankModel(conn, c, opts...),
	}
}
