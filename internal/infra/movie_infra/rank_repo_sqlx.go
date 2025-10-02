// internal/infra/movie_infra/rank_repo_sqlx.go
package movie_infra

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/types"
	"time"
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

	// 查找是否存在
	row, err := r.m.FindOneByRankKey(ctx, rk.RankKey)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return fmt.Errorf("find c_rank(%s) failed: %w", rk.RankKey, err)
	}

	if row != nil {
		// 存在 → 更新字段
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

	// 不存在 → 插入
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
func (r *RankRepoSqlx) AggregateByJavId(ctx context.Context, javId string) (int64, int64, int64, error) {
	firstDay, bestRank, daysInRank, err := r.m.AggregateByJavId(ctx, javId)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("聚合排行失败(javId=%s): %w", javId, err)
	}
	return firstDay, bestRank, daysInRank, nil
}
