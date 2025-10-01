package spider_infra

import (
	"context"
	"fmt"
	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/repo/spider_repo"

	"rudy_gc/internal/types"
)

type SeedRepoSqlx struct {
	m spiderx.DSeedModel
}

// 构造函数
func NewSeedRepoSqlx(m spiderx.DSeedModel) spider_repo.SeedRepo {
	return &SeedRepoSqlx{m: m}
}

// 实现 repo.SeedRepo 接口
func (r *SeedRepoSqlx) FindActiveByNameType(ctx context.Context, nameType int64) ([]*types.Seed, error) {
	rows, err := r.m.FindQueriesActive(ctx, nameType)
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

func (r *SeedRepoSqlx) UpdateProgress(ctx context.Context, id int64, pageNow int64, lastQueryTime int64, lastStatus int64, lastError string) error {
	// 用 goctl 的 Update 实现更新
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return fmt.Errorf("FindOne(id=%d) error: %w", id, err)
	}

	row.PageNow = pageNow
	row.LastQueryTime = lastQueryTime
	row.LastStatus = lastStatus
	row.LastError = lastError
	row.UpdatedOn = lastQueryTime // 或者 time.Now().Unix()

	return r.m.Update(ctx, row)
}
