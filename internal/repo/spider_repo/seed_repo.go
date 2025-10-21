package spider_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type SeedRepo interface {
	FindActiveByNameType(ctx context.Context, nameType int64) ([]*types.Seed, error)

	// UpdateProgress 更新抓取进度
	UpdateProgress(ctx context.Context, id int64, pageNow int64, lastQueryTime int64, lastStatus int64, lastError string) error

	Upsert(ctx context.Context, seed *types.Seed) (int64, error)

	// 新增：按 name 查询一条
	FindOneByName(ctx context.Context, name string) (*types.Seed, error)
}
