package moviex

import (
	"context"
	"strings"

	"rudy_gc/internal/consts"

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
		CountByFetchStatus(ctx context.Context, filter SukebeiFetchPageFilter) (map[int64]int64, error)
	}

	customTSukebeiTorrentFetchModel struct {
		*defaultTSukebeiTorrentFetchModel
	}
)

type SukebeiFetchPageFilter struct {
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
	queryBuilder := squirrel.Select("COUNT(1)").From(strings.Trim(m.table, "`") + " sf")
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
	queryBuilder := squirrel.Select("sf.*").
		From(strings.Trim(m.table, "`") + " sf").
		LeftJoin("`v_film` vf ON vf.movie_jav_id = sf.movie_jav_id").
		LeftJoin("`w_media` wm ON wm.movie_jav_id = sf.movie_jav_id")
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

func (m *customTSukebeiTorrentFetchModel) CountByFetchStatus(ctx context.Context, filter SukebeiFetchPageFilter) (map[int64]int64, error) {
	queryBuilder := squirrel.
		Select("sf.fetch_status", "COUNT(1) AS total").
		From(strings.Trim(m.table, "`") + " sf")
	queryBuilder = applySukebeiPageFilter(queryBuilder, filter)
	queryBuilder = queryBuilder.GroupBy("sf.fetch_status")

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

func applySukebeiPageFilter(queryBuilder squirrel.SelectBuilder, filter SukebeiFetchPageFilter) squirrel.SelectBuilder {
	queryBuilder = applyFetchSiteOwnedFilter(queryBuilder, filter.Owned, "v_film", "vf")
	queryBuilder = applyFetchSiteOwnedFilter(queryBuilder, filter.MediaOwned, "w_media", "wm")
	if filter.RequireVFilm {
		queryBuilder = queryBuilder.Where("EXISTS (SELECT 1 FROM `v_film` vf_sort WHERE vf_sort.movie_jav_id = sf.movie_jav_id)")
	}
	if filter.RequireWMedia {
		queryBuilder = queryBuilder.Where("EXISTS (SELECT 1 FROM `w_media` wm_sort WHERE wm_sort.movie_jav_id = sf.movie_jav_id)")
	}

	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		queryBuilder = queryBuilder.Where(squirrel.Or{
			squirrel.Like{"movie_name": like},
			squirrel.Like{"movie_jav_id": like},
		})
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
		queryBuilder = queryBuilder.Where("EXISTS (SELECT 1 FROM `v_film` vf_birth_from WHERE vf_birth_from.movie_jav_id = sf.movie_jav_id AND vf_birth_from.birth_time >= ?)", filter.FilmBirthFrom)
	}
	if filter.HasFilmBirthTo {
		queryBuilder = queryBuilder.Where("EXISTS (SELECT 1 FROM `v_film` vf_birth_to WHERE vf_birth_to.movie_jav_id = sf.movie_jav_id AND vf_birth_to.birth_time <= ?)", filter.FilmBirthTo)
	}
	if filter.HasMediaBirthFrom {
		queryBuilder = queryBuilder.Where("EXISTS (SELECT 1 FROM `w_media` wm_birth_from WHERE wm_birth_from.movie_jav_id = sf.movie_jav_id AND wm_birth_from.birth_time >= ?)", filter.MediaBirthFrom)
	}
	if filter.HasMediaBirthTo {
		queryBuilder = queryBuilder.Where("EXISTS (SELECT 1 FROM `w_media` wm_birth_to WHERE wm_birth_to.movie_jav_id = sf.movie_jav_id AND wm_birth_to.birth_time <= ?)", filter.MediaBirthTo)
	}
	return queryBuilder
}

func applyFetchSiteOwnedFilter(queryBuilder squirrel.SelectBuilder, owned int64, tableName string, alias string) squirrel.SelectBuilder {
	switch owned {
	case 0, consts.MovieAll:
		return queryBuilder
	case consts.OwnedAll:
		return queryBuilder.Where("EXISTS (SELECT 1 FROM `" + tableName + "` " + alias + " WHERE " + alias + ".movie_jav_id = sf.movie_jav_id)")
	case consts.OwnedAllNotRemoved:
		return queryBuilder.Where("EXISTS (SELECT 1 FROM `"+tableName+"` "+alias+" WHERE "+alias+".movie_jav_id = sf.movie_jav_id AND "+alias+".is_removed = ?)", consts.FilmIsNotRemoved)
	case consts.OwnedHasSubNotRemoved:
		return queryBuilder.Where("EXISTS (SELECT 1 FROM `"+tableName+"` "+alias+" WHERE "+alias+".movie_jav_id = sf.movie_jav_id AND "+alias+".is_removed = ? AND "+alias+".has_sub = ?)", consts.FilmIsNotRemoved, consts.FilmHasSub)
	case consts.OwnedNoSubNotRemoved:
		return queryBuilder.Where("EXISTS (SELECT 1 FROM `"+tableName+"` "+alias+" WHERE "+alias+".movie_jav_id = sf.movie_jav_id AND "+alias+".is_removed = ? AND "+alias+".has_sub = ?)", consts.FilmIsNotRemoved, consts.FilmNoSub)
	case consts.OwnedRemoved:
		return queryBuilder.Where("EXISTS (SELECT 1 FROM `"+tableName+"` "+alias+" WHERE "+alias+".movie_jav_id = sf.movie_jav_id AND "+alias+".is_removed = ?)", consts.FilmIsRemoved)
	case consts.OwnedNotOwned:
		return queryBuilder.Where("NOT EXISTS (SELECT 1 FROM `" + tableName + "` " + alias + " WHERE " + alias + ".movie_jav_id = sf.movie_jav_id)")
	default:
		return queryBuilder
	}
}
