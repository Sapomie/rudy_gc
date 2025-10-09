package movie_repo

import "context"

type MovieGenreRepo interface {
	TryLink(ctx context.Context, movieJavId string, genreId, ts int64) error
	ListGenreIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error)
}
