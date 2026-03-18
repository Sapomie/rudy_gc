package spider

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
)

type PrefixRepoSqlx struct {
	m moviex.AmPrefixModel
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

// 已有的 GetOrCreateByName 省略

func (r *PrefixRepoSqlx) FindOne(ctx context.Context, id int64) (*moviex.AmPrefix, error) {
	return r.m.FindOne(ctx, id)
}

func (r *PrefixRepoSqlx) UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error {
	movieNumber, ownedMovieNumber, err := r.m.GetMovieNumbersByID(ctx, id, ownedRemovedStatus)
	if err != nil {
		return err
	}

	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber {
		return nil
	}

	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedMovieNumber
	row.UpdatedOn = now

	return r.m.Update(ctx, row)
}

func (r *PrefixRepoSqlx) ListAllIDs(ctx context.Context) ([]int64, error) {
	return r.m.ListAllIDs(ctx)
}
