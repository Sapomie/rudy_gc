package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.MovieRepo = (*MovieRepoSqlx)(nil)

type MovieRepoSqlx struct {
	m moviex.AMovieModel
}

func NewMovieRepoSqlx(m moviex.AMovieModel) movie_repo.MovieRepo {
	return &MovieRepoSqlx{m: m}
}

func (r *MovieRepoSqlx) UpsertByJavId(ctx context.Context, mv *moviex.AMovie) (*moviex.AMovie, error) {
	// 1) 先按 jav_id 查是否已存在
	old, err := r.m.FindOneByJavId(ctx, mv.JavId)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}

	now := time.Now().Unix()

	// 2) 不存在 -> Insert
	if old == nil {
		if mv.CreatedOn == 0 {
			mv.CreatedOn = now
		}
		if mv.UpdatedOn == 0 {
			mv.UpdatedOn = now
		}
		if _, err := r.m.Insert(ctx, mv); err != nil {
			return nil, err
		}
		// 回查以获得自增 id 等字段
		return r.m.FindOneByJavId(ctx, mv.JavId)
	}

	// 3) 已存在 -> Update（保留旧 CreatedOn）
	toUpdate := *mv
	toUpdate.Id = old.Id
	if toUpdate.CreatedOn == 0 {
		toUpdate.CreatedOn = old.CreatedOn
	}
	if toUpdate.UpdatedOn == 0 {
		toUpdate.UpdatedOn = now
	}
	if err := r.m.Update(ctx, &toUpdate); err != nil {
		return nil, err
	}
	return r.m.FindOne(ctx, old.Id)
}
