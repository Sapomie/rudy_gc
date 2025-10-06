package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.LabelRepo = (*LabelRepoSqlx)(nil)

type LabelRepoSqlx struct {
	m moviex.AmLabelModel
}

func NewLabelRepoSqlx(m moviex.AmLabelModel) movie_repo.LabelRepo {
	return &LabelRepoSqlx{m: m}
}

func (r *LabelRepoSqlx) GetOrCreateByName(ctx context.Context, name, javId string) (int64, error) {
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
	res, err := r.m.Insert(ctx, &moviex.AmLabel{
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

func (r *LabelRepoSqlx) FindOne(ctx context.Context, id int64) (*moviex.AmLabel, error) {
	return r.m.FindOne(ctx, id)
}
