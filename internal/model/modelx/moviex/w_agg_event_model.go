package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WAggEventModel = (*customWAggEventModel)(nil)

type (
	WAggEventModel interface {
		wAggEventModel
		TableName() string
	}

	customWAggEventModel struct {
		*defaultWAggEventModel
	}
)

func NewWAggEventModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WAggEventModel {
	return &customWAggEventModel{
		defaultWAggEventModel: newWAggEventModel(conn, c, opts...),
	}
}

func (m *customWAggEventModel) TableName() string {
	return m.table
}
