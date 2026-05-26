package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmrMovieCastModel = (*customAmrMovieCastModel)(nil)

type (
	// AmrMovieCastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmrMovieCastModel.
	AmrMovieCastModel interface {
		amrMovieCastModel
	}

	customAmrMovieCastModel struct {
		*defaultAmrMovieCastModel
	}
)

// NewAmrMovieCastModel returns a model for the database table.
func NewAmrMovieCastModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmrMovieCastModel {
	return &customAmrMovieCastModel{
		defaultAmrMovieCastModel: newAmrMovieCastModel(conn, c, opts...),
	}
}
