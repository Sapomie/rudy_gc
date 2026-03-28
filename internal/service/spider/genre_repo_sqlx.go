package spider

import (
	"context"
	"errors"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

type GenreRepoSqlx struct {
	m moviex.AmGenreModel
}

func (r *GenreRepoSqlx) GetOrCreateByName(ctx context.Context, name, javId string) (int64, error) {
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
	res, err := r.m.Insert(ctx, &moviex.AmGenre{
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

// 新增：按 ID 查，直接透传 modelx
func (r *GenreRepoSqlx) FindOne(ctx context.Context, id int64) (*moviex.AmGenre, error) {
	return r.m.FindOne(ctx, id)
}

func (r *GenreRepoSqlx) UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error {
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

func (r *GenreRepoSqlx) ListAllIDs(ctx context.Context) ([]int64, error) {
	return r.m.ListAllIDs(ctx)
}
