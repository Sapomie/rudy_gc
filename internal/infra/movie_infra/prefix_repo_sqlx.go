package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.PrefixRepo = (*PrefixRepoSqlx)(nil)

type PrefixRepoSqlx struct {
	m moviex.AmPrefixModel
}

func NewPrefixRepoSqlx(m moviex.AmPrefixModel) movie_repo.PrefixRepo {
	return &PrefixRepoSqlx{m: m}
}

func (r *PrefixRepoSqlx) GetOrCreateByName(ctx context.Context, name string) (int64, error) {
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
	res, err := r.m.Insert(ctx, &moviex.AmPrefix{
		Name:      name,
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

var _ movie_repo.PrefixRepo = (*PrefixRepoSqlx)(nil)

// 已有的 GetOrCreateByName 省略

func (r *PrefixRepoSqlx) FindOne(ctx context.Context, id int64) (*moviex.AmPrefix, error) {
	return r.m.FindOne(ctx, id)
}
