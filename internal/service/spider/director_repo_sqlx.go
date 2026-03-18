package spider

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
)

type DirectorRepoSqlx struct {
	m moviex.AmDirectorModel
}

func (r *DirectorRepoSqlx) GetOrCreateByName(ctx context.Context, name, javId string) (int64, error) {
	if name == "" {
		return 0, nil
	}

	// 先查（通过 Name 唯一）
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if row != nil {
		return row.Id, nil
	}

	// 插入（JavId 如果有就存，没有就空串）
	now := time.Now().Unix()
	res, err := r.m.Insert(ctx, &moviex.AmDirector{
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

func (r *DirectorRepoSqlx) FindOne(ctx context.Context, id int64) (*moviex.AmDirector, error) {
	return r.m.FindOne(ctx, id)
}

func (r *DirectorRepoSqlx) UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error {
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

func (r *DirectorRepoSqlx) ListAllIDs(ctx context.Context) ([]int64, error) {
	return r.m.ListAllIDs(ctx)
}
