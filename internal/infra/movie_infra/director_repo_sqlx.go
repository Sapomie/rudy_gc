package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

type directorRepoSqlx struct {
	model moviex.AmDirectorModel
}

func NewDirectorRepoSqlx(model moviex.AmDirectorModel) movie_repo.DirectorRepo {
	return &directorRepoSqlx{model: model}
}

func (r *directorRepoSqlx) GetOrCreateByName(ctx context.Context, name, javId string) (int64, error) {
	if name == "" {
		return 0, nil
	}

	// 先查（通过 Name 唯一）
	row, err := r.model.FindOneByName(ctx, name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if row != nil {
		return row.Id, nil
	}

	// 插入（JavId 如果有就存，没有就空串）
	now := time.Now().Unix()
	res, err := r.model.Insert(ctx, &moviex.AmDirector{
		Name:             name,
		JavId:            javId,
		MovieNumber:      0,
		OwnedMovieNumber: 0,
		CreatedOn:        now,
		UpdatedOn:        now,
	})
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}
