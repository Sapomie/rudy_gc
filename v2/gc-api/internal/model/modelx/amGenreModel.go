package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmGenreModel = (*customAmGenreModel)(nil)

type (
	// AmGenreModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmGenreModel.
	AmGenreModel interface {
		amGenreModel
	}

	customAmGenreModel struct {
		*defaultAmGenreModel
	}
)

// NewAmGenreModel returns a model for the database table.
func NewAmGenreModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmGenreModel {
	return &customAmGenreModel{
		defaultAmGenreModel: newAmGenreModel(conn, c, opts...),
	}
}
