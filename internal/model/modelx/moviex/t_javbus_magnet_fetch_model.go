package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TJavbusMagnetFetchModel = (*customTJavbusMagnetFetchModel)(nil)

type (
	// TJavbusMagnetFetchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTJavbusMagnetFetchModel.
	TJavbusMagnetFetchModel interface {
		tJavbusMagnetFetchModel
		ListByFetchStatuses(ctx context.Context, statuses []int64, limit int64) ([]*TJavbusMagnetFetch, error)
		ListRecent(ctx context.Context, offset int64, limit int64) ([]*TJavbusMagnetFetch, error)
		CountPageRows(ctx context.Context, filter JavbusFetchPageFilter) (int64, error)
		ListPageRows(ctx context.Context, offset int64, limit int64, orderBy string, filter JavbusFetchPageFilter) ([]*TJavbusMagnetFetch, error)
	}

	customTJavbusMagnetFetchModel struct {
		*defaultTJavbusMagnetFetchModel
	}
)

type JavbusFetchPageFilter struct {
	Keyword            string
	FetchStatus        int64
	FetchStatusSet     bool
	HasErrorOnly       bool
	HasNoErrorOnly     bool
	LastFetchFrom      int64
	LastFetchTo        int64
	HasLastFetchFrom   bool
	HasLastFetchTo     bool
	ReleaseDateFrom    int64
	ReleaseDateTo      int64
	HasReleaseDateFrom bool
	HasReleaseDateTo   bool
}

// NewTJavbusMagnetFetchModel returns a model for the database table.
func NewTJavbusMagnetFetchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TJavbusMagnetFetchModel {
	return &customTJavbusMagnetFetchModel{
		defaultTJavbusMagnetFetchModel: newTJavbusMagnetFetchModel(conn, c, opts...),
	}
}

func (m *customTJavbusMagnetFetchModel) ListByFetchStatuses(ctx context.Context, statuses []int64, limit int64) ([]*TJavbusMagnetFetch, error) {
	if len(statuses) == 0 {
		return []*TJavbusMagnetFetch{}, nil
	}

	queryBuilder := squirrel.
		Select(tJavbusMagnetFetchRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"fetch_status": statuses}).
		OrderBy("`last_fetch_time` ASC", "`id` ASC")
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TJavbusMagnetFetch
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TJavbusMagnetFetch{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTJavbusMagnetFetchModel) ListRecent(ctx context.Context, offset int64, limit int64) ([]*TJavbusMagnetFetch, error) {
	queryBuilder := squirrel.
		Select(tJavbusMagnetFetchRows).
		From(strings.Trim(m.table, "`")).
		OrderBy("`updated_on` DESC", "`id` DESC")
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		queryBuilder = queryBuilder.Offset(uint64(offset))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TJavbusMagnetFetch
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TJavbusMagnetFetch{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTJavbusMagnetFetchModel) CountPageRows(ctx context.Context, filter JavbusFetchPageFilter) (int64, error) {
	queryBuilder := squirrel.Select("COUNT(1)").From(strings.Trim(m.table, "`"))
	queryBuilder = applyJavbusPageFilter(queryBuilder, filter)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.CachedConn.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func (m *customTJavbusMagnetFetchModel) ListPageRows(ctx context.Context, offset int64, limit int64, orderBy string, filter JavbusFetchPageFilter) ([]*TJavbusMagnetFetch, error) {
	queryBuilder := squirrel.Select(tJavbusMagnetFetchRows).From(strings.Trim(m.table, "`"))
	queryBuilder = applyJavbusPageFilter(queryBuilder, filter)
	if strings.TrimSpace(orderBy) != "" {
		queryBuilder = queryBuilder.OrderBy(orderBy)
	}
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		queryBuilder = queryBuilder.Offset(uint64(offset))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TJavbusMagnetFetch
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TJavbusMagnetFetch{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func applyJavbusPageFilter(queryBuilder squirrel.SelectBuilder, filter JavbusFetchPageFilter) squirrel.SelectBuilder {
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		queryBuilder = queryBuilder.Where(squirrel.Or{
			squirrel.Like{"movie_code": like},
			squirrel.Like{"movie_jav_id": like},
		})
	}
	if filter.FetchStatusSet {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"fetch_status": filter.FetchStatus})
	}
	if filter.HasErrorOnly {
		queryBuilder = queryBuilder.Where("TRIM(`last_error`) <> ''")
	}
	if filter.HasNoErrorOnly {
		queryBuilder = queryBuilder.Where("TRIM(`last_error`) = ''")
	}
	if filter.HasLastFetchFrom {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"last_fetch_time": filter.LastFetchFrom})
	}
	if filter.HasLastFetchTo {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"last_fetch_time": filter.LastFetchTo})
	}
	if filter.HasReleaseDateFrom {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"release_date": filter.ReleaseDateFrom})
	}
	if filter.HasReleaseDateTo {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"release_date": filter.ReleaseDateTo})
	}
	return queryBuilder
}
