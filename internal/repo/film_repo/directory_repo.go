package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type DirSort string

const (
	DirSortName      DirSort = "name"
	DirSortUpdatedOn DirSort = "updated_on"
)

type BucketKind string

const (
	BucketNone  BucketKind = "none"
	BucketMonth BucketKind = "month"
	BucketYear  BucketKind = "year"
)

type DirectoryRepo interface {
	// 已有
	GetOrCreateChainWithLevels(ctx context.Context, parts []string) ([4]int64, error)
	FindOneByID(ctx context.Context, id int64) (*types.Directory, error)
	FindOneByName(ctx context.Context, name string) (*types.Directory, error)

	// 新增（仅目录浏览所需）
	FindOneByPath(ctx context.Context, path string) (*types.Directory, error)
	ListRoots(ctx context.Context, page, size int, sort DirSort, asc bool, withAgg bool) (items []*types.DirSummary, total int64, err error)
	ListChildren(ctx context.Context, parentID int64, page, size int, sort DirSort, asc bool, withAgg bool) (items []*types.DirSummary, total int64, err error)
	ListSiblings(ctx context.Context, id int64) ([]*types.DirSummary, error)
	BuildBreadcrumbs(ctx context.Context, id int64) ([]types.Breadcrumb, error)
	AggregateStats(ctx context.Context, id int64, recursive bool, bucket BucketKind) (*types.DirStats, error)

	ListSubtreeIDs(ctx context.Context, id int64) ([]int64, error)
}
