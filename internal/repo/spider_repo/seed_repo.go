package spider_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type SeedRepo interface {
	FindActiveByNameType(ctx context.Context, nameType int64) ([]*types.Seed, error)

	// UpdateProgress 更新抓取进度
	UpdateProgress(ctx context.Context, id int64, pageNow int64, lastQueryTime int64, lastStatus int64, lastError string) error
}
