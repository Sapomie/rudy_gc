package vfilm

import (
	"context"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func (s *DirectoryService) ListRootPage(ctx context.Context, page, pageSize int64) ([]*types.DirSummaryWithStats, int64, error) {
	summarys, total, err := s.directoryListRoots(ctx, int(page), int(pageSize))
	if err != nil {
		return nil, 0, err
	}

	items := make([]*types.DirSummaryWithStats, 0, len(summarys))
	for _, summary := range summarys {
		stats, err := s.buildDirectoryRecursiveStats(ctx, summary.Id)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, &types.DirSummaryWithStats{
			Summary: summary,
			Stats:   []*types.DirStats{stats},
		})
	}
	return items, total, nil
}

func (s *DirectoryService) GetDirDetail(ctx context.Context, id int64) (*types.DirDetail, error) {
	dir, err := s.directoryFindOneByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, nil
	}
	crumbs, err := s.directoryBuildBreadcrumbs(ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.DirDetail{Directory: dir, Breadcrumbs: crumbs}, nil
}

func (s *DirectoryService) ListFilmsForDirPage(
	ctx context.Context, req *types.ListDirFilmsRequest,
) (mts []*types.MovieType, stats *types.DirStats, allFilms []*types.Film, total int64, err error) {

	dirIDs := []int64{req.DirID}
	if req.Recursive {
		if subIDs, e := s.directoryListSubtreeIDs(ctx, req.DirID); e == nil && len(subIDs) > 0 {
			dirIDs = append(dirIDs, subIDs...)
		}
	}

	all, vfilms, total, err := s.filmListByDirectories(ctx, dirIDs, int(req.Page), int(req.PageSize), req.OrderBy)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	mts = make([]*types.MovieType, len(vfilms))
	for i, vf := range vfilms {
		mt, e := s.movieSvc.GetMovieType(ctx, vf.MovieJavId)
		if e != nil {
			return nil, nil, nil, 0, e
		}
		mts[i] = mt
	}

	stats = buildDirStatsFromAll(all, req.Recursive)
	return mts, stats, all, total, nil
}

func (s *DirectoryService) buildDirectoryRecursiveStats(ctx context.Context, dirID int64) (*types.DirStats, error) {
	dirIDs, err := s.directoryListSubtreeIDs(ctx, dirID)
	if err != nil {
		return nil, err
	}
	if len(dirIDs) == 0 {
		dirIDs = []int64{dirID}
	}

	allFilms, _, _, err := s.filmListByDirectories(ctx, dirIDs, 1, 1, consts.OrderByBirthTime)
	if err != nil {
		return nil, err
	}
	return buildDirStatsFromAll(allFilms, true), nil
}

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
	return st
}
