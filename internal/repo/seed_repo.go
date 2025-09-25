package repo

import (
	"context"
	"rudy_gc/internal/types"
)

type SeedRepo interface {
	FindActiveByNameType(ctx context.Context, nameType int64) ([]*types.Seed, error)
}
