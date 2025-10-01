// internal/infra/movie_infra/minfo_repo_sqlx.go
package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
)

var _ movie_repo.MinfoRepo = (*MinfoRepoSqlx)(nil)

type MinfoRepoSqlx struct {
	m moviex.BmMinfoModel
}

func NewMinfoRepoSqlx(m moviex.BmMinfoModel) movie_repo.MinfoRepo {
	return &MinfoRepoSqlx{m: m}
}

func (r *MinfoRepoSqlx) UpsertPreserve(ctx context.Context, in *moviex.BmMinfo) error {
	// 查旧记录（按 jav_id）
	old, err := r.m.FindOneByJavId(ctx, in.JavId)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return err
	}

	now := time.Now().Unix()

	// 不存在：直接插入（CreatedOn/UpdatedOn 兜底）
	if old == nil {
		if in.CreatedOn == 0 {
			in.CreatedOn = now
		}
		if in.UpdatedOn == 0 {
			in.UpdatedOn = now
		}
		_, err := r.m.Insert(ctx, in)
		return err
	}

	// 已存在：更新但保留历史字段
	up := *in
	up.Id = old.Id

	// 保留历史值（如果旧值存在）
	if old.Chinese != "" {
		up.Chinese = old.Chinese
	}
	up.FirstRankDayNumber = old.FirstRankDayNumber
	up.HighestRank = old.HighestRank
	up.DaysInRank = old.DaysInRank
	up.NeedDownload = old.NeedDownload

	// 保留创建时间，更新时间置为 now
	up.CreatedOn = old.CreatedOn
	up.UpdatedOn = now

	return r.m.Update(ctx, &up)
}
