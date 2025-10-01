package repo

import (
	"context"
	"rudy_gc/internal/types"
)

type ItemRepo interface {
	// 已有的方法……
	TryInsert(ctx context.Context, it *types.Item) (bool, error)
	FindByDetailStatus(ctx context.Context, status int64) ([]*types.Item, error)
	UpdateDetailMeta(ctx context.Context, id int64, needScan, birthTime, updateTime, updatedOn, hasDetail int64) error

	// 新增：按 DetailNeedScan 状态查找
	FindByDetailNeedScan(ctx context.Context, needScan int64) ([]*types.Item, error)
}
