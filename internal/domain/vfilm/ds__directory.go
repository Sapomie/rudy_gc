package vfilm

import (
	"context"
	"rudy_gc/internal/domain/movie"
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
	items, _, err := s.deps.DirectoryRepo.ListRoots(ctx, 1, 1)
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

func (s *DirectoryService) ListFilmsForDirPage(
	ctx context.Context, req *types.ListDirFilmsRequest,
) (mts []*types.MovieType, stats *types.DirStats, allFilms []*types.Film, total int64, err error) {

	// 目录集合（是否递归）
	dirIDs := []int64{req.DirID}
	if req.Recursive {
		if subIDs, e := s.deps.DirectoryRepo.ListSubtreeIDs(ctx, req.DirID); e == nil && len(subIDs) > 0 {
			dirIDs = append(dirIDs, subIDs...)
		}
	}

	// 只查一次库：all = 全量，vfilms = 当前页
	all, vfilms, total, err := s.deps.FilmRepo.ListByDirectories(ctx, dirIDs, int(req.Page), int(req.PageSize), req.OrderBy)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	// 当前页 -> MovieType
	mts = make([]*types.MovieType, len(vfilms))
	for i, vf := range vfilms {
		mt, e := s.movieSvc.GetMovieType(ctx, vf.MovieJavId)
		if e != nil {
			return nil, nil, nil, 0, e
		}
		mts[i] = mt
	}

	// 顶部统计：对 all 做（递归与否由上面 dirIDs 决定）
	stats = buildDirStatsFromAll(all, req.Recursive)

	// 返回 all 给上层，避免任何重复查询
	return mts, stats, all, total, nil
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
