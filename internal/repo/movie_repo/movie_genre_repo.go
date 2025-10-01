// internal/repo/movie_repo/movie_genre_repo.go
package movie_repo

import "context"

type MovieGenreRepo interface {
	// TryLink: 以 (movie_id, genre_id) 为幂等键创建关系；若已存在则不报错。
	TryLink(ctx context.Context, movieId, genreId, ts int64) error
}
