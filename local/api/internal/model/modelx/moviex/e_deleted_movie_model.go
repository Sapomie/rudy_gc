package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EDeletedMovieModel = (*customEDeletedMovieModel)(nil)

type (
	// EDeletedMovieModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEDeletedMovieModel.
	EDeletedMovieModel interface {
		eDeletedMovieModel
	}

	customEDeletedMovieModel struct {
		*defaultEDeletedMovieModel
	}
)

// NewEDeletedMovieModel returns a model for the database table.
func NewEDeletedMovieModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) EDeletedMovieModel {
	return &customEDeletedMovieModel{
		defaultEDeletedMovieModel: newEDeletedMovieModel(conn, c, opts...),
	}
}
