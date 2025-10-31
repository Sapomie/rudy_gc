package spider_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type BestinvRepo interface {
	// 已有：按 Name 唯一键幂等保存
	Upsert(ctx context.Context, b *types.Bestinv) error

	// 已有：列出需要扫描的 ID
	ListNeedScanIDs(ctx context.Context, limit int) ([]int64, error)

	// 已有：按主键读取一条
	FindOne(ctx context.Context, id int64) (*types.Bestinv, error)

	// 已有：标记已扫描（NeedScan -> NoNeedScan）
	MarkScanned(ctx context.Context, id int64, ts int64) error

	// ✅ 新增：标记已做过排名检查（NeedRankCheck -> NoNeedRankCheck）
	MarkRankChecked(ctx context.Context, id int64, ts int64) error

	// 已有：按 need_rank_check 筛 ID
	ListIDsByRankCheck(ctx context.Context, flag int64, limit int64) ([]int64, error)

	LatestDayNumber(ctx context.Context) (int64, error)
}
