package spider

import (
	"context"
	"errors"

	"rudy_gc/internal/model/modelx/moviex"
)

type CafoRepoSqlx struct {
	m moviex.CCafoModel
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
