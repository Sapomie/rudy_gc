package spider

import (
	"context"

	"rudy_gc/data/modelx/moviex"
)

type MovieGenreRepoSqlx struct {
	m moviex.AmrMovieGenreModel
}

// TryLink 基于 movie_jav_id + genre_id 去重建立关系
func (r *MovieGenreRepoSqlx) TryLink(ctx context.Context, movieJavId string, genreId, ts int64) error {
	// 已存在则直接返回
	if exist, err := r.m.FindOneByMovieJavIdGenreId(ctx, movieJavId, genreId); err == nil && exist != nil {
		return nil
	}

	// 插入
	row := &moviex.AmrMovieGenre{
		MovieJavId: movieJavId,
		GenreId:    genreId,
		CreatedOn:  ts,
		UpdatedOn:  ts,
	}
	_, err := r.m.Insert(ctx, row)
	return err
}

// ListGenreIDsByMovieJavId 返回该影片关联的所有 genre_id（不走缓存）
func (r *MovieGenreRepoSqlx) ListGenreIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error) {
	return r.m.ListGenreIDsByMovieJavId(ctx, movieJavId)
}
