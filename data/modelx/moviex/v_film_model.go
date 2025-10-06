package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VFilmModel = (*customVFilmModel)(nil)

type (
	// VFilmModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVFilmModel.
	VFilmModel interface {
		vFilmModel
	}

	customVFilmModel struct {
		*defaultVFilmModel
	}
)

// NewVFilmModel returns a model for the database table.
func NewVFilmModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VFilmModel {
	return &customVFilmModel{
		defaultVFilmModel: newVFilmModel(conn, c, opts...),
	}
}
