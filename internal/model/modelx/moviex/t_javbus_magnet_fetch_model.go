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
		CountByFetchStatus(ctx context.Context, filter JavbusFetchPageFilter) (map[int64]int64, error)
	}

	customTJavbusMagnetFetchModel struct {
		*defaultTJavbusMagnetFetchModel
	}
)

type JavbusFetchPageFilter struct {
	Owned              int64
	MediaOwned         int64
	RequireVFilm       bool
	RequireWMedia      bool
	Keyword            string
	FetchStatuses      []int64
	HasFetchStatuses   bool
	LastFetchFrom      int64
	LastFetchTo        int64
	HasLastFetchFrom   bool
	HasLastFetchTo     bool
	ReleaseDateFrom    int64
	ReleaseDateTo      int64
	HasReleaseDateFrom bool
	HasReleaseDateTo   bool
	FilmBirthFrom      int64
	FilmBirthTo        int64
	HasFilmBirthFrom   bool
	HasFilmBirthTo     bool
	MediaBirthFrom     int64
	MediaBirthTo       int64
	HasMediaBirthFrom  bool
	HasMediaBirthTo    bool
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
	queryBuilder := squirrel.Select("COUNT(1)").From(strings.Trim(m.table, "`") + " jf")
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
	queryBuilder := squirrel.Select("jf.*").
		From(strings.Trim(m.table, "`") + " jf").
		LeftJoin(buildLegacyWMediaJoin("`w_media`", "vf", "jf.movie_jav_id")).
		LeftJoin(buildNativeWMediaJoin("`w_media`", "wm", "jf.movie_jav_id"))
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

func (m *customTJavbusMagnetFetchModel) CountByFetchStatus(ctx context.Context, filter JavbusFetchPageFilter) (map[int64]int64, error) {
	queryBuilder := squirrel.
		Select("jf.fetch_status", "COUNT(1) AS total").
		From(strings.Trim(m.table, "`") + " jf")
	queryBuilder = applyJavbusPageFilter(queryBuilder, filter)
	queryBuilder = queryBuilder.GroupBy("jf.fetch_status")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []struct {
		FetchStatus int64 `db:"fetch_status"`
		Total       int64 `db:"total"`
	}
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return map[int64]int64{}, nil
		}
		return nil, err
	}

	result := make(map[int64]int64, len(rows))
	for _, row := range rows {
		result[row.FetchStatus] = row.Total
	}
	return result, nil
}

func applyJavbusPageFilter(queryBuilder squirrel.SelectBuilder, filter JavbusFetchPageFilter) squirrel.SelectBuilder {
	queryBuilder = applyFetchSiteOwnedFilter(queryBuilder, "jf", filter.Owned, "legacy_w_media", "vf")
	queryBuilder = applyFetchSiteOwnedFilter(queryBuilder, "jf", filter.MediaOwned, "w_media", "wm")
	if filter.RequireVFilm {
		queryBuilder = queryBuilder.Where(buildLegacyWMediaExists("`w_media`", "vf_sort", "jf.movie_jav_id"))
	}
	if filter.RequireWMedia {
		queryBuilder = queryBuilder.Where(buildNativeWMediaExists("`w_media`", "wm_sort", "jf.movie_jav_id"))
	}

	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		queryBuilder = queryBuilder.Where("(jf.movie_name LIKE ? OR jf.movie_jav_id LIKE ?)", like, like)
	}
	if filter.HasFetchStatuses && len(filter.FetchStatuses) > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"fetch_status": filter.FetchStatuses})
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
	if filter.HasFilmBirthFrom {
		queryBuilder = queryBuilder.Where(buildLegacyWMediaExists("`w_media`", "vf_birth_from", "jf.movie_jav_id", "vf_birth_from.birth_time >= ?"), filter.FilmBirthFrom)
	}
	if filter.HasFilmBirthTo {
		queryBuilder = queryBuilder.Where(buildLegacyWMediaExists("`w_media`", "vf_birth_to", "jf.movie_jav_id", "vf_birth_to.birth_time <= ?"), filter.FilmBirthTo)
	}
	if filter.HasMediaBirthFrom {
		queryBuilder = queryBuilder.Where(buildNativeWMediaExists("`w_media`", "wm_birth_from", "jf.movie_jav_id", "wm_birth_from.birth_time >= ?"), filter.MediaBirthFrom)
	}
	if filter.HasMediaBirthTo {
		queryBuilder = queryBuilder.Where(buildNativeWMediaExists("`w_media`", "wm_birth_to", "jf.movie_jav_id", "wm_birth_to.birth_time <= ?"), filter.MediaBirthTo)
	}
	return queryBuilder
}
