package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmLabelModel = (*customAmLabelModel)(nil)

type (
	// AmLabelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmLabelModel.
	AmLabelModel interface {
		amLabelModel
	}

	customAmLabelModel struct {
		*defaultAmLabelModel
	}
)

// NewAmLabelModel returns a model for the database table.
func NewAmLabelModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmLabelModel {
	return &customAmLabelModel{
		defaultAmLabelModel: newAmLabelModel(conn, c, opts...),
	}
}
