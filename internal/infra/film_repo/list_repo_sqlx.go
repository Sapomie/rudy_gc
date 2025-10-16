package film_infra

import (
	"context"
	"errors"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/types"
	"time"
)

var _ film_repo.GListRepo = (*GListRepoSqlx)(nil)

type GListRepoSqlx struct {
	m moviex.GListModel
}

func NewGListRepoSqlx(m moviex.GListModel) *GListRepoSqlx {
	return &GListRepoSqlx{m: m}
}

func (r *GListRepoSqlx) FindAll(ctx context.Context) ([]*types.GList, error) {
	rows, err := r.m.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, v := range rows {
		out = append(out, glistMapModelxToTypes(v))
	}
	return out, nil
}

func (r *GListRepoSqlx) Upsert(ctx context.Context, in *types.GList) (*types.GList, error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	now := time.Now().Unix()

	old, err := r.m.FindOneByName(ctx, in.Name)
	if err == nil && old != nil {
		changed := false
		if old.ScName != in.ScName {
			old.ScName = in.ScName
			changed = true
		}
		if old.MovieJavId != in.MovieJavId {
			old.MovieJavId = in.MovieJavId
			changed = true
		}
		if old.IsCome != in.IsCome {
			old.IsCome = in.IsCome
			changed = true
		}
		if changed {
			old.UpdatedOn = now
			if err := r.m.Update(ctx, old); err != nil {
				return nil, err
			}
		}
		return glistMapModelxToTypes(old), nil
	}

	// 插入
	row := &moviex.GList{
		Name:       in.Name,
		ScName:     in.ScName,
		MovieJavId: in.MovieJavId,
		IsCome:     in.IsCome,
		CreatedOn:  now,
		UpdatedOn:  now,
	}
	if _, err := r.m.Insert(ctx, row); err != nil {
		// 并发兜底
		if re, e2 := r.m.FindOneByName(ctx, in.Name); e2 == nil && re != nil {
			return glistMapModelxToTypes(re), nil
		}
		return nil, err
	}
	// 读回
	re, err := r.m.FindOneByName(ctx, in.Name)
	if err != nil || re == nil {
		return nil, err
	}
	return glistMapModelxToTypes(re), nil
}

func (r *GListRepoSqlx) FindGList(ctx context.Context, scName string, isCome *int64, page, pageSize int) ([]*types.GList, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	offset := int64((page - 1) * pageSize)

	rows, err := r.m.ListByFilters(ctx, scName, isCome, offset, int64(pageSize))
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, v := range rows {
		out = append(out, glistMapModelxToTypes(v))
	}
	return out, nil
}

func (r *GListRepoSqlx) FindGListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*types.GList, error) {
	rows, err := r.m.ListByMovieJavIds(ctx, movieJavIds)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapGListModelToTypes(v))
	}
	return out, nil
}

func (r *GListRepoSqlx) FindGListByMovieJavId(ctx context.Context, movieJavId string) ([]*types.GList, error) {
	rows, err := r.m.ListByMovieJavId(ctx, movieJavId)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapGListModelToTypes(v))
	}
	return out, nil
}

/******** helpers ********/

func mapGListModelToTypes(g *moviex.GList) *types.GList {
	if g == nil {
		return nil
	}
	return &types.GList{
		Id:         g.Id,
		Name:       g.Name,
		ScName:     g.ScName,
		MovieJavId: g.MovieJavId,
		IsCome:     g.IsCome,
		CreatedOn:  g.CreatedOn,
		UpdatedOn:  g.UpdatedOn,
	}
}

/******** helpers ********/

func glistMapModelxToTypes(v *moviex.GList) *types.GList {
	return &types.GList{
		Id:         v.Id,
		Name:       v.Name,
		ScName:     v.ScName,
		MovieJavId: v.MovieJavId,
		IsCome:     v.IsCome,
		CreatedOn:  v.CreatedOn,
		UpdatedOn:  v.UpdatedOn,
	}
}
