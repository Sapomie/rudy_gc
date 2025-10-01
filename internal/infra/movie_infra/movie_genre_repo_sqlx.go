// internal/infra/movie_infra/movie_genre_repo_sqlx.go
package movie_infra

import (
	"context"
	"errors"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.MovieGenreRepo = (*MovieGenreRepoSqlx)(nil)

type MovieGenreRepoSqlx struct {
	m moviex.AmrMovieGenreModel
}

func NewMovieGenreRepoSqlx(m moviex.AmrMovieGenreModel) movie_repo.MovieGenreRepo {
	return &MovieGenreRepoSqlx{m: m}
}

func (r *MovieGenreRepoSqlx) TryLink(ctx context.Context, movieId, genreId, ts int64) error {
	// 先查（FindOneByMovieIdGenreId 走缓存，命中则不会打 DB）
	row, err := r.m.FindOneByMovieIdGenreId(ctx, movieId, genreId)
	if err == nil && row != nil {
		return nil // 已存在关系，幂等返回
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return err // 真正异常
	}

	// 不存在才插入
	_, ierr := r.m.Insert(ctx, &moviex.AmrMovieGenre{
		MovieId:   movieId,
		GenreId:   genreId,
		CreatedOn: ts,
		UpdatedOn: ts,
	})
	return ierr
}
