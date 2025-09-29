// internal/repo/item_repo.go
package repo

import (
	"context"
	"rudy_gc/internal/types"
)

type ItemRepo interface {
	TryInsert(ctx context.Context, it *types.Item) (bool, error)
	FindByDetailStatus(ctx context.Context, status int64) ([]*types.Item, error)

	// 新增：只更新与详情相关的元字段
	UpdateDetailMeta(ctx context.Context, id int64, needScan, birthTime, updateTime, updatedOn int64) error

	// 如你还在别处用到可保留：
	MarkHasDetail(ctx context.Context, id int64, newStatus int64, ts int64) error
}
