package movie_infra

import (
	"context"

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

// 已有的 TryLink：建立 movie-genre 关系
func (r *MovieGenreRepoSqlx) TryLink(ctx context.Context, movieId, genreId, ts int64) error {
	exist, err := r.m.FindOneByMovieIdGenreId(ctx, movieId, genreId)
	if err == nil && exist != nil {
		return nil
	}
	row := &moviex.AmrMovieGenre{
		MovieId:   movieId,
		GenreId:   genreId,
		CreatedOn: ts,
		UpdatedOn: ts,
	}
	_, err = r.m.Insert(ctx, row)
	return err
}

// ✅ 新增：列出该电影的所有 genreId
func (r *MovieGenreRepoSqlx) ListGenreIDsByMovie(ctx context.Context, movieId int64) ([]int64, error) {
	return r.m.ListGenreIDsByMovie(ctx, movieId)
}
