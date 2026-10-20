package moviex

import (
	"context"
	"errors"
	"rudy_gc/local/api/internal/consts"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AMovieModel = (*customAMovieModel)(nil)

type (
	// 继承 goctl 生成接口 + 扩展方法
	AMovieModel interface {
		aMovieModel
		ListPageWithTotal(ctx context.Context, offset, limit int64, orderKey string) ([]*AMovie, int64, error)
		CountAll(ctx context.Context) (int64, error)
		FindMoviesByName(ctx context.Context, name string) ([]*AMovie, error)
		FindMoviesByEncode(ctx context.Context, encode string) ([]*AMovie, error)
		FindMovieIdentities(ctx context.Context, javIDs []string, names []string) ([]*AMovieIdentity, error)
		ListDistinctReleaseMonths(ctx context.Context) ([]int64, error)
		ListDistinctReleaseMonthsByMode(ctx context.Context, aggMode string) ([]int64, error)
		ListDistinctReleaseDaysByRange(ctx context.Context, startUnix, endUnix int64) ([]int64, error)
		ListDistinctReleaseDaysByRangeMode(ctx context.Context, startUnix, endUnix int64, aggMode string) ([]int64, error)
		CalcReleaseBucket(ctx context.Context, startUnix, endUnix int64) (*MovieReleaseBucketCalc, error)
		CalcReleaseBucketByMode(ctx context.Context, startUnix, endUnix int64, aggMode string) (*MovieReleaseBucketCalc, error)
		CalcTopCastsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)
		CalcTopCastsByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error)
		CalcTopDirectorsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)
		CalcTopDirectorsByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error)
		CalcTopLabelsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)
		CalcTopLabelsByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error)
		CalcTopPrefixesByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)
		CalcTopPrefixesByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error)

		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customAMovieModel struct {
		*defaultAMovieModel
	}

	MovieReleaseBucketCalc struct {
		CountAll            int64 `db:"count_all"`
		CountOwned          int64 `db:"count_owned"`
		RemovedCount        int64 `db:"removed_count"`
		SizeBytes           int64 `db:"size_bytes"`
		LatestReleasingDate int64 `db:"latest_releasing_date"`
	}

	MovieReleaseTopCalc struct {
		AggKey     string `db:"agg_key"`
		AggID      int64  `db:"agg_id"`
		AggName    string `db:"agg_name"`
		CountAll   int64  `db:"count_all"`
		CountOwned int64  `db:"count_owned"`
	}

	AMovieIdentity struct {
		JavID string `db:"jav_id"`
		Name  string `db:"name"`
	}
)

func NewAMovieModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AMovieModel {
	return &customAMovieModel{
		defaultAMovieModel: newAMovieModel(conn, c, opts...),
	}
}

func (m *customAMovieModel) TableName() string {
	return m.table
}

func (m *customAMovieModel) CountAll(ctx context.Context) (int64, error) {
	q, args, err := squirrel.Select("COUNT(*)").From(m.table).ToSql()
	if err != nil {
		return 0, err
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, q, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customAMovieModel) ListDistinctReleaseMonths(ctx context.Context) ([]int64, error) {
	return m.ListDistinctReleaseMonthsByMode(ctx, "all")
}

func (m *customAMovieModel) ListDistinctReleaseMonthsByMode(ctx context.Context, aggMode string) ([]int64, error) {
	query := `
SELECT DISTINCT
	CAST(UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(releasing_date), '%Y-%m-01 00:00:00')) AS SIGNED) AS bucket_month
FROM a_movie
WHERE releasing_date > 0
ORDER BY bucket_month ASC
`
	if strings.TrimSpace(aggMode) == "owned" {
		query = `
SELECT DISTINCT
	CAST(UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(am.releasing_date), '%Y-%m-01 00:00:00')) AS SIGNED) AS bucket_month
FROM a_movie am
JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
 AND wm.is_removed IN (?, ?)
WHERE am.releasing_date > 0
ORDER BY bucket_month ASC
`
	}
	type row struct {
		BucketMonth int64 `db:"bucket_month"`
	}

	var rows []*row
	var err error
	if strings.TrimSpace(aggMode) == "owned" {
		err = m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.WMediaSourceNative, consts.FilmIsNotRemoved, consts.FilmIsRemoved)
	} else {
		err = m.QueryRowsNoCacheCtx(ctx, &rows, query)
	}
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}
	out := make([]int64, 0, len(rows))
	for _, item := range rows {
		if item == nil || item.BucketMonth <= 0 {
			continue
		}
		out = append(out, item.BucketMonth)
	}
	return out, nil
}

func (m *customAMovieModel) ListDistinctReleaseDaysByRange(ctx context.Context, startUnix, endUnix int64) ([]int64, error) {
	return m.ListDistinctReleaseDaysByRangeMode(ctx, startUnix, endUnix, "all")
}

func (m *customAMovieModel) ListDistinctReleaseDaysByRangeMode(ctx context.Context, startUnix, endUnix int64, aggMode string) ([]int64, error) {
	query := `
SELECT DISTINCT
	CAST(releasing_date AS SIGNED) AS bucket_day
FROM a_movie
WHERE releasing_date >= ? AND releasing_date <= ?
ORDER BY bucket_day ASC
`
	if strings.TrimSpace(aggMode) == "owned" {
		query = `
SELECT DISTINCT
	CAST(am.releasing_date AS SIGNED) AS bucket_day
FROM a_movie am
STRAIGHT_JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
 AND wm.is_removed IN (?, ?)
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
ORDER BY bucket_day ASC
`
	}
	type row struct {
		BucketDay int64 `db:"bucket_day"`
	}

	var rows []*row
	var err error
	if strings.TrimSpace(aggMode) == "owned" {
		err = m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.WMediaSourceNative, consts.FilmIsNotRemoved, consts.FilmIsRemoved, startUnix, endUnix)
	} else {
		err = m.QueryRowsNoCacheCtx(ctx, &rows, query, startUnix, endUnix)
	}
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}
	out := make([]int64, 0, len(rows))
	for _, item := range rows {
		if item == nil || item.BucketDay <= 0 {
			continue
		}
		out = append(out, item.BucketDay)
	}
	return out, nil
}

func (m *customAMovieModel) CalcReleaseBucket(ctx context.Context, startUnix, endUnix int64) (*MovieReleaseBucketCalc, error) {
	return m.CalcReleaseBucketByMode(ctx, startUnix, endUnix, "all")
}

func (m *customAMovieModel) CalcReleaseBucketByMode(ctx context.Context, startUnix, endUnix int64, aggMode string) (*MovieReleaseBucketCalc, error) {
	query := `
SELECT
	COUNT(*) AS count_all,
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN 1 ELSE 0 END), 0) AS count_owned,
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN 1 ELSE 0 END), 0) AS removed_count,
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN COALESCE(wm.size, 0) ELSE 0 END), 0) AS size_bytes,
	COALESCE(MAX(COALESCE(am.releasing_date, 0)), 0) AS latest_releasing_date
FROM a_movie am
LEFT JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
`
	args := []any{
		consts.FilmIsNotRemoved,
		consts.FilmIsRemoved,
		consts.FilmIsNotRemoved,
		consts.WMediaSourceNative,
		startUnix,
		endUnix,
	}
	if strings.TrimSpace(aggMode) == "owned" {
		query = `
SELECT
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN 1 ELSE 0 END), 0) AS count_all,
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN 1 ELSE 0 END), 0) AS count_owned,
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN 1 ELSE 0 END), 0) AS removed_count,
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN COALESCE(wm.size, 0) ELSE 0 END), 0) AS size_bytes,
	COALESCE(MAX(COALESCE(am.releasing_date, 0)), 0) AS latest_releasing_date
FROM a_movie am
STRAIGHT_JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
  AND wm.is_removed IN (?, ?)
`
		args = []any{
			consts.FilmIsNotRemoved,
			consts.FilmIsNotRemoved,
			consts.FilmIsRemoved,
			consts.FilmIsNotRemoved,
			consts.WMediaSourceNative,
			startUnix,
			endUnix,
			consts.FilmIsNotRemoved,
			consts.FilmIsRemoved,
		}
	}
	var out MovieReleaseBucketCalc
	if err := m.QueryRowNoCacheCtx(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &MovieReleaseBucketCalc{}, nil
		}
		return nil, err
	}
	return &out, nil
}

func (m *customAMovieModel) CalcTopCastsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	return m.CalcTopCastsByReleaseRangeMode(ctx, startUnix, endUnix, limit, "all")
}

func (m *customAMovieModel) CalcTopCastsByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error) {
	query := `
SELECT
	CASE
		WHEN COALESCE(c.person_id, 0) > 0 THEN CONCAT('p:', CAST(c.person_id AS CHAR))
		ELSE CONCAT('n:', LOWER(TRIM(c.name)))
	END AS agg_key,
	MAX(COALESCE(c.person_id, 0)) AS agg_id,
	COALESCE(MAX(NULLIF(p.chinese, '')), MAX(NULLIF(p.name, '')), MAX(c.name)) AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN am.jav_id ELSE NULL END) AS count_owned
FROM a_movie am
JOIN amr_movie_cast amc ON amc.movie_jav_id = am.jav_id
JOIN am_cast c ON c.id = amc.cast_id
LEFT JOIN c_person p ON p.id = c.person_id
LEFT JOIN ` + buildNativeWMediaJoin("w_media", "wm", "am.jav_id") + `
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(c.name, '')) <> ''
GROUP BY agg_key
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	if strings.TrimSpace(aggMode) == "owned" {
		query = `
SELECT
	CASE
		WHEN COALESCE(c.person_id, 0) > 0 THEN CONCAT('p:', CAST(c.person_id AS CHAR))
		ELSE CONCAT('n:', LOWER(TRIM(c.name)))
	END AS agg_key,
	MAX(COALESCE(c.person_id, 0)) AS agg_id,
	COALESCE(MAX(NULLIF(p.chinese, '')), MAX(NULLIF(p.name, '')), MAX(c.name)) AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT am.jav_id) AS count_owned
FROM a_movie am
JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
 AND wm.is_removed = ?
JOIN amr_movie_cast amc ON amc.movie_jav_id = am.jav_id
JOIN am_cast c ON c.id = amc.cast_id
LEFT JOIN c_person p ON p.id = c.person_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(c.name, '')) <> ''
GROUP BY agg_key
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	}
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit, aggMode)
}

func (m *customAMovieModel) CalcTopDirectorsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	return m.CalcTopDirectorsByReleaseRangeMode(ctx, startUnix, endUnix, limit, "all")
}

func (m *customAMovieModel) CalcTopDirectorsByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error) {
	query := `
SELECT
	CONCAT('d:', CAST(d.id AS CHAR)) AS agg_key,
	d.id AS agg_id,
	d.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN am.jav_id ELSE NULL END) AS count_owned
FROM a_movie am
JOIN am_director d ON d.id = am.director_id
LEFT JOIN ` + buildNativeWMediaJoin("w_media", "wm", "am.jav_id") + `
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(d.name, '')) <> ''
GROUP BY d.id, d.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	if strings.TrimSpace(aggMode) == "owned" {
		query = `
SELECT
	CONCAT('d:', CAST(d.id AS CHAR)) AS agg_key,
	d.id AS agg_id,
	d.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT am.jav_id) AS count_owned
FROM a_movie am
JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
 AND wm.is_removed = ?
JOIN am_director d ON d.id = am.director_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(d.name, '')) <> ''
GROUP BY d.id, d.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	}
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit, aggMode)
}

func (m *customAMovieModel) CalcTopLabelsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	return m.CalcTopLabelsByReleaseRangeMode(ctx, startUnix, endUnix, limit, "all")
}

func (m *customAMovieModel) CalcTopLabelsByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error) {
	query := `
SELECT
	CONCAT('l:', CAST(l.id AS CHAR)) AS agg_key,
	l.id AS agg_id,
	l.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN am.jav_id ELSE NULL END) AS count_owned
FROM a_movie am
JOIN am_label l ON l.id = am.label_id
LEFT JOIN ` + buildNativeWMediaJoin("w_media", "wm", "am.jav_id") + `
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(l.name, '')) <> ''
GROUP BY l.id, l.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	if strings.TrimSpace(aggMode) == "owned" {
		query = `
SELECT
	CONCAT('l:', CAST(l.id AS CHAR)) AS agg_key,
	l.id AS agg_id,
	l.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT am.jav_id) AS count_owned
FROM a_movie am
JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
 AND wm.is_removed = ?
JOIN am_label l ON l.id = am.label_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(l.name, '')) <> ''
GROUP BY l.id, l.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	}
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit, aggMode)
}

func (m *customAMovieModel) CalcTopPrefixesByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	return m.CalcTopPrefixesByReleaseRangeMode(ctx, startUnix, endUnix, limit, "all")
}

func (m *customAMovieModel) CalcTopPrefixesByReleaseRangeMode(ctx context.Context, startUnix, endUnix int64, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error) {
	query := `
SELECT
	CONCAT('pfx:', CAST(pf.id AS CHAR)) AS agg_key,
	pf.id AS agg_id,
	pf.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN am.jav_id ELSE NULL END) AS count_owned
FROM a_movie am
JOIN am_prefix pf ON pf.id = am.prefix_id
LEFT JOIN ` + buildNativeWMediaJoin("w_media", "wm", "am.jav_id") + `
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(pf.name, '')) <> ''
GROUP BY pf.id, pf.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	if strings.TrimSpace(aggMode) == "owned" {
		query = `
SELECT
	CONCAT('pfx:', CAST(pf.id AS CHAR)) AS agg_key,
	pf.id AS agg_id,
	pf.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT am.jav_id) AS count_owned
FROM a_movie am
JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = ?
 AND wm.is_removed = ?
JOIN am_prefix pf ON pf.id = am.prefix_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(pf.name, '')) <> ''
GROUP BY pf.id, pf.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	}
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit, aggMode)
}

func (m *customAMovieModel) queryReleaseTopRows(ctx context.Context, query string, startUnix, endUnix, limit int64, aggMode string) ([]*MovieReleaseTopCalc, error) {
	if limit <= 0 {
		limit = 30
	}
	var rows []*MovieReleaseTopCalc
	var err error
	if strings.TrimSpace(aggMode) == "owned" {
		err = m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.WMediaSourceNative, consts.FilmIsNotRemoved, startUnix, endUnix, limit)
	} else {
		err = m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.FilmIsNotRemoved, startUnix, endUnix, limit)
	}
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*MovieReleaseTopCalc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customAMovieModel) FindMoviesByName(ctx context.Context, name string) ([]*AMovie, error) {
	q, args, err := squirrel.
		Select(aMovieRows).
		From(m.table).
		Where("`name` = ?", name).
		ToSql()
	if err != nil {
		return nil, err
	}

	var list []*AMovie
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*AMovie{}, nil
		}
		return nil, err
	}
	return list, nil
}

func (m *customAMovieModel) FindMovieIdentities(ctx context.Context, javIDs []string, names []string) ([]*AMovieIdentity, error) {
	javIDs = uniqueNonEmptyMovieIdentityValues(javIDs)
	names = uniqueNonEmptyMovieIdentityValues(names)
	if len(javIDs) == 0 && len(names) == 0 {
		return []*AMovieIdentity{}, nil
	}

	where := squirrel.Or{}
	if len(javIDs) > 0 {
		where = append(where, squirrel.Eq{"jav_id": javIDs})
	}
	if len(names) > 0 {
		where = append(where, squirrel.Eq{"name": names})
	}
	query, args, err := squirrel.
		Select("jav_id", "name").
		From(m.table).
		Where(where).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows := make([]*AMovieIdentity, 0)
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*AMovieIdentity{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func uniqueNonEmptyMovieIdentityValues(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (m *customAMovieModel) FindMoviesByEncode(ctx context.Context, encode string) ([]*AMovie, error) {
	q, args, err := squirrel.
		Select(aMovieRows).
		From(m.table).
		Where("`encode_name` = ?", encode).
		ToSql()
	if err != nil {
		return nil, err
	}

	var list []*AMovie
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*AMovie{}, nil
		}
		return nil, err
	}
	return list, nil
}

func (m *customAMovieModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAMovieModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}
