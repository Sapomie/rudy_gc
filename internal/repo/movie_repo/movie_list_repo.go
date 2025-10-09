package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type MovieListRepo interface {
	ListFull(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error)
}
