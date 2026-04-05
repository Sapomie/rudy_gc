package moviex

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"

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
		ListDistinctReleaseMonths(ctx context.Context) ([]int64, error)
		ListDistinctReleaseDaysByRange(ctx context.Context, startUnix, endUnix int64) ([]int64, error)
		CalcReleaseBucket(ctx context.Context, startUnix, endUnix int64) (*MovieReleaseBucketCalc, error)
		CalcTopCastsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)
		CalcTopDirectorsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)
		CalcTopLabelsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)
		CalcTopPrefixesByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error)

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
	const query = `
SELECT DISTINCT
	CAST(UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(releasing_date), '%Y-%m-01 00:00:00')) AS SIGNED) AS bucket_month
FROM a_movie
WHERE releasing_date > 0
ORDER BY bucket_month ASC
`
	type row struct {
		BucketMonth int64 `db:"bucket_month"`
	}

	var rows []*row
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
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
	const query = `
SELECT DISTINCT
	CAST(releasing_date AS SIGNED) AS bucket_day
FROM a_movie
WHERE releasing_date >= ? AND releasing_date <= ?
ORDER BY bucket_day ASC
`
	type row struct {
		BucketDay int64 `db:"bucket_day"`
	}

	var rows []*row
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, startUnix, endUnix); err != nil {
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
	const query = `
SELECT
	COUNT(*) AS count_all,
	SUM(CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN 1 ELSE 0 END) AS count_owned,
	SUM(CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN COALESCE(wm.size, 0) ELSE 0 END) AS size_bytes,
	MAX(COALESCE(am.releasing_date, 0)) AS latest_releasing_date
FROM a_movie am
LEFT JOIN w_media wm ON wm.movie_jav_id = am.jav_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
`
	var out MovieReleaseBucketCalc
	if err := m.QueryRowNoCacheCtx(ctx, &out, query, consts.FilmIsNotRemoved, consts.FilmIsNotRemoved, startUnix, endUnix); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &MovieReleaseBucketCalc{}, nil
		}
		return nil, err
	}
	return &out, nil
}

func (m *customAMovieModel) CalcTopCastsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	const query = `
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
LEFT JOIN w_media wm ON wm.movie_jav_id = am.jav_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(c.name, '')) <> ''
GROUP BY agg_key
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit)
}

func (m *customAMovieModel) CalcTopDirectorsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	const query = `
SELECT
	CONCAT('d:', CAST(d.id AS CHAR)) AS agg_key,
	d.id AS agg_id,
	d.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN am.jav_id ELSE NULL END) AS count_owned
FROM a_movie am
JOIN am_director d ON d.id = am.director_id
LEFT JOIN w_media wm ON wm.movie_jav_id = am.jav_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(d.name, '')) <> ''
GROUP BY d.id, d.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit)
}

func (m *customAMovieModel) CalcTopLabelsByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	const query = `
SELECT
	CONCAT('l:', CAST(l.id AS CHAR)) AS agg_key,
	l.id AS agg_id,
	l.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN am.jav_id ELSE NULL END) AS count_owned
FROM a_movie am
JOIN am_label l ON l.id = am.label_id
LEFT JOIN w_media wm ON wm.movie_jav_id = am.jav_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(l.name, '')) <> ''
GROUP BY l.id, l.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit)
}

func (m *customAMovieModel) CalcTopPrefixesByReleaseRange(ctx context.Context, startUnix, endUnix int64, limit int64) ([]*MovieReleaseTopCalc, error) {
	const query = `
SELECT
	CONCAT('pfx:', CAST(pf.id AS CHAR)) AS agg_key,
	pf.id AS agg_id,
	pf.name AS agg_name,
	COUNT(DISTINCT am.jav_id) AS count_all,
	COUNT(DISTINCT CASE WHEN COALESCE(wm.is_removed, 0) = ? THEN am.jav_id ELSE NULL END) AS count_owned
FROM a_movie am
JOIN am_prefix pf ON pf.id = am.prefix_id
LEFT JOIN w_media wm ON wm.movie_jav_id = am.jav_id
WHERE am.releasing_date >= ? AND am.releasing_date <= ?
	AND TRIM(COALESCE(pf.name, '')) <> ''
GROUP BY pf.id, pf.name
ORDER BY count_all DESC, count_owned DESC, agg_name ASC
LIMIT ?
`
	return m.queryReleaseTopRows(ctx, query, startUnix, endUnix, limit)
}

func (m *customAMovieModel) queryReleaseTopRows(ctx context.Context, query string, startUnix, endUnix, limit int64) ([]*MovieReleaseTopCalc, error) {
	if limit <= 0 {
		limit = 30
	}
	var rows []*MovieReleaseTopCalc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.FilmIsNotRemoved, startUnix, endUnix, limit); err != nil {
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
