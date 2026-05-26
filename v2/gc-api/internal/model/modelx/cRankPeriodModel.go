package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CRankPeriodModel = (*customCRankPeriodModel)(nil)

type (
	// CRankPeriodModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCRankPeriodModel.
	CRankPeriodModel interface {
		cRankPeriodModel
	}

	customCRankPeriodModel struct {
		*defaultCRankPeriodModel
	}
)

// NewCRankPeriodModel returns a model for the database table.
func NewCRankPeriodModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CRankPeriodModel {
	return &customCRankPeriodModel{
		defaultCRankPeriodModel: newCRankPeriodModel(conn, c, opts...),
	}
}
