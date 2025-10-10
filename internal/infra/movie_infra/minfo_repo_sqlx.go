// internal/infra/movie_infra/minfo_repo_sqlx.go
package movie_infra

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/internal/types"
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

func (r *MinfoRepoSqlx) UpdateRankStatsByJavId(ctx context.Context, javId string, firstDay, highestRank, daysInRank, updatedOn int64) error {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		return fmt.Errorf("查询 minfo 失败(javId=%s): %w", javId, err)
	}
	// 更新排行相关字段
	row.FirstRankDayNumber = firstDay
	row.HighestRank = highestRank
	row.DaysInRank = daysInRank
	row.UpdatedOn = updatedOn

	if err := r.m.Update(ctx, row); err != nil {
		return fmt.Errorf("更新 minfo 失败(javId=%s): %w", javId, err)
	}
	return nil
}

// ✅ 新增：按 jav_id 查询
func (r *MinfoRepoSqlx) FindOneByJavId(ctx context.Context, javId string) (*moviex.BmMinfo, error) {
	return r.m.FindOneByJavId(ctx, javId)
}

func (r *MinfoRepoSqlx) UpdatePartialByJavId(ctx context.Context, javId string, patch types.MinfoPatch) error {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		return err
	}
	changed := false

	if patch.Chinese != nil && row.Chinese != *patch.Chinese {
		row.Chinese = *patch.Chinese
		changed = true
	}
	if patch.EncodeName != nil && row.EncodeName != *patch.EncodeName {
		row.EncodeName = *patch.EncodeName
		changed = true
	}
	if patch.FirstRankDayNumber != nil && row.FirstRankDayNumber != *patch.FirstRankDayNumber {
		row.FirstRankDayNumber = *patch.FirstRankDayNumber
		changed = true
	}
	if patch.HighestRank != nil && row.HighestRank != *patch.HighestRank {
		row.HighestRank = *patch.HighestRank
		changed = true
	}
	if patch.DaysInRank != nil && row.DaysInRank != *patch.DaysInRank {
		row.DaysInRank = *patch.DaysInRank
		changed = true
	}
	if patch.NeedDownload != nil && row.NeedDownload != *patch.NeedDownload {
		row.NeedDownload = *patch.NeedDownload
		changed = true
	}

	if !changed && patch.UpdatedOn == nil {
		// 没任何业务字段变化且未强制更新时间 → 直接返回
		return nil
	}

	if patch.UpdatedOn != nil {
		row.UpdatedOn = *patch.UpdatedOn
	} else {
		row.UpdatedOn = time.Now().Unix()
	}

	return r.m.Update(ctx, row) // 用 go-zero 生成的 Update，自动清缓存
}
