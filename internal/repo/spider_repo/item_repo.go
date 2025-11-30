// internal/repo/spider_repo/item_repo.go
package spider_repo

import (
	"context"
	"rudy_gc/internal/types"
)

// internal/repo/spider_repo/item_repo.go
type ItemRepo interface {
	TryInsert(ctx context.Context, it *types.Item) (bool, error)
	// 现有：按 javId 查 todo:types.Item
	UpdatePartialByJavId(ctx context.Context, javId string, patch types.ItemPatch) error
	UpdateDetailMeta(ctx context.Context, id int64, needScan, birthTime, updateTime, updatedOn, hasDetail int64) error
	FindOneByJavId(ctx context.Context, javId string) (*types.Item, error)

	FindByDetailStatus(ctx context.Context, status int64) ([]*types.Item, error)
	FindByDetailNeedScan(ctx context.Context, needScan int64) ([]*types.Item, error)
	FindByDownloadCoverStatus(ctx context.Context, downloadCoverStatus int64) ([]*types.Item, error)
	FindByTranslateStatus(ctx context.Context, translateStatus int64) ([]*types.Item, error)

	FindOldestByLastQueryDetailTime(ctx context.Context, num int64) ([]*types.Item, error)
}
