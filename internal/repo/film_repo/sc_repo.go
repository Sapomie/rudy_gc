package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type ScRepo interface {
	FindAll(ctx context.Context) ([]*types.GSc, error)
	FindByNames(ctx context.Context, names []string) ([]*types.GSc, error)
	Upsert(ctx context.Context, in *types.GSc) (*types.GSc, error)
	FindTopNRecentSc(ctx context.Context, n uint64) ([]*types.GSc, error)
	FindNearest(ctx context.Context, t int64) (*types.GSc, error)
	FindOneByName(ctx context.Context, name string) (*types.GSc, error)
	ListPage(ctx context.Context, page, pageSize int, sortField, sortOrder string) ([]*types.GSc, int64, error)
}
