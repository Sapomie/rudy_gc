package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MovieReleaseTopStatModel = (*customMovieReleaseTopStatModel)(nil)

type (
	// MovieReleaseTopStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMovieReleaseTopStatModel.
	MovieReleaseTopStatModel interface {
		movieReleaseTopStatModel
		DeleteByScopeAggType(ctx context.Context, scopeKey, aggType string) error
		ListAll(ctx context.Context) ([]*MovieReleaseTopStat, error)
		ListByScopeAggType(ctx context.Context, scopeKey, aggType string) ([]*MovieReleaseTopStat, error)
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customMovieReleaseTopStatModel struct {
		*defaultMovieReleaseTopStatModel
	}
)

// NewMovieReleaseTopStatModel returns a model for the database table.
func NewMovieReleaseTopStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MovieReleaseTopStatModel {
	return &customMovieReleaseTopStatModel{
		defaultMovieReleaseTopStatModel: newMovieReleaseTopStatModel(conn, c, opts...),
	}
}

func (m *customMovieReleaseTopStatModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customMovieReleaseTopStatModel) TableName() string {
	return m.table
}

func (m *customMovieReleaseTopStatModel) ListAll(ctx context.Context) ([]*MovieReleaseTopStat, error) {
	query, args, err := squirrel.
		Select(movieReleaseTopStatRows).
		From(m.table).
		OrderBy("`scope_key` ASC", "`agg_type` ASC", "`rank_no` ASC", "`agg_name` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*MovieReleaseTopStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*MovieReleaseTopStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customMovieReleaseTopStatModel) ListByScopeAggType(ctx context.Context, scopeKey, aggType string) ([]*MovieReleaseTopStat, error) {
	query, args, err := squirrel.
		Select(movieReleaseTopStatRows).
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
	var rows []*MovieReleaseTopStat
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*MovieReleaseTopStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customMovieReleaseTopStatModel) DeleteByScopeAggType(ctx context.Context, scopeKey, aggType string) error {
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
