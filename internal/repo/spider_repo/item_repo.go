// internal/repo/spider_repo/item_repo.go
package spider_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/types"
)

type ItemRepo interface {
	TryInsert(ctx context.Context, it *types.Item) (bool, error)
	FindByDetailStatus(ctx context.Context, status int64) ([]*types.Item, error)
	FindByDetailNeedScan(ctx context.Context, needScan int64) ([]*types.Item, error)
	UpdateDetailMeta(ctx context.Context, id int64, needScan, birthTime, updateTime, updatedOn, hasDetail int64) error

	// ✅ 新增：按 jav_id 查 EItem 记录（透传 modelx 的 FindOneByJavId）
	FindOneByJavId(ctx context.Context, javId string) (*moviex.EItem, error)
}
