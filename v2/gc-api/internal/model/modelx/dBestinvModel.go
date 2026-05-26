package modelx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DBestinvModel = (*customDBestinvModel)(nil)

type (
	// DBestinvModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDBestinvModel.
	DBestinvModel interface {
		dBestinvModel
	}

	customDBestinvModel struct {
		*defaultDBestinvModel
	}
)

// NewDBestinvModel returns a model for the database table.
func NewDBestinvModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) DBestinvModel {
	return &customDBestinvModel{
		defaultDBestinvModel: newDBestinvModel(conn, c, opts...),
	}
}
