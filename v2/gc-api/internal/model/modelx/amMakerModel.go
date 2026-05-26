package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmMakerModel = (*customAmMakerModel)(nil)

type (
	// AmMakerModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmMakerModel.
	AmMakerModel interface {
		amMakerModel
	}

	customAmMakerModel struct {
		*defaultAmMakerModel
	}
)

// NewAmMakerModel returns a model for the database table.
func NewAmMakerModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmMakerModel {
	return &customAmMakerModel{
		defaultAmMakerModel: newAmMakerModel(conn, c, opts...),
	}
}
