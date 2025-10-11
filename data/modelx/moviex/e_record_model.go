package moviex

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ERecordModel = (*customERecordModel)(nil)

type (
	// ERecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customERecordModel.
	ERecordModel interface {
		eRecordModel
	}

	customERecordModel struct {
		*defaultERecordModel
	}
)

// NewERecordModel returns a model for the database table.
func NewERecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ERecordModel {
	return &customERecordModel{
		defaultERecordModel: newERecordModel(conn, c, opts...),
	}
}
