package spider_infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"

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

func (r *SeedRepoSqlx) FindActiveByNameType(ctx context.Context, nameType int64) ([]*types.Seed, error) {
	rows, err := r.m.FindQueriesActive(ctx, nameType)
	if err != nil {
		return nil, err
	}
	var seeds []*types.Seed
	for _, row := range rows {
		seeds = append(seeds, toSeed(row))
	}
	return seeds, nil
}

func (r *SeedRepoSqlx) UpdateProgress(ctx context.Context, id int64, pageNow int64, lastQueryTime int64, lastStatus int64, lastError string) error {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return fmt.Errorf("FindOne(id=%d) error: %w", id, err)
	}
	row.PageNow = pageNow
	row.LastQueryTime = lastQueryTime
	row.LastStatus = lastStatus
	row.LastError = lastError
	row.UpdatedOn = time.Now().Unix()
	return r.m.Update(ctx, row)
}

// Upsert：按 name 插入或更新；返回最终 id
func (r *SeedRepoSqlx) Upsert(ctx context.Context, s *types.Seed) (int64, error) {
	if s == nil {
		return 0, fmt.Errorf("nil seed")
	}

	// 先按 name 查
	exist, err := r.m.FindOneByName(ctx, s.Name)
	switch {
	case err == nil && exist != nil:
		// 更新
		exist.Active = s.Active
		exist.SearchType = s.SearchType
		exist.NameType = s.NameType
		exist.PageNow = s.PageNow
		exist.Offset = s.Offset
		exist.StartPage = s.StartPage
		exist.EndPage = s.EndPage
		exist.LastQueryTime = s.LastQueryTime
		exist.LastStatus = s.LastStatus
		exist.LastError = s.LastError
		exist.UpdatedOn = time.Now().Unix()
		if uerr := r.m.Update(ctx, exist); uerr != nil {
			return 0, uerr
		}
		return exist.Id, nil

	case errors.Is(sqlc.ErrNotFound, err):
		// 插入
		now := time.Now().Unix()
		row := &spiderx.DSeed{
			Name:          s.Name,
			Active:        s.Active,
			SearchType:    s.SearchType,
			NameType:      s.NameType,
			PageNow:       s.PageNow,
			Offset:        s.Offset,
			StartPage:     s.StartPage,
			EndPage:       s.EndPage,
			LastQueryTime: s.LastQueryTime,
			LastStatus:    s.LastStatus,
			LastError:     s.LastError,
			CreatedOn:     now,
			UpdatedOn:     now,
		}
		res, ierr := r.m.Insert(ctx, row)
		if ierr != nil {
			return 0, ierr
		}
		id, _ := res.LastInsertId()
		// 若驱动不支持 LastInsertId，可再按 name 读回
		if id == 0 {
			newRow, ferr := r.m.FindOneByName(ctx, s.Name)
			if ferr != nil {
				return 0, ferr
			}
			return newRow.Id, nil
		}
		return id, nil

	default:
		// 其它查询错误
		return 0, err
	}
}

func (r *SeedRepoSqlx) FindOneByName(ctx context.Context, name string) (*types.Seed, error) {
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return toSeed(row), nil
}

func toSeed(row *spiderx.DSeed) *types.Seed {
	return &types.Seed{
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
	}
}
