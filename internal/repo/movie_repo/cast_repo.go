package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type CastRepo interface {
	// 已有
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)

	// 新增：对外都返回 types.Cast
	FindOne(ctx context.Context, id int64) (*types.Cast, error)
	FindOneByName(ctx context.Context, name string) (*types.Cast, error)
	FindByNames(ctx context.Context, names []string) ([]*types.Cast, error)
	CountOwnedScMovieNumbersByNames(ctx context.Context, names []string) (map[string]int64, error)

	// 新增：Upsert（以 name 作为幂等键；存在则更新，不存在则插入）
	Upsert(ctx context.Context, in *types.Cast) (*types.Cast, error)

	// 更新统计字段（仅单行）
	UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error

	// 全量 ID 列表
	ListAllIDs(ctx context.Context) ([]int64, error)
	ListPage(ctx context.Context, page, pageSize int, sortField, sortOrder string, filter types.CastListFilter) ([]*types.Cast, int64, error)
}
