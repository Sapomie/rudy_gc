package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaBirthTopStatModel = (*customWMediaBirthTopStatModel)(nil)

type (
	// WMediaBirthTopStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaBirthTopStatModel.
	WMediaBirthTopStatModel interface {
		wMediaBirthTopStatModel
	}

	customWMediaBirthTopStatModel struct {
		*defaultWMediaBirthTopStatModel
	}
)

// NewWMediaBirthTopStatModel returns a model for the database table.
func NewWMediaBirthTopStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaBirthTopStatModel {
	return &customWMediaBirthTopStatModel{
		defaultWMediaBirthTopStatModel: newWMediaBirthTopStatModel(conn, c, opts...),
	}
}
