package spider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

type MinfoRepoSqlx struct {
	m moviex.BmMinfoModel
}

func (r *MinfoRepoSqlx) UpsertPreserve(ctx context.Context, in *types.Minfo) error {
	if in == nil {
		return fmt.Errorf("nil minfo")
	}
	// 查旧记录（按 jav_id）
	old, err := r.m.FindOneByJavId(ctx, in.JavId)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return err
	}

	now := time.Now().Unix()
	if old == nil {
		mv := mapTypesToModelx(in)
		if mv.CreatedOn == 0 {
			mv.CreatedOn = now
		}
		if mv.UpdatedOn == 0 {
			mv.UpdatedOn = now
		}
		_, err := r.m.Insert(ctx, mv)
		return err
	}

	// 已存在：更新但保留历史字段
	up := mapTypesToModelx(in)
	up.Id = old.Id

	// 保留既有字段（与你原逻辑一致）
	if old.Chinese != "" {
		up.Chinese = old.Chinese
	}
	up.FirstRankDayNumber = old.FirstRankDayNumber
	up.HighestRank = old.HighestRank
	up.DaysInRank = old.DaysInRank

	up.CreatedOn = old.CreatedOn
	up.UpdatedOn = now

	return r.m.Update(ctx, up)
}

func (r *MinfoRepoSqlx) FindOneByJavId(ctx context.Context, javId string) (*types.Minfo, error) {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		return nil, err
	}
	return mapModelxToTypes(row), nil
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

	if !changed && patch.UpdatedOn == nil {
		return nil
	}
	if patch.UpdatedOn != nil {
		row.UpdatedOn = *patch.UpdatedOn
	} else {
		row.UpdatedOn = time.Now().Unix()
	}
	return r.m.Update(ctx, row) // go-zero 生成的 Update，带缓存失效
}

/************ 映射函数 ************/

func mapTypesToModelx(in *types.Minfo) *moviex.BmMinfo {
	return &moviex.BmMinfo{
		Id:                 in.Id,
		JavId:              in.JavId, // 注意拼写：如果你的 types 字段是 JavId
		Name:               in.Name,
		Chinese:            in.Chinese,
		FirstRankDayNumber: in.FirstRankDayNumber,
		HighestRank:        in.HighestRank,
		DaysInRank:         in.DaysInRank,
		CreatedOn:          in.CreatedOn,
		UpdatedOn:          in.UpdatedOn,
		ReleasingDate:      in.ReleasingDate,
	}
}

func mapModelxToTypes(m *moviex.BmMinfo) *types.Minfo {
	return &types.Minfo{
		Id:                 m.Id,
		JavId:              m.JavId,
		Name:               m.Name,
		Chinese:            m.Chinese,
		FirstRankDayNumber: m.FirstRankDayNumber,
		HighestRank:        m.HighestRank,
		DaysInRank:         m.DaysInRank,
		CreatedOn:          m.CreatedOn,
		UpdatedOn:          m.UpdatedOn,
		ReleasingDate:      m.ReleasingDate,
	}
}
