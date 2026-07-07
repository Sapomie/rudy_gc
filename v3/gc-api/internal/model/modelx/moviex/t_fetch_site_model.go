package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TFetchSiteModel = (*customTFetchSiteModel)(nil)

type (
	// TFetchSiteModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTFetchSiteModel.
	TFetchSiteModel interface {
		tFetchSiteModel
		ListAll(ctx context.Context) ([]*TFetchSite, error)
	}

	customTFetchSiteModel struct {
		*defaultTFetchSiteModel
	}
)

// NewTFetchSiteModel returns a model for the database table.
func NewTFetchSiteModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TFetchSiteModel {
	return &customTFetchSiteModel{
		defaultTFetchSiteModel: newTFetchSiteModel(conn, c, opts...),
	}
}

func (m *customTFetchSiteModel) ListAll(ctx context.Context) ([]*TFetchSite, error) {
	query := "select " + tFetchSiteRows + " from " + m.table + " order by `site_code` asc"
	var rows []*TFetchSite
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TFetchSite{}, nil
		}
		return nil, err
	}
	return rows, nil
}
