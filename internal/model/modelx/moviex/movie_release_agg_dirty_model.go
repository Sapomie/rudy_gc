package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MovieReleaseAggDirtyModel = (*customMovieReleaseAggDirtyModel)(nil)

type (
	// MovieReleaseAggDirtyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMovieReleaseAggDirtyModel.
	MovieReleaseAggDirtyModel interface {
		movieReleaseAggDirtyModel
		ListAll(ctx context.Context, limit int) ([]*MovieReleaseAggDirty, error)
		TouchMonth(ctx context.Context, bucketMonth int64, scopeKey string, now int64) error
	}

	customMovieReleaseAggDirtyModel struct {
		*defaultMovieReleaseAggDirtyModel
	}
)

// NewMovieReleaseAggDirtyModel returns a model for the database table.
func NewMovieReleaseAggDirtyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MovieReleaseAggDirtyModel {
	return &customMovieReleaseAggDirtyModel{
		defaultMovieReleaseAggDirtyModel: newMovieReleaseAggDirtyModel(conn, c, opts...),
	}
}

func (m *customMovieReleaseAggDirtyModel) ListAll(ctx context.Context, limit int) ([]*MovieReleaseAggDirty, error) {
	query := "select " + movieReleaseAggDirtyRows + " from " + m.table + " order by `bucket_month` asc"
	args := make([]any, 0, 1)
	if limit > 0 {
		query += " limit ?"
		args = append(args, limit)
	}
	var rows []*MovieReleaseAggDirty
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*MovieReleaseAggDirty{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customMovieReleaseAggDirtyModel) TouchMonth(ctx context.Context, bucketMonth int64, scopeKey string, now int64) error {
	row, err := m.FindOneByBucketMonth(ctx, bucketMonth)
	if err == nil && row != nil {
		row.ScopeKey = scopeKey
		row.UpdatedOn = now
		return m.Update(ctx, row)
	}
	if err != nil && err != ErrNotFound {
		return err
	}
	_, err = m.Insert(ctx, &MovieReleaseAggDirty{
		BucketMonth: bucketMonth,
		ScopeKey:    scopeKey,
		CreatedOn:   now,
		UpdatedOn:   now,
	})
	return err
}
