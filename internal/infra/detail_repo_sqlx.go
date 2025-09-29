package infra

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/repo"
	"rudy_gc/internal/types"
)

var _ repo.DetailRepo = (*DetailRepoSqlx)(nil)

type DetailRepoSqlx struct {
	m spiderx.DDetailModel
}

func NewDetailRepoSqlx(m spiderx.DDetailModel) *DetailRepoSqlx {
	return &DetailRepoSqlx{m: m}
}

func (r *DetailRepoSqlx) Upsert(ctx context.Context, d *types.Detail) error {
	row, err := r.m.FindOneByJavId(ctx, d.JavId)
	if err == nil && row != nil {
		// update
		row.Name = d.Name
		row.Prefix = d.Prefix
		row.QueryUrl = d.QueryUrl
		row.Content = d.Content
		row.LastQueryTime = d.LastQueryTime
		row.UpdatedOn = d.UpdatedOn
		return r.m.Update(ctx, row)
	}
	if err != nil && !errors.Is(err, spiderx.ErrNotFound) {
		return fmt.Errorf("FindOneByJavId(%s): %w", d.JavId, err)
	}

	// insert
	_, ierr := r.m.Insert(ctx, &spiderx.DDetail{
		Name:          d.Name,
		JavId:         d.JavId,
		Prefix:        d.Prefix,
		QueryUrl:      d.QueryUrl,
		Content:       d.Content,
		LastQueryTime: d.LastQueryTime,
		CreatedOn:     d.CreatedOn,
		UpdatedOn:     d.UpdatedOn,
	})
	return ierr
}

func (r *DetailRepoSqlx) FindOneByJavId(ctx context.Context, javId string) (*types.Detail, error) {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		return nil, err
	}
	return detailRowToType(row), nil
}

func detailRowToType(row *spiderx.DDetail) *types.Detail {
	if row == nil {
		return nil
	}
	return &types.Detail{
		Id:            row.Id,
		Name:          row.Name,
		JavId:         row.JavId,
		Prefix:        row.Prefix,
		QueryUrl:      row.QueryUrl,
		Content:       row.Content,
		LastQueryTime: row.LastQueryTime,
		CreatedOn:     row.CreatedOn,
		UpdatedOn:     row.UpdatedOn,
	}
}
