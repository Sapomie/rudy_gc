package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MovieReleaseBucketStatModel = (*customMovieReleaseBucketStatModel)(nil)

type (
	MovieReleaseBucketStatListFilter struct {
		Level        string
		ScopeKeyLike string
		Year         *int64
		Quarter      *int64
		Month        *int64
		Day          *int64
		Sort         string
		Dir          string
		Page         int64
		PageSize     int64
	}

	// MovieReleaseBucketStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMovieReleaseBucketStatModel.
	MovieReleaseBucketStatModel interface {
		movieReleaseBucketStatModel
		ListAll(ctx context.Context) ([]*MovieReleaseBucketStat, error)
		ListByLevel(ctx context.Context, level string, year, quarter, month int64, positiveOnly bool) ([]*MovieReleaseBucketStat, error)
		ListPage(ctx context.Context, filter MovieReleaseBucketStatListFilter) ([]*MovieReleaseBucketStat, int64, error)
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customMovieReleaseBucketStatModel struct {
		*defaultMovieReleaseBucketStatModel
	}
)

// NewMovieReleaseBucketStatModel returns a model for the database table.
func NewMovieReleaseBucketStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MovieReleaseBucketStatModel {
	return &customMovieReleaseBucketStatModel{
		defaultMovieReleaseBucketStatModel: newMovieReleaseBucketStatModel(conn, c, opts...),
	}
}

func (m *customMovieReleaseBucketStatModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customMovieReleaseBucketStatModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customMovieReleaseBucketStatModel) TableName() string {
	return m.table
}

func (m *customMovieReleaseBucketStatModel) ListAll(ctx context.Context) ([]*MovieReleaseBucketStat, error) {
	query, args, err := squirrel.
		Select(movieReleaseBucketStatRows).
		From(m.table).
		OrderBy("`level` ASC", "`year` DESC", "`quarter` DESC", "`month` DESC", "`day` DESC", "`scope_key` DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*MovieReleaseBucketStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*MovieReleaseBucketStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customMovieReleaseBucketStatModel) ListByLevel(ctx context.Context, level string, year, quarter, month int64, positiveOnly bool) ([]*MovieReleaseBucketStat, error) {
	builder := squirrel.
		Select(movieReleaseBucketStatRows).
		From(m.table).
		Where(squirrel.Eq{"level": strings.TrimSpace(level)})
	if year > 0 {
		builder = builder.Where(squirrel.Eq{"year": year})
	}
	if quarter > 0 {
		builder = builder.Where(squirrel.Eq{"quarter": quarter})
	}
	if month > 0 {
		builder = builder.Where(squirrel.Eq{"month": month})
	}
	if positiveOnly {
		builder = builder.Where(squirrel.Gt{"count_all": 0})
	}
	query, args, err := builder.OrderBy("`year` DESC", "`quarter` DESC", "`month` DESC", "`day` DESC", "`scope_key` DESC").ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*MovieReleaseBucketStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*MovieReleaseBucketStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customMovieReleaseBucketStatModel) ListPage(ctx context.Context, filter MovieReleaseBucketStatListFilter) ([]*MovieReleaseBucketStat, int64, error) {
	builder := squirrel.Select(movieReleaseBucketStatRows).From(m.table)
	countBuilder := squirrel.Select("COUNT(*)").From(m.table)

	if level := strings.TrimSpace(filter.Level); level != "" {
		builder = builder.Where(squirrel.Eq{"level": level})
		countBuilder = countBuilder.Where(squirrel.Eq{"level": level})
	}
	if scopeKeyLike := strings.TrimSpace(filter.ScopeKeyLike); scopeKeyLike != "" {
		like := "%" + scopeKeyLike + "%"
		builder = builder.Where(squirrel.Like{"scope_key": like})
		countBuilder = countBuilder.Where(squirrel.Like{"scope_key": like})
	}
	if filter.Year != nil {
		builder = builder.Where(squirrel.Eq{"year": *filter.Year})
		countBuilder = countBuilder.Where(squirrel.Eq{"year": *filter.Year})
	}
	if filter.Quarter != nil {
		builder = builder.Where(squirrel.Eq{"quarter": *filter.Quarter})
		countBuilder = countBuilder.Where(squirrel.Eq{"quarter": *filter.Quarter})
	}
	if filter.Month != nil {
		builder = builder.Where(squirrel.Eq{"month": *filter.Month})
		countBuilder = countBuilder.Where(squirrel.Eq{"month": *filter.Month})
	}
	if filter.Day != nil {
		builder = builder.Where(squirrel.Eq{"day": *filter.Day})
		countBuilder = countBuilder.Where(squirrel.Eq{"day": *filter.Day})
	}

	orderBy := []string{"`updated_on` DESC", "`id` DESC"}
	dir := "DESC"
	idDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(filter.Dir), "asc") {
		dir = "ASC"
		idDir = "ASC"
	}
	switch strings.TrimSpace(filter.Sort) {
	case "all":
		orderBy = []string{"`count_all` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "owned":
		orderBy = []string{"`count_owned` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "size":
		orderBy = []string{"`size_bytes` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "latest_release":
		orderBy = []string{"`latest_releasing_date` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "scope":
		orderBy = []string{"`scope_key` " + dir, "`id` " + idDir}
	default:
		orderBy = []string{"`updated_on` " + dir, "`id` " + idDir}
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		builder = builder.Limit(uint64(pageSize)).Offset(uint64(offset))
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, countArgs...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*MovieReleaseBucketStat{}, 0, nil
		}
		return nil, 0, err
	}

	query, args, err := builder.OrderBy(orderBy...).ToSql()
	if err != nil {
		return nil, 0, err
	}
	var rows []*MovieReleaseBucketStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*MovieReleaseBucketStat{}, total, nil
		}
		return nil, 0, err
	}
	return rows, total, nil
}
