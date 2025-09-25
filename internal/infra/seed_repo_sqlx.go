package infra

import (
	"context"
	"rudy_gc/data/modelx/spider"

	"rudy_gc/internal/repo"
	"rudy_gc/internal/types"
)

type seedRepoSqlx struct {
	model spider.DSeedModel
}

// 构造函数
func NewSeedRepoSqlx(m spider.DSeedModel) repo.SeedRepo {
	return &seedRepoSqlx{model: m}
}

// 实现 repo.SeedRepo 接口
func (r *seedRepoSqlx) FindActiveByNameType(ctx context.Context, nameType int64) ([]*types.Seed, error) {
	rows, err := r.model.FindQueriesActive(ctx, nameType)
	if err != nil {
		return nil, err
	}

	var seeds []*types.Seed
	for _, row := range rows {
		seeds = append(seeds, &types.Seed{
			Id:            row.Id,
			Name:          row.Name,
			Active:        row.Active,
			SearchType:    row.SearchType,
			NameType:      row.NameType,
			PageNow:       row.PageNow,
			Offset:        row.Offset,
			StartPage:     row.StartPage,
			EndPage:       row.EndPage,
			LastQueryTime: row.LastQueryTime,
			LastStatus:    row.LastStatus,
			LastError:     row.LastError,
			CreatedOn:     row.CreatedOn,
			UpdatedOn:     row.UpdatedOn,
		})
	}
	return seeds, nil
}
