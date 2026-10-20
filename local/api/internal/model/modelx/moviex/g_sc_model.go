package moviex

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/types"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

var _ GScModel = (*customGScModel)(nil)

type scEventMovieSortStat struct {
	ScName          string `db:"sc_name"`
	TotalMovieCount int64  `db:"total_movie_count"`
	ScMovieCount    int64  `db:"sc_movie_count"`
	JavMovieCount   int64  `db:"jav_movie_count"`
	FC2MovieCount   int64  `db:"fc2_movie_count"`
}

type (
	// GScModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGScModel.
	GScModel interface {
		gScModel
		FindAll(ctx context.Context) ([]*GSc, error)
		ListByNames(ctx context.Context, names []string) ([]*GSc, error)
		ListByComeMovieName(ctx context.Context, movieName string) ([]*GSc, error)
		ListTopNByScTime(ctx context.Context, n uint64) ([]*GSc, error)
		FindNearest(ctx context.Context, t int64) (*GSc, error)
		ListPage(ctx context.Context, offset, limit int64, orderBy string) ([]*GSc, error)
		ListPageByFilter(ctx context.Context, offset, limit int64, orderBy string, filter types.ScEventListFilter) ([]*GSc, error)
		ListPageByFilterWithMovieStats(ctx context.Context, offset, limit int64, sortField, sortOrder string, filter types.ScEventListFilter) ([]*GSc, error)
		CountAll(ctx context.Context) (int64, error)
		CountByFilter(ctx context.Context, filter types.ScEventListFilter) (int64, error)
		Invalidate(ctx context.Context, rows ...*GSc) error
	}

	customGScModel struct {
		*defaultGScModel
	}
)

// NewGScModel returns a model for the database table.
func NewGScModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GScModel {
	return &customGScModel{
		defaultGScModel: newGScModel(conn, c, opts...),
	}
}

func (m *customGScModel) Invalidate(ctx context.Context, rows ...*GSc) error {
	keys := make([]string, 0, len(rows)*2)
	for _, row := range rows {
		if row == nil {
			continue
		}
		keys = append(keys,
			fmt.Sprintf("%s%d", cacheRudyGcGScIdPrefix, row.Id),
			fmt.Sprintf("%s%s", cacheRudyGcGScNamePrefix, row.Name),
		)
	}
	if len(keys) == 0 {
		return nil
	}
	return m.DelCacheCtx(ctx, keys...)
}

func (m *customGScModel) ListTopNByScTime(ctx context.Context, n uint64) ([]*GSc, error) {
	sqlStr, args, err := squirrel.
		Select(gScRows).
		From(m.table).
		OrderBy("sc_time DESC").
		Limit(n).
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGScModel) FindAll(ctx context.Context) ([]*GSc, error) {
	sqlStr, args, err := squirrel.
		Select(gScRows).
		From(m.table).
		OrderBy("sc_time DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGScModel) ListByNames(ctx context.Context, names []string) ([]*GSc, error) {
	if len(names) == 0 {
		return []*GSc{}, nil
	}

	sqlStr, args, err := squirrel.
		Select(gScRows).
		From(m.table).
		Where(squirrel.Eq{"name": names}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGScModel) ListByComeMovieName(ctx context.Context, movieName string) ([]*GSc, error) {
	if movieName == "" {
		return []*GSc{}, nil
	}

	sqlStr, args, err := squirrel.
		Select(gScRows).
		From(m.table).
		Where(squirrel.Eq{"come_movie_name": movieName}).
		OrderBy("sc_time DESC", "id DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGScModel) FindNearest(ctx context.Context, t int64) (*GSc, error) {
	sqlStr, args, err := squirrel.
		Select(gScRows).
		From(m.table).
		Where(squirrel.Lt{"sc_time": t}).
		OrderBy("sc_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row GSc
	if err := m.QueryRowNoCacheCtx(ctx, &row, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, sqlx.ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customGScModel) ListPage(ctx context.Context, offset, limit int64, orderBy string) ([]*GSc, error) {
	return m.ListPageByFilter(ctx, offset, limit, orderBy, types.ScEventListFilter{})
}

func (m *customGScModel) ListPageByFilter(ctx context.Context, offset, limit int64, orderBy string, filter types.ScEventListFilter) ([]*GSc, error) {
	if orderBy == "" {
		orderBy = "sc_time DESC"
	}

	query := squirrel.
		Select(gScRows).
		From(m.table)
	query = applyGScListFilter(query, filter)

	sqlStr, args, err := query.
		OrderBy(orderBy).
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGScModel) ListPageByFilterWithMovieStats(ctx context.Context, offset, limit int64, sortField, sortOrder string, filter types.ScEventListFilter) ([]*GSc, error) {
	query := squirrel.
		Select(gScRows).
		From(m.table)
	query = applyGScListFilter(query, filter)
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	if len(rows) == 0 {
		return rows, nil
	}
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			names = append(names, row.Name)
		}
	}
	statsQuery, statsArgs, err := squirrel.Select(
		"sc_name",
		"COUNT(*) AS total_movie_count",
		"SUM(CASE WHEN is_sc = ? THEN 1 ELSE 0 END) AS sc_movie_count",
		"SUM(CASE WHEN movie_type = ? THEN 1 ELSE 0 END) AS jav_movie_count",
		"SUM(CASE WHEN movie_type = ? THEN 1 ELSE 0 END) AS fc2_movie_count",
	).From("`g_list`").Where(squirrel.Eq{"sc_name": names}).GroupBy("sc_name").ToSql()
	if err != nil {
		return nil, err
	}
	statsArgs = append([]any{
		consts.GListIsSc,
		moviedomain.SehuatangThreadMovieTypeJAV,
		moviedomain.SehuatangThreadMovieTypeFC2PPV,
	}, statsArgs...)
	var statRows []*scEventMovieSortStat
	if err := m.QueryRowsNoCacheCtx(ctx, &statRows, statsQuery, statsArgs...); err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, err
	}
	stats := make(map[string]*scEventMovieSortStat, len(statRows))
	for _, row := range statRows {
		if row != nil {
			stats[row.ScName] = row
		}
	}
	ascending := strings.EqualFold(strings.TrimSpace(sortOrder), "asc")
	normalizedSortField := strings.ToLower(strings.TrimSpace(sortField))
	sort.SliceStable(rows, func(i, j int) bool {
		left, right := stats[rows[i].Name], stats[rows[j].Name]
		leftValue := scEventMovieStatValue(left, normalizedSortField)
		rightValue := scEventMovieStatValue(right, normalizedSortField)
		if leftValue != rightValue {
			if ascending {
				return leftValue < rightValue
			}
			return leftValue > rightValue
		}
		if rows[i].ScTime != rows[j].ScTime {
			return rows[i].ScTime > rows[j].ScTime
		}
		return rows[i].Id > rows[j].Id
	})
	if offset < 0 {
		offset = 0
	}
	if offset >= int64(len(rows)) {
		return []*GSc{}, nil
	}
	end := int64(len(rows))
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return rows[int(offset):int(end)], nil
}

func scEventMovieStatValue(stat *scEventMovieSortStat, sortField string) int64 {
	if stat == nil {
		return 0
	}
	switch sortField {
	case "total_movie_count":
		return stat.TotalMovieCount
	case "jav_movie_count":
		return stat.JavMovieCount
	case "fc2_movie_count":
		return stat.FC2MovieCount
	default:
		return stat.ScMovieCount
	}
}

func (m *customGScModel) CountAll(ctx context.Context) (int64, error) {
	return m.CountByFilter(ctx, types.ScEventListFilter{})
}

func (m *customGScModel) CountByFilter(ctx context.Context, filter types.ScEventListFilter) (int64, error) {
	query := squirrel.
		Select("COUNT(*)").
		From(m.table)
	query = applyGScListFilter(query, filter)

	sqlStr, args, err := query.
		ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, sqlStr, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func applyGScListFilter(query squirrel.SelectBuilder, filter types.ScEventListFilter) squirrel.SelectBuilder {
	if filter.StartTime > 0 {
		query = query.Where(squirrel.GtOrEq{"sc_time": filter.StartTime})
	}
	if filter.EndTime > 0 {
		query = query.Where(squirrel.Lt{"sc_time": filter.EndTime})
	}
	if filter.KindOptionID > 0 {
		query = query.Where(squirrel.Eq{"kind_option_id": filter.KindOptionID})
	}
	if filter.FGOptionID > 0 {
		query = query.Where(squirrel.Eq{"fg_option_id": filter.FGOptionID})
	}
	if filter.VesselOptionID > 0 {
		query = query.Where(squirrel.Eq{"vessel_option_id": filter.VesselOptionID})
	}
	return query
}
