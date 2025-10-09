// internal/repo/movie_repo/movie_cast_repo.go
package movie_repo

import "context"

type MovieCastRepo interface {
	TryLink(ctx context.Context, movieJavId string, castId, ts int64) error
	ListCastIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error)
}
