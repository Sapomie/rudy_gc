package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type MovieRepo interface {
	UpsertByJavId(ctx context.Context, mv *types.Movie) (*types.Movie, error)

	FindByJavId(ctx context.Context, javId string) (*types.Movie, error)
}
