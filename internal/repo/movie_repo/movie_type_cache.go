// internal/repo/movie_repo/movie_type_cache.go
package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
	"time"
)

type MovieTypeCache interface {
	GetMovieType(ctx context.Context, javId string) (*types.MovieType, error)
	SetMovieType(ctx context.Context, javId string, v *types.MovieType, ttl time.Duration) error
	DelMovieType(ctx context.Context, javId string) error
}
