package film_infra

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/types"
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

func (r *ScRepoSqlx) FindByNames(ctx context.Context, names []string) ([]*types.GSc, error) {
	rows, err := r.m.ListByNames(ctx, names)
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
		if old.ImagePath != in.ImagePath {
			old.ImagePath = in.ImagePath
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
		ImagePath:     in.ImagePath,
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

func (r *ScRepoSqlx) FindOneByName(ctx context.Context, name string) (*types.GSc, error) {
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return mapModelToTypes(row), nil
}

func (r *ScRepoSqlx) ListPage(ctx context.Context, page, pageSize int, sortField, sortOrder string) ([]*types.GSc, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}

	total, err := r.m.CountAll(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*types.GSc{}, 0, nil
	}

	offset := int64((page - 1) * pageSize)
	rows, err := r.m.ListPage(ctx, offset, int64(pageSize), buildScOrderBy(sortField, sortOrder))
	if err != nil {
		return nil, 0, err
	}

	out := make([]*types.GSc, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapModelToTypes(v))
	}
	return out, total, nil
}

func buildScOrderBy(sortField, sortOrder string) string {
	field := normalizeScSortField(sortField)
	order := normalizeScSortOrder(sortOrder)

	column := "sc_time"
	switch field {
	case "movie_number":
		column = "movie_number"
	case "come_movie_name":
		column = "come_movie_name"
	case "cooldown":
		column = "cooldown"
	case "movie_cast":
		column = "movie_cast"
	case "vessel":
		column = "vessel"
	case "fg":
		column = "fg"
	default:
		column = "sc_time"
	}

	if column == "sc_time" {
		return column + " " + order + ", id DESC"
	}
	return column + " " + order + ", sc_time DESC, id DESC"
}

func normalizeScSortField(sortField string) string {
	switch strings.ToLower(strings.TrimSpace(sortField)) {
	case "movie_number", "come_movie_name", "cooldown", "movie_cast", "vessel", "fg", "sc_time":
		return strings.ToLower(strings.TrimSpace(sortField))
	default:
		return "sc_time"
	}
}

func normalizeScSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "ASC"
	}
	return "DESC"
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
		ImagePath:     v.ImagePath,
		CreatedOn:     v.CreatedOn,
		UpdatedOn:     v.UpdatedOn,
	}
}
