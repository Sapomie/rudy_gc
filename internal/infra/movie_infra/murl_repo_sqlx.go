// internal/infra/movie_infra/murl_repo_sqlx.go
package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/types"
)

var _ movie_repo.MurlRepo = (*MurlRepoSqlx)(nil)

type MurlRepoSqlx struct {
	m moviex.BmMurlModel
}

func NewMurlRepoSqlx(m moviex.BmMurlModel) movie_repo.MurlRepo {
	return &MurlRepoSqlx{m: m}
}

func (r *MurlRepoSqlx) FindOneByJavId(ctx context.Context, javId string) (*types.Murl, error) {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		return nil, err
	}
	return modelxToTypes(row), nil
}

func (r *MurlRepoSqlx) UpsertByJavIdPreserveLocal(ctx context.Context, in *types.Murl) error {
	if in == nil {
		return errors.New("nil murl")
	}
	old, err := r.m.FindOneByJavId(ctx, in.JavId)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return err
	}

	now := time.Now().Unix()

	if old == nil {
		row := typesToModelx(in)
		if row.CreatedOn == 0 {
			row.CreatedOn = now
		}
		if row.UpdatedOn == 0 {
			row.UpdatedOn = now
		}
		_, err := r.m.Insert(ctx, row)
		return err
	}

	// 更新：保留本地路径字段
	toUpdate := typesToModelx(in)
	toUpdate.Id = old.Id
	if old.JacketImgLocal != "" {
		toUpdate.JacketImgLocal = old.JacketImgLocal
	}
	if old.SmallImgLocal != "" {
		toUpdate.SmallImgLocal = old.SmallImgLocal
	}
	if toUpdate.CreatedOn == 0 {
		toUpdate.CreatedOn = old.CreatedOn
	}
	toUpdate.UpdatedOn = now

	return r.m.Update(ctx, toUpdate)
}

/******** helpers ********/

func modelxToTypes(m *moviex.BmMurl) *types.Murl {
	if m == nil {
		return nil
	}
	return &types.Murl{
		Id:             m.Id,
		JavId:          m.JavId,
		Name:           m.Name,
		JacketImg:      m.JacketImg,
		JacketImgLocal: m.JacketImgLocal,
		SmallImg:       m.SmallImg,
		SmallImgLocal:  m.SmallImgLocal,
		CreatedOn:      m.CreatedOn,
		UpdatedOn:      m.UpdatedOn,
	}
}

func typesToModelx(t *types.Murl) *moviex.BmMurl {
	if t == nil {
		return nil
	}
	return &moviex.BmMurl{
		Id:             t.Id,
		JavId:          t.JavId,
		Name:           t.Name,
		JacketImg:      t.JacketImg,
		JacketImgLocal: t.JacketImgLocal,
		SmallImg:       t.SmallImg,
		SmallImgLocal:  t.SmallImgLocal,
		CreatedOn:      t.CreatedOn,
		UpdatedOn:      t.UpdatedOn,
	}
}
