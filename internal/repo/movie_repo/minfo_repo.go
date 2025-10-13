package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type MinfoRepo interface {
	// 改为接收/返回 types.Minfo
	UpsertPreserve(ctx context.Context, minfo *types.Minfo) error
	FindOneByJavId(ctx context.Context, javId string) (*types.Minfo, error)
	UpdatePartialByJavId(ctx context.Context, javId string, patch types.MinfoPatch) error
}
