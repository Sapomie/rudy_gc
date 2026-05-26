package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CPersonModel = (*customCPersonModel)(nil)

type (
	// CPersonModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCPersonModel.
	CPersonModel interface {
		cPersonModel
	}

	customCPersonModel struct {
		*defaultCPersonModel
	}
)

// NewCPersonModel returns a model for the database table.
func NewCPersonModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CPersonModel {
	return &customCPersonModel{
		defaultCPersonModel: newCPersonModel(conn, c, opts...),
	}
}
