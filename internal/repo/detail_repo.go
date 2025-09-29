package repo

import (
	"context"
	"rudy_gc/internal/types"
)

type DetailRepo interface {
	// Upsert：以 JavId 作为幂等键，存在则更新，不存在则插入
	Upsert(ctx context.Context, d *types.Detail) error

	// FindOneByJavId：调试/校验
	FindOneByJavId(ctx context.Context, javId string) (*types.Detail, error)
}
