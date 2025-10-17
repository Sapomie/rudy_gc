package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type ScRepo interface {
	FindAll(ctx context.Context) ([]*types.GSc, error)
	Upsert(ctx context.Context, in *types.GSc) (*types.GSc, error)
	FindTopNRecentSc(ctx context.Context, n uint64) ([]*types.GSc, error)
	FindNearest(ctx context.Context, t int64) (*types.GSc, error)
}
