package spider

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/types"
)

type DetailRepoSqlx struct {
	m spiderx.DDetailModel
}

func (r *DetailRepoSqlx) Upsert(ctx context.Context, d *types.Detail) error {
	row, err := r.m.FindOneByJavId(ctx, d.JavId)
	if err == nil && row != nil {
		// update
		row.Name = d.Name
		row.Prefix = d.Prefix
		row.QueryUrl = d.QueryUrl
		row.Content = d.Content
		row.UpdatedOn = d.UpdatedOn
		return r.m.Update(ctx, row)
	}
	if err != nil && !errors.Is(err, spiderx.ErrNotFound) {
		return fmt.Errorf("FindOneByJavId(%s): %w", d.JavId, err)
	}

	// insert
	_, ierr := r.m.Insert(ctx, &spiderx.DDetail{
		Name:      d.Name,
		JavId:     d.JavId,
		Prefix:    d.Prefix,
		QueryUrl:  d.QueryUrl,
		Content:   d.Content,
		CreatedOn: d.CreatedOn,
		UpdatedOn: d.UpdatedOn,
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
		Id:        row.Id,
		Name:      row.Name,
		JavId:     row.JavId,
		Prefix:    row.Prefix,
		QueryUrl:  row.QueryUrl,
		Content:   row.Content,
		CreatedOn: row.CreatedOn,
		UpdatedOn: row.UpdatedOn,
	}
}

func (r *DetailRepoSqlx) AllJavIds(ctx context.Context) ([]string, error) {
	return r.m.ListAllJavIds(ctx)
}
