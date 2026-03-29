package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TSukebeiTorrentFetchModel = (*customTSukebeiTorrentFetchModel)(nil)

type (
	// TSukebeiTorrentFetchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTSukebeiTorrentFetchModel.
	TSukebeiTorrentFetchModel interface {
		tSukebeiTorrentFetchModel
		ListByFetchStatuses(ctx context.Context, statuses []int64, limit int64) ([]*TSukebeiTorrentFetch, error)
		ListRecent(ctx context.Context, offset int64, limit int64) ([]*TSukebeiTorrentFetch, error)
		CountPageRows(ctx context.Context, filter SukebeiFetchPageFilter) (int64, error)
		ListPageRows(ctx context.Context, offset int64, limit int64, orderBy string, filter SukebeiFetchPageFilter) ([]*TSukebeiTorrentFetch, error)
	}

	customTSukebeiTorrentFetchModel struct {
		*defaultTSukebeiTorrentFetchModel
	}
)

type SukebeiFetchPageFilter struct {
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

// NewTSukebeiTorrentFetchModel returns a model for the database table.
func NewTSukebeiTorrentFetchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TSukebeiTorrentFetchModel {
	return &customTSukebeiTorrentFetchModel{
		defaultTSukebeiTorrentFetchModel: newTSukebeiTorrentFetchModel(conn, c, opts...),
	}
}

func (m *customTSukebeiTorrentFetchModel) ListByFetchStatuses(ctx context.Context, statuses []int64, limit int64) ([]*TSukebeiTorrentFetch, error) {
	if len(statuses) == 0 {
		return []*TSukebeiTorrentFetch{}, nil
	}

	queryBuilder := squirrel.
		Select(tSukebeiTorrentFetchRows).
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

	var rows []*TSukebeiTorrentFetch
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TSukebeiTorrentFetch{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTSukebeiTorrentFetchModel) ListRecent(ctx context.Context, offset int64, limit int64) ([]*TSukebeiTorrentFetch, error) {
	queryBuilder := squirrel.
		Select(tSukebeiTorrentFetchRows).
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

	var rows []*TSukebeiTorrentFetch
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TSukebeiTorrentFetch{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTSukebeiTorrentFetchModel) CountPageRows(ctx context.Context, filter SukebeiFetchPageFilter) (int64, error) {
	queryBuilder := squirrel.Select("COUNT(1)").From(strings.Trim(m.table, "`"))
	queryBuilder = applySukebeiPageFilter(queryBuilder, filter)

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

func (m *customTSukebeiTorrentFetchModel) ListPageRows(ctx context.Context, offset int64, limit int64, orderBy string, filter SukebeiFetchPageFilter) ([]*TSukebeiTorrentFetch, error) {
	queryBuilder := squirrel.Select(tSukebeiTorrentFetchRows).From(strings.Trim(m.table, "`"))
	queryBuilder = applySukebeiPageFilter(queryBuilder, filter)
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

	var rows []*TSukebeiTorrentFetch
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TSukebeiTorrentFetch{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func applySukebeiPageFilter(queryBuilder squirrel.SelectBuilder, filter SukebeiFetchPageFilter) squirrel.SelectBuilder {
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
