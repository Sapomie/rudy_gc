package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmDirectorModel = (*customAmDirectorModel)(nil)

type (
	// AmDirectorModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmDirectorModel.
	AmDirectorModel interface {
		amDirectorModel
	}

	customAmDirectorModel struct {
		*defaultAmDirectorModel
	}
)

// NewAmDirectorModel returns a model for the database table.
func NewAmDirectorModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmDirectorModel {
	return &customAmDirectorModel{
		defaultAmDirectorModel: newAmDirectorModel(conn, c, opts...),
	}
}
