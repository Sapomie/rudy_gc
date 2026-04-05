package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaBirthBucketStatModel = (*customWMediaBirthBucketStatModel)(nil)

type (
	WMediaBirthBucketStatListFilter struct {
		Level        string
		ScopeKeyLike string
		Year         *int64
		Quarter      *int64
		Month        *int64
		Sort         string
		Dir          string
		Page         int64
		PageSize     int64
	}

	// WMediaBirthBucketStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaBirthBucketStatModel.
	WMediaBirthBucketStatModel interface {
		wMediaBirthBucketStatModel
		ListAll(ctx context.Context) ([]*WMediaBirthBucketStat, error)
		ListByLevel(ctx context.Context, level string, year, quarter, month int64, positiveOnly bool) ([]*WMediaBirthBucketStat, error)
		ListPage(ctx context.Context, filter WMediaBirthBucketStatListFilter) ([]*WMediaBirthBucketStat, int64, error)
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customWMediaBirthBucketStatModel struct {
		*defaultWMediaBirthBucketStatModel
	}
)

// NewWMediaBirthBucketStatModel returns a model for the database table.
func NewWMediaBirthBucketStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaBirthBucketStatModel {
	return &customWMediaBirthBucketStatModel{
		defaultWMediaBirthBucketStatModel: newWMediaBirthBucketStatModel(conn, c, opts...),
	}
}

func (m *customWMediaBirthBucketStatModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaBirthBucketStatModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaBirthBucketStatModel) TableName() string {
	return m.table
}

func (m *customWMediaBirthBucketStatModel) ListAll(ctx context.Context) ([]*WMediaBirthBucketStat, error) {
	query, args, err := squirrel.
		Select(wMediaBirthBucketStatRows).
		From(m.table).
		OrderBy("`level` ASC", "`year` DESC", "`quarter` DESC", "`month` DESC", "`day` DESC", "`scope_key` DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMediaBirthBucketStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*WMediaBirthBucketStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaBirthBucketStatModel) ListByLevel(ctx context.Context, level string, year, quarter, month int64, positiveOnly bool) ([]*WMediaBirthBucketStat, error) {
	builder := squirrel.
		Select(wMediaBirthBucketStatRows).
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
		builder = builder.Where(squirrel.Gt{"media_count": 0})
	}

	orderBy := []string{"`year` DESC", "`quarter` DESC", "`month` DESC", "`day` DESC", "`scope_key` DESC"}
	query, args, err := builder.OrderBy(orderBy...).ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMediaBirthBucketStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*WMediaBirthBucketStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaBirthBucketStatModel) ListPage(ctx context.Context, filter WMediaBirthBucketStatListFilter) ([]*WMediaBirthBucketStat, int64, error) {
	builder := squirrel.
		Select(wMediaBirthBucketStatRows).
		From(m.table)
	countBuilder := squirrel.
		Select("COUNT(*)").
		From(m.table)

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

	orderBy := []string{"`updated_on` DESC", "`id` DESC"}
	dir := "DESC"
	idDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(filter.Dir), "asc") {
		dir = "ASC"
		idDir = "ASC"
	}
	switch strings.TrimSpace(filter.Sort) {
	case "media":
		orderBy = []string{"`media_count` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "removed":
		orderBy = []string{"`removed_count` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "size":
		orderBy = []string{"`size_bytes` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "subtitle":
		orderBy = []string{"`has_sub_count` " + dir, "`updated_on` DESC", "`id` " + idDir}
	case "latest_birth":
		orderBy = []string{"`latest_birth_time` " + dir, "`updated_on` DESC", "`id` " + idDir}
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
			return []*WMediaBirthBucketStat{}, 0, nil
		}
		return nil, 0, err
	}

	query, args, err := builder.OrderBy(orderBy...).ToSql()
	if err != nil {
		return nil, 0, err
	}
	var rows []*WMediaBirthBucketStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*WMediaBirthBucketStat{}, total, nil
		}
		return nil, 0, err
	}
	return rows, total, nil
}
