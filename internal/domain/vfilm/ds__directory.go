package vfilm

import (
	"context"
	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type DirectoryService struct {
	deps     *svc.Deps
	movieSvc *movie.MovieService
}

func NewDirectoryService(deps *svc.Deps) *DirectoryService {
	return &DirectoryService{
		deps:     deps,
		movieSvc: movie.NewMovieService(deps),
	}
}

// 单根场景：直接返回根目录详情（可递归统计）
func (s *DirectoryService) GetRootDetail(ctx context.Context) (*types.DirDetail, error) {
	// 若根ID固定，可以直接 FindOneByID(ctx, 1)
	items, _, err := s.deps.DirectoryRepo.ListRoots(ctx, 1, 1, film_repo.DirSortName, true, true)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &types.DirDetail{}, nil
	}
	root, err := s.deps.DirectoryRepo.FindOneByID(ctx, items[0].Id)
	if err != nil {
		return nil, err
	}
	return &types.DirDetail{
		Directory:   root,
		Breadcrumbs: []types.Breadcrumb{}, // 根无上级
	}, nil
}

// 任意目录详情（含面包屑 + 可选递归统计）
func (s *DirectoryService) GetDirDetail(ctx context.Context, id int64) (*types.DirDetail, error) {
	dir, err := s.deps.DirectoryRepo.FindOneByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, nil
	}
	crumbs, _ := s.deps.DirectoryRepo.BuildBreadcrumbs(ctx, id)
	return &types.DirDetail{Directory: dir, Breadcrumbs: crumbs}, nil
}

// 子目录列表（支持分页/排序/是否聚合）
func (s *DirectoryService) ListChildren(ctx context.Context, parentID int64, page, size int64, sort film_repo.DirSort, asc bool, withAgg bool) ([]*types.DirSummary, int64, error) {
	return s.deps.DirectoryRepo.ListChildren(ctx, parentID, int(page), int(size), sort, asc, withAgg)
}

func (s *DirectoryService) ListFilmsForDirPage(ctx context.Context, req *types.ListDirFilmsRequest) ([]*types.MovieType, *types.DirStats, int64, error) {
	// 1) 目录ID集合
	dirIDs := []int64{req.DirID}
	if req.Recursive {
		if subIDs, err := s.deps.DirectoryRepo.ListSubtreeIDs(ctx, req.DirID); err == nil && len(subIDs) > 0 {
			dirIDs = append(dirIDs, subIDs...)
		}
	}

	// 2) 查询：all = 全量，list = 当前页
	all, list, total, err := s.deps.FilmRepo.ListByDirectories(ctx, dirIDs, int(req.Page), int(req.PageSize), req.OrderBy)
	if err != nil {
		return nil, nil, 0, err
	}

	// 3) 当前页转 MovieType（按你现有逻辑）
	mts := make([]*types.MovieType, len(list))
	for i, vf := range list {
		mt, err := s.movieSvc.GetMovieType(ctx, vf.MovieJavId)
		if err != nil {
			return nil, nil, 0, err
		}
		mts[i] = mt
	}

	// 4) 基于 all 聚合目录统计
	stats := buildDirStatsFromAll(all, req.Recursive)

	return mts, stats, total, nil
}

// 基于全量影片聚合目录统计
func buildDirStatsFromAll(all []*types.Film, recursive bool) *types.DirStats {
	st := &types.DirStats{
		Recursive: recursive,
	}
	if len(all) == 0 {
		return st
	}
	var (
		sumSize      int64
		maxBirthTime int64
		maxUpdatedOn int64
	)
	for _, f := range all {
		sumSize += f.Size
		if f.BirthTime > maxBirthTime {
			maxBirthTime = f.BirthTime
		}
		if f.UpdatedOn > maxUpdatedOn {
			maxUpdatedOn = f.UpdatedOn
		}
	}
	st.FilmCount = int64(len(all))
	st.TotalSize = sumSize
	st.LastFilmBirth = maxBirthTime
	st.LastUpdatedOn = maxUpdatedOn
	// 如果后续需要时间分桶（Buckets），在这里追加即可
	return st
}

// 同级目录（便于前端做快速切换）
func (s *DirectoryService) ListSiblings(ctx context.Context, id int64) ([]*types.DirSummary, error) {
	return s.deps.DirectoryRepo.ListSiblings(ctx, id)
}

// 面包屑
func (s *DirectoryService) GetBreadcrumbs(ctx context.Context, id int64) ([]types.Breadcrumb, error) {
	return s.deps.DirectoryRepo.BuildBreadcrumbs(ctx, id)
}
