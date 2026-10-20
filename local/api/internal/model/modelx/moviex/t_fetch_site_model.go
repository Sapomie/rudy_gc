package moviex

import (
	"context"
	"database/sql"
	"fmt"

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
		InsertIfMissing(ctx context.Context, data *TFetchSite) (sql.Result, error)
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

// InsertIfMissing 只插入缺失的站点，保留数据库中已有配置。
func (m *customTFetchSiteModel) InsertIfMissing(ctx context.Context, data *TFetchSite) (sql.Result, error) {
	key := fmt.Sprintf("%s%s", cacheRudyGcTFetchSiteSiteCodePrefix, data.SiteCode)
	return m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `id` = `id`", m.table, tFetchSiteRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.SiteCode, data.SiteName, data.BaseUrl, data.UserAgent, data.Cookie, data.Proxy, data.TimeoutSeconds, data.RequestSleepMs, data.RetrySleepMs, data.MaxRetryTimes, data.Status, data.CreatedOn, data.UpdatedOn)
	}, key)
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
