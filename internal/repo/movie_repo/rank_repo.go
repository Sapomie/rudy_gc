// internal/repo/movie_repo/rank_repo.go
package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type RankRepo interface {
	// 幂等写入：按唯一键（建议 name 或 (movie_jav_id, day_number, category)）插入/更新
	Upsert(ctx context.Context, rk *types.Rank) error
	AggregateByJavId(ctx context.Context, javId string) (firstDay int64, bestRank int64, daysInRank int64, err error)
}
