package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type MovieRepo interface {
	UpsertByJavId(ctx context.Context, mv *types.Movie) (*types.Movie, error)
	FindOneByJavId(ctx context.Context, javId string) (*types.Movie, error)

	FindMoviesByName(ctx context.Context, name string) ([]*types.Movie, error)
	CountAll(ctx context.Context) (int64, error)
}
