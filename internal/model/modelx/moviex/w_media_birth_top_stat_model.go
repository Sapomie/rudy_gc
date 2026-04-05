package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaBirthTopStatModel = (*customWMediaBirthTopStatModel)(nil)

type (
	// WMediaBirthTopStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaBirthTopStatModel.
	WMediaBirthTopStatModel interface {
		wMediaBirthTopStatModel
		DeleteByScopeAggType(ctx context.Context, scopeKey, aggType string) error
		ListAll(ctx context.Context) ([]*WMediaBirthTopStat, error)
		ListByScopeAggType(ctx context.Context, scopeKey, aggType string) ([]*WMediaBirthTopStat, error)
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customWMediaBirthTopStatModel struct {
		*defaultWMediaBirthTopStatModel
	}
)

// NewWMediaBirthTopStatModel returns a model for the database table.
func NewWMediaBirthTopStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaBirthTopStatModel {
	return &customWMediaBirthTopStatModel{
		defaultWMediaBirthTopStatModel: newWMediaBirthTopStatModel(conn, c, opts...),
	}
}

func (m *customWMediaBirthTopStatModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaBirthTopStatModel) TableName() string {
	return m.table
}

func (m *customWMediaBirthTopStatModel) ListAll(ctx context.Context) ([]*WMediaBirthTopStat, error) {
	query, args, err := squirrel.
		Select(wMediaBirthTopStatRows).
		From(m.table).
		OrderBy("`scope_key` ASC", "`agg_type` ASC", "`rank_no` ASC", "`agg_name` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMediaBirthTopStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*WMediaBirthTopStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaBirthTopStatModel) ListByScopeAggType(ctx context.Context, scopeKey, aggType string) ([]*WMediaBirthTopStat, error) {
	query, args, err := squirrel.
		Select(wMediaBirthTopStatRows).
		From(m.table).
		Where(squirrel.Eq{
			"scope_key": strings.TrimSpace(scopeKey),
			"agg_type":  strings.TrimSpace(aggType),
		}).
		OrderBy("`rank_no` ASC", "`agg_name` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMediaBirthTopStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*WMediaBirthTopStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaBirthTopStatModel) DeleteByScopeAggType(ctx context.Context, scopeKey, aggType string) error {
	rows, err := m.ListByScopeAggType(ctx, scopeKey, aggType)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if err := m.Delete(ctx, row.Id); err != nil {
			return err
		}
	}
	return nil
}
