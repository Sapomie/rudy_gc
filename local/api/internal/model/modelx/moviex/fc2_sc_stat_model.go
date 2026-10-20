package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FC2ScStatModel interface {
	fc2ScStatModel
	ListByMovieIDs(ctx context.Context, movieIDs []int64) ([]*FC2ScStat, error)
	ListAllMovieIDs(ctx context.Context) ([]int64, error)
}

func (m *customFC2ScStatModel) ListAllMovieIDs(ctx context.Context) ([]int64, error) {
	var rows []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT `fc2_movie_id` FROM "+m.table+" WHERE `fc2_movie_id` > 0"); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}
	return rows, nil
}

type customFC2ScStatModel struct {
	*defaultFC2ScStatModel
}

func NewFC2ScStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) FC2ScStatModel {
	return &customFC2ScStatModel{defaultFC2ScStatModel: newFC2ScStatModel(conn, c, opts...)}
}

func (m *customFC2ScStatModel) ListByMovieIDs(ctx context.Context, movieIDs []int64) ([]*FC2ScStat, error) {
	ids := uniquePositiveInt64s(movieIDs)
	if len(ids) == 0 {
		return []*FC2ScStat{}, nil
	}
	rows := make([]*FC2ScStat, 0, len(ids))
	for _, batch := range sqlValueBatches(ids, 0) {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(batch)), ",")
		args := make([]any, len(batch))
		for index, id := range batch {
			args[index] = id
		}
		var batchRows []*FC2ScStat
		if err := m.QueryRowsNoCacheCtx(ctx, &batchRows, "SELECT "+fc2ScStatRows+" FROM "+m.table+" WHERE `fc2_movie_id` IN ("+placeholders+")", args...); err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				continue
			}
			return nil, err
		}
		rows = append(rows, batchRows...)
	}
	return rows, nil
}

func uniquePositiveInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
