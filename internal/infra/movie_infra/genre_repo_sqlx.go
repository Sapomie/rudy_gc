package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.GenreRepo = (*GenreRepoSqlx)(nil)

type GenreRepoSqlx struct {
	m moviex.AmGenreModel
}

func NewGenreRepoSqlx(m moviex.AmGenreModel) movie_repo.GenreRepo {
	return &GenreRepoSqlx{m: m}
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
