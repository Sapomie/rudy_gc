package movie_repo

import "context"

type MovieGenreRepo interface {
	TryLink(ctx context.Context, movieId, genreId, ts int64) error
	ListGenreIDsByMovie(ctx context.Context, movieId int64) ([]int64, error)
}
