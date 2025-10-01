package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.CastRepo = (*CastRepoSqlx)(nil)

type CastRepoSqlx struct {
	m moviex.AmCastModel
}

func NewCastRepoSqlx(m moviex.AmCastModel) movie_repo.CastRepo {
	return &CastRepoSqlx{m: m}
}

func (r *CastRepoSqlx) GetOrCreateByName(ctx context.Context, name, javId string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if row != nil {
		return row.Id, nil
	}
	now := time.Now().Unix()
	res, err := r.m.Insert(ctx, &moviex.AmCast{
		Name:      name,
		JavId:     javId,
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}
