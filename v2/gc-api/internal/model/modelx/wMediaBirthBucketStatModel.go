package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaBirthBucketStatModel = (*customWMediaBirthBucketStatModel)(nil)

type (
	// WMediaBirthBucketStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaBirthBucketStatModel.
	WMediaBirthBucketStatModel interface {
		wMediaBirthBucketStatModel
	}

	customWMediaBirthBucketStatModel struct {
		*defaultWMediaBirthBucketStatModel
	}
)

// NewWMediaBirthBucketStatModel returns a model for the database table.
func NewWMediaBirthBucketStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaBirthBucketStatModel {
	return &customWMediaBirthBucketStatModel{
		defaultWMediaBirthBucketStatModel: newWMediaBirthBucketStatModel(conn, c, opts...),
	}
}
