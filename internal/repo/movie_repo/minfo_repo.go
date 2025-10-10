package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/types"
)

type MinfoRepo interface {
	// 已有
	UpsertPreserve(ctx context.Context, minfo *moviex.BmMinfo) error
	UpdateRankStatsByJavId(ctx context.Context, javId string, firstDay, highestRank, daysInRank, updatedOn int64) error

	FindOneByJavId(ctx context.Context, javId string) (*moviex.BmMinfo, error)
	UpdatePartialByJavId(ctx context.Context, javId string, patch types.MinfoPatch) error
}
