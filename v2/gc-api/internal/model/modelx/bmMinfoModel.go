package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BmMinfoModel = (*customBmMinfoModel)(nil)

type (
	// BmMinfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmMinfoModel.
	BmMinfoModel interface {
		bmMinfoModel
	}

	customBmMinfoModel struct {
		*defaultBmMinfoModel
	}
)

// NewBmMinfoModel returns a model for the database table.
func NewBmMinfoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmMinfoModel {
	return &customBmMinfoModel{
		defaultBmMinfoModel: newBmMinfoModel(conn, c, opts...),
	}
}
