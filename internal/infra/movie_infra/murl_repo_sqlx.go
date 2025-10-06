package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.MurlRepo = (*MurlRepoSqlx)(nil)

type MurlRepoSqlx struct {
	m moviex.BmMurlModel
}

func NewMurlRepoSqlx(m moviex.BmMurlModel) movie_repo.MurlRepo {
	return &MurlRepoSqlx{m: m}
}

func (r *MurlRepoSqlx) FindOneByJavId(ctx context.Context, javId string) (*moviex.BmMurl, error) {
	return r.m.FindOneByJavId(ctx, javId)
}

func (r *MurlRepoSqlx) UpsertByJavIdPreserveLocal(ctx context.Context, murl *moviex.BmMurl) error {
	old, err := r.m.FindOneByJavId(ctx, murl.JavId)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return err
	}

	now := time.Now().Unix()

	if old == nil {
		if murl.CreatedOn == 0 {
			murl.CreatedOn = now
		}
		if murl.UpdatedOn == 0 {
			murl.UpdatedOn = now
		}
		_, err := r.m.Insert(ctx, murl)
		return err
	}

	toUpdate := *murl
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

	return r.m.Update(ctx, &toUpdate)
}
