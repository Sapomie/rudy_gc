package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type DirectoryRepo interface {
	GetOrCreateChainWithLevels(ctx context.Context, parts []string) ([4]int64, error)
	FindOneByID(ctx context.Context, id int64) (*types.Directory, error)
	FindOneByName(ctx context.Context, name string) (*types.Directory, error)
}
