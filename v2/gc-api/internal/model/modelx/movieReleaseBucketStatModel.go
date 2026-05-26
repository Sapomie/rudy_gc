package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MovieReleaseBucketStatModel = (*customMovieReleaseBucketStatModel)(nil)

type (
	// MovieReleaseBucketStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMovieReleaseBucketStatModel.
	MovieReleaseBucketStatModel interface {
		movieReleaseBucketStatModel
	}

	customMovieReleaseBucketStatModel struct {
		*defaultMovieReleaseBucketStatModel
	}
)

// NewMovieReleaseBucketStatModel returns a model for the database table.
func NewMovieReleaseBucketStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MovieReleaseBucketStatModel {
	return &customMovieReleaseBucketStatModel{
		defaultMovieReleaseBucketStatModel: newMovieReleaseBucketStatModel(conn, c, opts...),
	}
}
