package spider

import (
	"context"

	"rudy_gc/internal/model/modelx/moviex"
)

type MovieCastRepoSqlx struct {
	m moviex.AmrMovieCastModel
}

// TryLink 基于 movie_jav_id + cast_id 去重
func (r *MovieCastRepoSqlx) TryLink(ctx context.Context, movieJavId string, castId, ts int64) error {
	// 先查是否已有关系
	exist, err := r.m.FindOneByMovieJavIdCastId(ctx, movieJavId, castId)
	if err == nil && exist != nil {
		// 已存在，直接返回
		return nil
	}

	// 插入新关系
	row := &moviex.AmrMovieCast{
		MovieJavId: movieJavId,
		CastId:     castId,
		CreatedOn:  ts,
		UpdatedOn:  ts,
	}
	_, err = r.m.Insert(ctx, row)
	return err
}

// ListCastIDsByMovieJavId 返回该影片关联的所有演员ID
func (r *MovieCastRepoSqlx) ListCastIDsByMovieJavId(ctx context.Context, movieJavId string) ([]int64, error) {
	return r.m.ListCastIDsByMovieJavId(ctx, movieJavId)
}

func (r *MovieCastRepoSqlx) ListMovieJavIDsByCastID(ctx context.Context, castId int64) ([]string, error) {
	return r.m.ListMovieJavIDsByCastID(ctx, castId)
}
