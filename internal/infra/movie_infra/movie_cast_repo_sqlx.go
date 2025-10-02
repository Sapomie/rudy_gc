// internal/infra/movie_infra/movie_cast_repo_sqlx.go
package movie_infra

import (
	"context"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.MovieCastRepo = (*MovieCastRepoSqlx)(nil)

type MovieCastRepoSqlx struct {
	m moviex.AmrMovieCastModel
}

func NewMovieCastRepoSqlx(m moviex.AmrMovieCastModel) movie_repo.MovieCastRepo {
	return &MovieCastRepoSqlx{m: m}
}

func (r *MovieCastRepoSqlx) TryLink(ctx context.Context, movieId, castId, ts int64) error {
	// 先查是否已有关系
	exist, err := r.m.FindOneByMovieIdCastId(ctx, movieId, castId)
	if err == nil && exist != nil {
		// 已存在，直接返回
		return nil
	}

	// 插入新关系
	row := &moviex.AmrMovieCast{
		MovieId:   movieId,
		CastId:    castId,
		CreatedOn: ts,
		UpdatedOn: ts,
	}
	_, err = r.m.Insert(ctx, row)
	return err
}

func (r *MovieCastRepoSqlx) ListCastIDsByMovie(ctx context.Context, movieId int64) ([]int64, error) {
	return r.m.ListCastIDsByMovie(ctx, movieId)
}
