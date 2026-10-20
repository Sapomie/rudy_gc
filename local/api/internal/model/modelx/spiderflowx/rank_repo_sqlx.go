package spiderflowx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RankRepoSqlx struct {
	m moviex.CRankModel
}

func (r *RankRepoSqlx) FindHighestRank(ctx context.Context, movieJavId string, limit uint64) ([]*types.Rank, error) {
	rows, err := r.m.FindHighestRank(ctx, movieJavId, limit)
	if err != nil {
		// 查无数据时返回空列表而不是报错
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*types.Rank{}, nil
		}
		return nil, err
	}
	out := make([]*types.Rank, 0, len(rows))
	for _, v := range rows {
		if v == nil {
			continue
		}
		out = append(out, mapCRankToTypes(v))
	}
	return out, nil
}

func (r *RankRepoSqlx) ListByDayNumber(ctx context.Context, dayNumber int64) ([]*types.Rank, error) {
	rows, err := r.m.FindByDayNumber(ctx, dayNumber)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Rank, 0, len(rows))
	for _, v := range rows {
		if v == nil {
			continue
		}
		out = append(out, mapCRankToTypes(v))
	}
	return out, nil
}

func (r *RankRepoSqlx) FindEarliestDayNumber(ctx context.Context) (int64, error) {
	return r.m.FindEarliestDayNumber(ctx)
}

func (r *RankRepoSqlx) FindLatestDayNumber(ctx context.Context) (int64, error) {
	return r.m.FindLatestDayNumber(ctx)
}

func (r *RankRepoSqlx) InsertByDayNumberCategory(
	ctx context.Context,
	dayNumber, category int64,
	ranks []*types.Rank,
) (bool, error) {
	now := time.Now().Unix()
	rows := make([]*moviex.CRank, 0, len(ranks))
	for _, rank := range ranks {
		if rank == nil {
			return false, fmt.Errorf("c_rank 首次写入数据存在空记录")
		}
		rows = append(rows, &moviex.CRank{
			RankKey:    rank.RankKey,
			MovieJavId: rank.MovieJavId,
			DayNumber:  rank.DayNumber,
			RankPos:    rank.RankPos,
			Category:   rank.Category,
			CreatedOn:  now,
			UpdatedOn:  now,
		})
	}
	return r.m.InsertByDayNumberCategory(ctx, dayNumber, category, rows)
}

func (r *RankRepoSqlx) AggregateByJavId(ctx context.Context, javId string) (int64, int64, int64, error) {
	firstDay, bestRank, daysInRank, err := r.m.AggregateByJavId(ctx, javId)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("聚合排行失败(javId=%s): %w", javId, err)
	}
	return firstDay, bestRank, daysInRank, nil
}

func (r *RankRepoSqlx) FindOneByRankKey(ctx context.Context, rankKey string) (*types.Rank, error) {
	row, err := r.m.FindOneByRankKey(ctx, rankKey)
	if err != nil {
		return nil, err
	}
	return mapCRankToTypes(row), nil
}

func (r *RankRepoSqlx) All(ctx context.Context) ([]*types.Rank, error) {
	rows, err := r.m.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Rank, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapCRankToTypes(v))
	}
	return out, nil
}

/******** helpers ********/
func mapCRankToTypes(v *moviex.CRank) *types.Rank {
	return &types.Rank{
		RankKey:    v.RankKey,
		MovieJavId: v.MovieJavId,
		DayNumber:  v.DayNumber,
		RankPos:    v.RankPos,
		Category:   v.Category,
	}
}
