// internal/repo/movie_repo/movie_cast_repo.go
package movie_repo

import "context"

type MovieCastRepo interface {
	// TryLink: 以 (movie_id, cast_id) 为幂等键创建关系；若已存在则不报错。
	TryLink(ctx context.Context, movieId, castId, ts int64) error
}
