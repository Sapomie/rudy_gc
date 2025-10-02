package spider_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type BestinvRepo interface {
	// 幂等保存：按 Name 唯一键插入/更新
	Upsert(ctx context.Context, b *types.Bestinv) error
}
