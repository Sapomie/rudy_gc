// internal/repo/spider_repo/item_repo.go
package spider_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type ItemRepo interface {
	TryInsert(ctx context.Context, it *types.Item) (bool, error)

	// 如果还在用，保留；否则可以删掉
	FindByDetailStatus(ctx context.Context, status int64) ([]*types.Item, error)

	// 现在以 DetailNeedScan 为准的查询
	FindByDetailNeedScan(ctx context.Context, needScan int64) ([]*types.Item, error)

	// 统一后的签名：… needScan, birthTime, updateTime, updatedOn, hasDetail
	UpdateDetailMeta(ctx context.Context, id int64, needScan, birthTime, updateTime, updatedOn, hasDetail int64) error
}
