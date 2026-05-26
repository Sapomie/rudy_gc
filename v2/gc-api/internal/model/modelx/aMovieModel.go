package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AMovieModel = (*customAMovieModel)(nil)

type (
	// AMovieModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAMovieModel.
	AMovieModel interface {
		aMovieModel
	}

	customAMovieModel struct {
		*defaultAMovieModel
	}
)

// NewAMovieModel returns a model for the database table.
func NewAMovieModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AMovieModel {
	return &customAMovieModel{
		defaultAMovieModel: newAMovieModel(conn, c, opts...),
	}
}
