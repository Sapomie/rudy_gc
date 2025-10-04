// internal/infra/movie_infra/cafo_repo_sqlx.go
package movie_infra

import (
	"context"
	"errors"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.CafoRepo = (*CafoRepoSqlx)(nil)

type CafoRepoSqlx struct {
	m moviex.CCafoModel
}

func NewCafoRepoSqlx(m moviex.CCafoModel) movie_repo.CafoRepo {
	return &CafoRepoSqlx{m: m}
}

// FindBirthByName 按名字读取生日（Unix 秒）。若不存在，found=false。
func (r *CafoRepoSqlx) FindBirthByName(ctx context.Context, name string) (int64, bool, error) {
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return row.BirthDay, true, nil
}
