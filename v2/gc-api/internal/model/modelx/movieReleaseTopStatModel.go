package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MovieReleaseTopStatModel = (*customMovieReleaseTopStatModel)(nil)

type (
	// MovieReleaseTopStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMovieReleaseTopStatModel.
	MovieReleaseTopStatModel interface {
		movieReleaseTopStatModel
	}

	customMovieReleaseTopStatModel struct {
		*defaultMovieReleaseTopStatModel
	}
)

// NewMovieReleaseTopStatModel returns a model for the database table.
func NewMovieReleaseTopStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MovieReleaseTopStatModel {
	return &customMovieReleaseTopStatModel{
		defaultMovieReleaseTopStatModel: newMovieReleaseTopStatModel(conn, c, opts...),
	}
}
