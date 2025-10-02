package spider_repo

import (
	"context"

	"rudy_gc/internal/types"
)

type BestinvRepo interface {
	// 已有：按 Name 唯一键幂等保存
	Upsert(ctx context.Context, b *types.Bestinv) error

	// 新增：列出需要扫描的 ID（加 limit，避免一次性扫太多）
	ListNeedScanIDs(ctx context.Context, limit int) ([]int64, error)

	// 新增：按主键读取一条
	FindOne(ctx context.Context, id int64) (*types.Bestinv, error)

	// 新增：标记已扫描（NeedScan -> NoNeedScan），并更新 UpdatedOn
	MarkScanned(ctx context.Context, id int64, ts int64) error
}
