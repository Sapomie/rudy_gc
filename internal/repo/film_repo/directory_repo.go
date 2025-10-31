package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type DirectoryRepo interface {
	// 已有
	GetOrCreateChainWithLevels(ctx context.Context, parts []string) ([4]int64, error)
	FindOneByID(ctx context.Context, id int64) (*types.Directory, error)
	FindOneByName(ctx context.Context, name string) (*types.Directory, error)

	ListRoots(ctx context.Context, page, size int) (items []*types.DirSummary, total int64, err error)
	ListChildren(ctx context.Context, parentID int64, page, size int) ([]*types.DirSummary, int64, error)
	ListSiblings(ctx context.Context, id int64) ([]*types.DirSummary, error)
	BuildBreadcrumbs(ctx context.Context, id int64) ([]types.Breadcrumb, error)
	ListSubtreeIDs(ctx context.Context, id int64) ([]int64, error)
}
