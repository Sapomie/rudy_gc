package film_infra

import (
	"context"
	"errors"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/types"
	"time"
)

var _ film_repo.ScRepo = (*ScRepoSqlx)(nil)

type ScRepoSqlx struct {
	m moviex.GScModel
}

func NewScRepoSqlx(m moviex.GScModel) *ScRepoSqlx {
	return &ScRepoSqlx{m: m}
}

func (r *ScRepoSqlx) FindAll(ctx context.Context) ([]*types.GSc, error) {
	rows, err := r.m.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GSc, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapModelToTypes(v))
	}
	return out, nil
}

func (r *ScRepoSqlx) Upsert(ctx context.Context, in *types.GSc) (*types.GSc, error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	now := time.Now().Unix()

	old, err := r.m.FindOneByName(ctx, in.Name)
	if err == nil && old != nil {
		changed := false

		if old.MovieNumber != in.MovieNumber {
			old.MovieNumber = in.MovieNumber
			changed = true
		}
		if old.ScTime != in.ScTime {
			old.ScTime = in.ScTime
			changed = true
		}
		if old.ComeMovieName != in.ComeMovieName {
			old.ComeMovieName = in.ComeMovieName
			changed = true
		}
		if old.Cooldown != in.Cooldown {
			old.Cooldown = in.Cooldown
			changed = true
		}
		if old.Duration != in.Duration {
			old.Duration = in.Duration
			changed = true
		}
		if old.Fg != in.Fg {
			old.Fg = in.Fg
			changed = true
		}
		if old.Vessel != in.Vessel {
			old.Vessel = in.Vessel
			changed = true
		}
		if old.MovieCast != in.MovieCast {
			old.MovieCast = in.MovieCast
			changed = true
		}
		if old.Remarks != in.Remarks {
			old.Remarks = in.Remarks
			changed = true
		}
		if changed {
			old.UpdatedOn = now
			if err := r.m.Update(ctx, old); err != nil {
				return nil, err
			}
		}
		return mapModelToTypes(old), nil
	}

	// 插入
	row := &moviex.GSc{
		Name:          in.Name,
		MovieNumber:   in.MovieNumber,
		ScTime:        in.ScTime,
		ComeMovieName: in.ComeMovieName,
		Cooldown:      in.Cooldown,
		Duration:      in.Duration,
		Fg:            in.Fg,
		Vessel:        in.Vessel,
		MovieCast:     in.MovieCast,
		Remarks:       in.Remarks,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	if _, err := r.m.Insert(ctx, row); err != nil {
		if again, e2 := r.m.FindOneByName(ctx, in.Name); e2 == nil && again != nil {
			return mapModelToTypes(again), nil
		}
		return nil, err
	}
	ins, err := r.m.FindOneByName(ctx, in.Name)
	if err != nil || ins == nil {
		return nil, err
	}
	return mapModelToTypes(ins), nil
}

func (r *ScRepoSqlx) FindTopNRecentSc(ctx context.Context, n uint64) ([]*types.GSc, error) {
	rows, err := r.m.ListTopNByScTime(ctx, n)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GSc, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapModelToTypes(v))
	}
	return out, nil
}

func (r *ScRepoSqlx) FindNearest(ctx context.Context, t int64) (*types.GSc, error) {
	row, err := r.m.FindNearest(ctx, t)
	if err != nil {
		return nil, err
	}
	return mapModelToTypes(row), nil
}

/******** helpers ********/

func mapModelToTypes(v *moviex.GSc) *types.GSc {
	return &types.GSc{
		Id:            v.Id,
		Name:          v.Name,
		MovieNumber:   v.MovieNumber,
		ScTime:        v.ScTime,
		ComeMovieName: v.ComeMovieName,
		Cooldown:      v.Cooldown,
		Duration:      v.Duration,
		Fg:            v.Fg,
		Vessel:        v.Vessel,
		MovieCast:     v.MovieCast,
		Remarks:       v.Remarks,
		CreatedOn:     v.CreatedOn,
		UpdatedOn:     v.UpdatedOn,
	}
}
