package moviex

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaModel = (*customWMediaModel)(nil)

type (
	WMediaBirthBucketCalc struct {
		MediaCount      int64 `db:"media_count"`
		RemovedCount    int64 `db:"removed_count"`
		SizeBytes       int64 `db:"size_bytes"`
		HasSubCount     int64 `db:"has_sub_count"`
		LatestBirthTime int64 `db:"latest_birth_time"`
	}

	WMediaBirthTopCalc struct {
		AggKey     string `db:"agg_key"`
		AggId      int64  `db:"agg_id"`
		AggName    string `db:"agg_name"`
		MediaCount int64  `db:"media_count"`
		SizeBytes  int64  `db:"size_bytes"`
	}

	// WMediaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaModel.
	WMediaModel interface {
		wMediaModel
		CalcBirthBucket(ctx context.Context, start, end int64) (*WMediaBirthBucketCalc, error)
		CalcTopCastsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		CalcTopDirectorsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		CalcTopLabelsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		CalcTopPrefixesByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*WMedia, error)
		ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*WMedia, total int64, err error)
		ListDistinctBirthDays(ctx context.Context) ([]int64, error)
		ListByFullDirPrefixes(ctx context.Context, prefixes []string) ([]*WMedia, error)
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customWMediaModel struct {
		*defaultWMediaModel
	}
)

// NewWMediaModel returns a model for the database table.
func NewWMediaModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaModel {
	return &customWMediaModel{
		defaultWMediaModel: newWMediaModel(conn, c, opts...),
	}
}

func (m *customWMediaModel) TableName() string { return m.table }

func (m *customWMediaModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaModel) ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*WMedia, error) {
	if len(movieJavIds) == 0 {
		return []*WMedia{}, nil
	}

	query, args, err := squirrel.
		Select(wMediaRows).
		From(m.table).
		Where(squirrel.Eq{"movie_jav_id": movieJavIds}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMedia
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMedia{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaModel) ListDistinctBirthDays(ctx context.Context) ([]int64, error) {
	const query = `
SELECT birth_time
FROM w_media
WHERE birth_time > 0
ORDER BY birth_time ASC`

	var birthTimes []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &birthTimes, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}

	seen := make(map[int64]struct{}, len(birthTimes))
	days := make([]int64, 0, len(birthTimes))
	for _, birthTime := range birthTimes {
		bucketDay := normalizeLocalBirthBucketDay(birthTime)
		if bucketDay <= 0 {
			continue
		}
		if _, ok := seen[bucketDay]; ok {
			continue
		}
		seen[bucketDay] = struct{}{}
		days = append(days, bucketDay)
	}
	return days, nil
}

func (m *customWMediaModel) CalcBirthBucket(ctx context.Context, start, end int64) (*WMediaBirthBucketCalc, error) {
	const query = `
SELECT
	COALESCE(SUM(CASE WHEN is_removed = ? THEN 1 ELSE 0 END), 0) AS media_count,
	COALESCE(SUM(CASE WHEN is_removed = ? THEN 1 ELSE 0 END), 0) AS removed_count,
	COALESCE(SUM(CASE WHEN is_removed = ? THEN size ELSE 0 END), 0) AS size_bytes,
	COALESCE(SUM(CASE WHEN is_removed = ? AND has_sub = ? THEN 1 ELSE 0 END), 0) AS has_sub_count,
	COALESCE(MAX(CASE WHEN is_removed = ? THEN birth_time ELSE 0 END), 0) AS latest_birth_time
FROM w_media
WHERE birth_time >= ? AND birth_time <= ?`

	var row WMediaBirthBucketCalc
	if err := m.QueryRowNoCacheCtx(
		ctx,
		&row,
		query,
		consts.FilmIsNotRemoved,
		consts.FilmIsRemoved,
		consts.FilmIsNotRemoved,
		consts.FilmIsNotRemoved,
		consts.FilmHasSub,
		consts.FilmIsNotRemoved,
		start,
		end,
	); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &WMediaBirthBucketCalc{}, nil
		}
		return nil, err
	}
	return &row, nil
}

func (m *customWMediaModel) CalcTopCastsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	if limit <= 0 {
		limit = 30
	}
	const query = `
SELECT
	src.agg_key AS agg_key,
	src.agg_id AS agg_id,
	src.agg_name AS agg_name,
	COUNT(*) AS media_count,
	COALESCE(SUM(src.size), 0) AS size_bytes
FROM (
	SELECT DISTINCT
		wm.movie_jav_id,
		wm.size,
		COALESCE(ac.person_id, 0) AS agg_id,
		CASE
			WHEN TRIM(COALESCE(cp.chinese, '')) <> '' THEN TRIM(cp.chinese)
			WHEN TRIM(COALESCE(cp.name, '')) <> '' THEN TRIM(cp.name)
			ELSE TRIM(ac.name)
		END AS agg_name,
		CONCAT(COALESCE(ac.person_id, 0), ':', TRIM(COALESCE(ac.name, ''))) AS agg_key
	FROM w_media wm
	INNER JOIN amr_movie_cast amr ON amr.movie_jav_id = wm.movie_jav_id
	INNER JOIN am_cast ac ON ac.id = amr.cast_id
	LEFT JOIN c_person cp ON cp.id = ac.person_id
	WHERE wm.is_removed = ?
	  AND wm.birth_time >= ?
	  AND wm.birth_time <= ?
) src
WHERE TRIM(COALESCE(src.agg_name, '')) <> ''
GROUP BY src.agg_key, src.agg_id, src.agg_name
ORDER BY media_count DESC, size_bytes DESC, agg_name ASC
LIMIT ?`

	var rows []*WMediaBirthTopCalc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.FilmIsNotRemoved, start, end, limit); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMediaBirthTopCalc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaModel) CalcTopDirectorsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	return m.calcTopByMovieField(ctx, start, end, limit, "director")
}

func (m *customWMediaModel) CalcTopLabelsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	return m.calcTopByMovieField(ctx, start, end, limit, "label")
}

func (m *customWMediaModel) CalcTopPrefixesByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	return m.calcTopByMovieField(ctx, start, end, limit, "prefix")
}

func (m *customWMediaModel) calcTopByMovieField(ctx context.Context, start, end int64, limit int, aggType string) ([]*WMediaBirthTopCalc, error) {
	if limit <= 0 {
		limit = 30
	}

	type joinInfo struct {
		joinTable string
		joinAlias string
		nameExpr  string
		idExpr    string
		joinOn    string
	}

	infoMap := map[string]joinInfo{
		"director": {
			joinTable: "am_director",
			joinAlias: "ad",
			nameExpr:  "TRIM(COALESCE(ad.name, ''))",
			idExpr:    "COALESCE(ad.id, 0)",
			joinOn:    "ad.id = am.director_id",
		},
		"label": {
			joinTable: "am_label",
			joinAlias: "al",
			nameExpr:  "TRIM(COALESCE(al.name, ''))",
			idExpr:    "COALESCE(al.id, 0)",
			joinOn:    "al.id = am.label_id",
		},
		"prefix": {
			joinTable: "am_prefix",
			joinAlias: "ap",
			nameExpr:  "TRIM(COALESCE(ap.name, ''))",
			idExpr:    "COALESCE(ap.id, 0)",
			joinOn:    "ap.id = am.prefix_id",
		},
	}

	info, ok := infoMap[aggType]
	if !ok {
		return []*WMediaBirthTopCalc{}, nil
	}

	query := `
SELECT
	CONCAT(` + info.idExpr + `, ':', ` + info.nameExpr + `) AS agg_key,
	` + info.idExpr + ` AS agg_id,
	` + info.nameExpr + ` AS agg_name,
	COUNT(*) AS media_count,
	COALESCE(SUM(wm.size), 0) AS size_bytes
FROM w_media wm
INNER JOIN a_movie am ON am.jav_id = wm.movie_jav_id
LEFT JOIN ` + info.joinTable + ` ` + info.joinAlias + ` ON ` + info.joinOn + `
WHERE wm.is_removed = ?
  AND wm.birth_time >= ?
  AND wm.birth_time <= ?
  AND ` + info.nameExpr + ` <> ''
GROUP BY agg_key, agg_id, agg_name
ORDER BY media_count DESC, size_bytes DESC, agg_name ASC
LIMIT ?`

	var rows []*WMediaBirthTopCalc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.FilmIsNotRemoved, start, end, limit); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMediaBirthTopCalc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func normalizeLocalBirthBucketDay(birthTime int64) int64 {
	if birthTime <= 0 {
		return 0
	}
	t := time.Unix(birthTime, 0).In(time.Local)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
}

func (m *customWMediaModel) ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*WMedia, total int64, err error) {
	if len(dirIDs) == 0 {
		return []*WMedia{}, []*WMedia{}, 0, nil
	}

	orderParts := splitOrder(mapWMediaOrderBy(orderBy))
	if len(orderParts) == 0 {
		orderParts = []string{"wm.birth_time DESC", "wm.movie_name DESC"}
	}

	qAll, argsAll, err := squirrel.
		Select("wm.*").
		From(m.table + " AS wm").
		LeftJoin("`v_film` vf ON vf.movie_jav_id = wm.movie_jav_id").
		LeftJoin("`g_sc_stat` gss ON gss.movie_jav_id = wm.movie_jav_id").
		Where(squirrel.Eq{"wm.is_removed": consts.FilmIsNotRemoved}).
		Where(squirrel.Eq{"wm.directory_id": dirIDs}).
		OrderBy(orderParts...).
		ToSql()
	if err != nil {
		return nil, nil, 0, err
	}

	var rows []*WMedia
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, qAll, argsAll...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMedia{}, []*WMedia{}, 0, nil
		}
		return nil, nil, 0, err
	}

	total = int64(len(rows))
	if total == 0 {
		return []*WMedia{}, []*WMedia{}, 0, nil
	}

	start := (page - 1) * size
	if start >= len(rows) {
		return rows, []*WMedia{}, total, nil
	}

	end := start + size
	if end > len(rows) {
		end = len(rows)
	}
	paged = rows[start:end]
	return rows, paged, total, nil
}

func (m *customWMediaModel) ListByFullDirPrefixes(ctx context.Context, prefixes []string) ([]*WMedia, error) {
	if len(prefixes) == 0 {
		return []*WMedia{}, nil
	}

	conditions := make(squirrel.Or, 0, len(prefixes)*2)
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		conditions = append(conditions,
			squirrel.Eq{"wm.full_dir": prefix},
			squirrel.Like{"wm.full_dir": prefix + "/%"},
		)
	}
	if len(conditions) == 0 {
		return []*WMedia{}, nil
	}

	query, args, err := squirrel.
		Select(wMediaRows).
		From(m.table + " AS wm").
		Where(conditions).
		OrderBy("wm.id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMedia
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMedia{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func mapWMediaOrderBy(orderBy string) string {
	order := "wm.birth_time DESC,wm.movie_name DESC"
	switch orderBy {
	case consts.OrderByBirthTime:
		order = "CASE WHEN vf.birth_time IS NULL THEN wm.birth_time ELSE vf.birth_time END DESC,wm.movie_name DESC"
	case consts.OrderByMediaBirthTime:
		order = "wm.birth_time DESC,wm.movie_name DESC"
	case consts.OrderByScTimes:
		order = "COALESCE(gss.sc_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByComeTimes:
		order = "COALESCE(gss.come_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByLastScTime:
		order = "COALESCE(gss.last_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByReleasingDate:
		order = "wm.releasing_date DESC,wm.movie_name DESC"
	}
	return order
}
