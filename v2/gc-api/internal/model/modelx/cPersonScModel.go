package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CPersonScModel = (*customCPersonScModel)(nil)

type (
	// CPersonScModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCPersonScModel.
	CPersonScModel interface {
		cPersonScModel
	}

	customCPersonScModel struct {
		*defaultCPersonScModel
	}
)

// NewCPersonScModel returns a model for the database table.
func NewCPersonScModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CPersonScModel {
	return &customCPersonScModel{
		defaultCPersonScModel: newCPersonScModel(conn, c, opts...),
	}
}
