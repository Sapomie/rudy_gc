// internal/infra/movie_infra/rank_repo_sqlx.go
package movie_infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ movie_repo.RankRepo = (*RankRepoSqlx)(nil)

type RankRepoSqlx struct {
	m moviex.CRankModel
}

func NewRankRepoSqlx(m moviex.CRankModel) movie_repo.RankRepo {
	return &RankRepoSqlx{m: m}
}

// Upsert：按 rank_key 幂等保存（存在则更新，不存在则插入）
func (r *RankRepoSqlx) Upsert(ctx context.Context, rk *types.Rank) error {
	now := time.Now().Unix()

	row, err := r.m.FindOneByRankKey(ctx, rk.RankKey)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return fmt.Errorf("find c_rank(%s) failed: %w", rk.RankKey, err)
	}

	if row != nil {
		// 已存在 -> 更新必要字段
		row.MovieJavId = rk.MovieJavId
		row.DayNumber = rk.DayNumber
		row.RankPos = rk.RankPos
		row.Category = rk.Category
		row.UpdatedOn = now

		if uerr := r.m.Update(ctx, row); uerr != nil {
			return fmt.Errorf("update c_rank(%s) failed: %w", rk.RankKey, uerr)
		}
		return nil
	}

	// 不存在 -> 插入
	toIns := &moviex.CRank{
		RankKey:    rk.RankKey,
		MovieJavId: rk.MovieJavId,
		DayNumber:  rk.DayNumber,
		RankPos:    rk.RankPos,
		Category:   rk.Category,
		CreatedOn:  now,
		UpdatedOn:  now,
	}
	if _, ierr := r.m.Insert(ctx, toIns); ierr != nil {
		return fmt.Errorf("insert c_rank(%s) failed: %w", rk.RankKey, ierr)
	}
	return nil
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
