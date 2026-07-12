package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
)

var _ GScModel = (*customGScModel)(nil)

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
		ListPageByScTimeRange(ctx context.Context, offset, limit int64, orderBy string, startTime, endTime int64) ([]*GSc, error)
		ListPageByScTimeRangeWithMovieStats(ctx context.Context, offset, limit int64, sortField, sortOrder string, startTime, endTime int64) ([]*GSc, error)
		CountAll(ctx context.Context) (int64, error)
		CountByScTimeRange(ctx context.Context, startTime, endTime int64) (int64, error)
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
	return m.ListPageByScTimeRange(ctx, offset, limit, orderBy, 0, 0)
}

func (m *customGScModel) ListPageByScTimeRange(ctx context.Context, offset, limit int64, orderBy string, startTime, endTime int64) ([]*GSc, error) {
	if orderBy == "" {
		orderBy = "sc_time DESC"
	}

	query := squirrel.
		Select(gScRows).
		From(m.table)
	if startTime > 0 {
		query = query.Where(squirrel.GtOrEq{"sc_time": startTime})
	}
	if endTime > 0 {
		query = query.Where(squirrel.Lt{"sc_time": endTime})
	}

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

func (m *customGScModel) ListPageByScTimeRangeWithMovieStats(ctx context.Context, offset, limit int64, sortField, sortOrder string, startTime, endTime int64) ([]*GSc, error) {
	direction := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		direction = "ASC"
	}

	column := "COALESCE(sc_stats.sc_movie_count, 0)"
	if strings.EqualFold(strings.TrimSpace(sortField), "total_movie_count") {
		column = "COALESCE(sc_stats.total_movie_count, 0)"
	}

	statsSQL, statsArgs, err := squirrel.
		Select(
			"sc_name",
			"COUNT(*) AS total_movie_count",
			"SUM(CASE WHEN is_sc = ? THEN 1 ELSE 0 END) AS sc_movie_count",
		).
		From("`g_list`").
		GroupBy("sc_name").
		ToSql()
	if err != nil {
		return nil, err
	}
	statsArgs = append([]any{consts.GListIsSc}, statsArgs...)

	query := squirrel.
		Select("g_sc.*").
		From(m.table + " AS g_sc").
		LeftJoin("(" + statsSQL + ") AS sc_stats ON sc_stats.sc_name = g_sc.name")
	if startTime > 0 {
		query = query.Where(squirrel.GtOrEq{"g_sc.sc_time": startTime})
	}
	if endTime > 0 {
		query = query.Where(squirrel.Lt{"g_sc.sc_time": endTime})
	}

	sqlStr, args, err := query.
		OrderBy(column+" "+direction, "g_sc.sc_time DESC", "g_sc.id DESC").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}
	args = append(statsArgs, args...)

	var rows []*GSc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customGScModel) CountAll(ctx context.Context) (int64, error) {
	return m.CountByScTimeRange(ctx, 0, 0)
}

func (m *customGScModel) CountByScTimeRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	query := squirrel.
		Select("COUNT(*)").
		From(m.table)
	if startTime > 0 {
		query = query.Where(squirrel.GtOrEq{"sc_time": startTime})
	}
	if endTime > 0 {
		query = query.Where(squirrel.Lt{"sc_time": endTime})
	}

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
