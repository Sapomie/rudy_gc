package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmrMovieGenreModel = (*customAmrMovieGenreModel)(nil)

type (
	// AmrMovieGenreModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmrMovieGenreModel.
	AmrMovieGenreModel interface {
		amrMovieGenreModel
	}

	customAmrMovieGenreModel struct {
		*defaultAmrMovieGenreModel
	}
)

// NewAmrMovieGenreModel returns a model for the database table.
func NewAmrMovieGenreModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmrMovieGenreModel {
	return &customAmrMovieGenreModel{
		defaultAmrMovieGenreModel: newAmrMovieGenreModel(conn, c, opts...),
	}
}
