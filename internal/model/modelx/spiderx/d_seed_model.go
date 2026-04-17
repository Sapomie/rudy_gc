package spiderx

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DSeedModel = (*customDSeedModel)(nil)

type (
	// DSeedModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDSeedModel.
	DSeedModel interface {
		dSeedModel
		withSession(session sqlx.Session) DSeedModel
		FindQueriesActive(ctx context.Context, nameType int64) ([]*DSeed, error)
		ListPage(ctx context.Context, offset, limit int64, orderBy string, filter DSeedListFilter) ([]*DSeed, error)
		CountPage(ctx context.Context, filter DSeedListFilter) (int64, error)
		CalcMovieStats(ctx context.Context, nameType int64, name string) (*DSeedMovieStats, error)
	}

	customDSeedModel struct {
		*defaultDSeedModel
	}

	DSeedListFilter struct {
		NameKeyword string
		Active      *int64
		SearchType  *int64
		NameType    *int64
		LastStatus  *int64
	}

	DSeedMovieStats struct {
		MovieTotal                     int64  `db:"movie_total"`
		MovieLatestReleasingMovieJavId string `db:"movie_latest_releasing_movie_jav_id"`
		MovieLatestReleasingMovieName  string `db:"movie_latest_releasing_movie_name"`
		MovieLatestReleasingDate       int64  `db:"movie_latest_releasing_date"`
	}
)

const (
	QueryInactive = 1 + iota
	QueryActive
)

const (
	QueryNamePrefix = 1 + iota
	QueryNameLabel
)

const (
	QueryByOffset = 1 + iota
	QueryByStartEnd
)

// NewDSeedModel returns a model for the database table.
func NewDSeedModel(conn sqlx.SqlConn) DSeedModel {
	return &customDSeedModel{
		defaultDSeedModel: newDSeedModel(conn),
	}
}

func (m *customDSeedModel) withSession(session sqlx.Session) DSeedModel {
	return NewDSeedModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDSeedModel) FindQueriesActive(ctx context.Context, nameType int64) ([]*DSeed, error) {
	b := squirrel.
		Select("*").
		From(m.table).
		Where("`active` = ?", QueryActive).
		Where("`name_type` = ?", nameType).
		OrderBy("`updated_on` DESC, `name` ASC").
		PlaceholderFormat(squirrel.Question)

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build SQL failed: %w", err)
	}

	var rows []*DSeed
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customDSeedModel) ListPage(ctx context.Context, offset, limit int64, orderBy string, filter DSeedListFilter) ([]*DSeed, error) {
	if limit <= 0 {
		limit = 50
	}

	builder := squirrel.
		Select("*").
		From(m.table).
		OrderBy(orderBy).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Question)
	builder = applyDSeedListFilter(builder, filter)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build d_seed list sql failed: %w", err)
	}

	var rows []*DSeed
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customDSeedModel) CountPage(ctx context.Context, filter DSeedListFilter) (int64, error) {
	builder := squirrel.
		Select("COUNT(*)").
		From(m.table).
		PlaceholderFormat(squirrel.Question)
	builder = applyDSeedListFilter(builder, filter)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build d_seed count sql failed: %w", err)
	}

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customDSeedModel) CalcMovieStats(ctx context.Context, nameType int64, name string) (*DSeedMovieStats, error) {
	switch nameType {
	case QueryNamePrefix:
		return m.calcMovieStatsByPrefix(ctx, name)
	case QueryNameLabel:
		return m.calcMovieStatsByLabelJavID(ctx, name)
	default:
		return &DSeedMovieStats{}, nil
	}
}

func (m *customDSeedModel) calcMovieStatsByPrefix(ctx context.Context, name string) (*DSeedMovieStats, error) {
	const query = `
SELECT
	COUNT(*) AS movie_total,
	COALESCE(MAX(am.releasing_date), 0) AS movie_latest_releasing_date,
	COALESCE((
		SELECT am2.jav_id
		FROM a_movie am2
		JOIN am_prefix pf2 ON pf2.id = am2.prefix_id
		WHERE pf2.name = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_prefix pf3 ON pf3.id = am3.prefix_id
			WHERE pf3.name = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_jav_id,
	COALESCE((
		SELECT am2.name
		FROM a_movie am2
		JOIN am_prefix pf2 ON pf2.id = am2.prefix_id
		WHERE pf2.name = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_prefix pf3 ON pf3.id = am3.prefix_id
			WHERE pf3.name = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_name
FROM a_movie am
JOIN am_prefix pf ON pf.id = am.prefix_id
WHERE pf.name = ?
`
	var out DSeedMovieStats
	if err := m.conn.QueryRowCtx(ctx, &out, query, name, name, name, name, name); err != nil {
		return nil, err
	}
	return &out, nil
}

func (m *customDSeedModel) calcMovieStatsByLabelJavID(ctx context.Context, labelJavID string) (*DSeedMovieStats, error) {
	const query = `
SELECT
	COUNT(*) AS movie_total,
	COALESCE(MAX(am.releasing_date), 0) AS movie_latest_releasing_date,
	COALESCE((
		SELECT am2.jav_id
		FROM a_movie am2
		JOIN am_label lb2 ON lb2.id = am2.label_id
		WHERE lb2.jav_id = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_label lb3 ON lb3.id = am3.label_id
			WHERE lb3.jav_id = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_jav_id,
	COALESCE((
		SELECT am2.name
		FROM a_movie am2
		JOIN am_label lb2 ON lb2.id = am2.label_id
		WHERE lb2.jav_id = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_label lb3 ON lb3.id = am3.label_id
			WHERE lb3.jav_id = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_name
FROM a_movie am
JOIN am_label lb ON lb.id = am.label_id
WHERE lb.jav_id = ?
`
	var out DSeedMovieStats
	if err := m.conn.QueryRowCtx(ctx, &out, query, labelJavID, labelJavID, labelJavID, labelJavID, labelJavID); err != nil {
		return nil, err
	}
	return &out, nil
}

func applyDSeedListFilter(builder squirrel.SelectBuilder, filter DSeedListFilter) squirrel.SelectBuilder {
	if keyword := strings.TrimSpace(filter.NameKeyword); keyword != "" {
		builder = builder.Where("`name` LIKE ?", "%"+keyword+"%")
	}
	if filter.Active != nil {
		builder = builder.Where("`active` = ?", *filter.Active)
	}
	if filter.SearchType != nil {
		builder = builder.Where("`search_type` = ?", *filter.SearchType)
	}
	if filter.NameType != nil {
		builder = builder.Where("`name_type` = ?", *filter.NameType)
	}
	if filter.LastStatus != nil {
		builder = builder.Where("`last_status` = ?", *filter.LastStatus)
	}
	return builder
}
