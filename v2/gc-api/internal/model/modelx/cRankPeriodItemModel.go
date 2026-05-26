package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CRankPeriodItemModel = (*customCRankPeriodItemModel)(nil)

type (
	// CRankPeriodItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCRankPeriodItemModel.
	CRankPeriodItemModel interface {
		cRankPeriodItemModel
	}

	customCRankPeriodItemModel struct {
		*defaultCRankPeriodItemModel
	}
)

// NewCRankPeriodItemModel returns a model for the database table.
func NewCRankPeriodItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CRankPeriodItemModel {
	return &customCRankPeriodItemModel{
		defaultCRankPeriodItemModel: newCRankPeriodItemModel(conn, c, opts...),
	}
}
