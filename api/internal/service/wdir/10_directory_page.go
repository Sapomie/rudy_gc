package wdir

import (
	"context"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/repository/moviex"
	mediaservice "rudy_gc/internal/service/media"
	"rudy_gc/internal/types"
)

func (s *DirectoryService) ListRootPage(ctx context.Context, page, pageSize int64) ([]*types.DirSummaryWithStats, int64, error) {
	if err := s.ensureFolderTreeNormalized(ctx); err != nil {
		return nil, 0, err
	}

	summarys, total, err := s.folderListRoots(ctx, int(page), int(pageSize))
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
	if err := s.ensureFolderTreeNormalized(ctx); err != nil {
		return nil, err
	}

	dir, err := s.folderFindOneByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, nil
	}

	crumbs, err := s.folderBuildBreadcrumbs(ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.DirDetail{
		Directory:   dir,
		Breadcrumbs: crumbs,
	}, nil
}

func (s *DirectoryService) ListFilmsForDirPage(ctx context.Context, req *types.ListDirFilmsRequest) (mts []*types.MovieType, stats *types.DirStats, allFilms []*types.Film, total int64, err error) {
	dirIDs, err := s.directoryScopeIDs(ctx, req.DirID, req.Recursive)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	all, mediaRows, total, err := s.mediaListByDirectories(ctx, dirIDs, int(req.Page), int(req.PageSize), req.OrderBy)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	mts = make([]*types.MovieType, 0, len(mediaRows))
	for _, media := range mediaRows {
		mt, e := s.movieSvc.GetMovieType(ctx, media.MovieJavId)
		if e != nil {
			s.deps.Log.Warnf("wdir skip movie type, jav_id=%s, err=%v", media.MovieJavId, e)
			continue
		}
		mts = append(mts, mt)
	}

	stats = buildDirStatsFromAll(all, req.Recursive)
	return mts, stats, all, total, nil
}

func (s *DirectoryService) ListFC2CardsForDirPage(ctx context.Context, req *types.ListDirFilmsRequest) ([]*mediaservice.FC2CardView, int64, error) {
	dirIDs, err := s.directoryScopeIDs(ctx, req.DirID, req.Recursive)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.deps.FC2MediaModel.CountActiveByDirectoryIDs(ctx, dirIDs)
	if err != nil {
		return nil, 0, err
	}
	offset := (req.Page - 1) * req.PageSize
	mediaRows, err := s.deps.FC2MediaModel.ListPageByDirectoryIDs(ctx, dirIDs, offset, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	movieIDs := make([]int64, 0, len(mediaRows))
	for _, row := range mediaRows {
		if row != nil && row.FC2MovieID > 0 {
			movieIDs = append(movieIDs, row.FC2MovieID)
		}
	}
	movieRows, err := s.deps.FC2MovieModel.ListByIDs(ctx, movieIDs)
	if err != nil {
		return nil, 0, err
	}
	movieByID := make(map[int64]*moviex.FC2Movie, len(movieRows))
	for _, row := range movieRows {
		if row != nil {
			movieByID[row.ID] = row
		}
	}
	orderedRows := make([]*moviex.FC2Movie, 0, len(mediaRows))
	for _, mediaRow := range mediaRows {
		if mediaRow == nil {
			continue
		}
		if movieRow := movieByID[mediaRow.FC2MovieID]; movieRow != nil {
			orderedRows = append(orderedRows, movieRow)
		}
	}
	cards, err := s.mediaSvc.BuildFC2CardViews(ctx, orderedRows)
	if err != nil {
		return nil, 0, err
	}
	return cards, total, nil
}

func (s *DirectoryService) directoryScopeIDs(ctx context.Context, dirID int64, recursive bool) ([]int64, error) {
	if !recursive {
		return []int64{dirID}, nil
	}
	dirIDs, err := s.folderListSubtreeIDs(ctx, dirID)
	if err != nil {
		return nil, err
	}
	if len(dirIDs) == 0 {
		return []int64{dirID}, nil
	}
	return dirIDs, nil
}

func (s *DirectoryService) GetDirectoryPage(ctx context.Context, in *types.DirPageRequest) (*types.DirPageResult, error) {
	detail, err := s.GetDirDetail(ctx, in.DirID)
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.Directory == nil {
		return &types.DirPageResult{Detail: &types.DirDetail{}}, nil
	}
	dirID := detail.Directory.Id

	summarys, _, err := s.folderListChildren(ctx, dirID, int(in.ChildrenPage), int(in.ChildrenSize))
	if err != nil {
		return nil, err
	}

	listReq := &types.ListDirFilmsRequest{
		DirID:     dirID,
		Page:      in.Page,
		PageSize:  in.PageSize,
		OrderBy:   in.SortField,
		Recursive: in.Recursive,
	}
	mts, stats, allFilms, total, err := s.ListFilmsForDirPage(ctx, listReq)
	if err == nil {
		detail.MovieTypes = mts
		detail.Stats = stats
	}

	statsByLeaf := aggregateDirStatsByLeaf(allFilms)
	children := make([]*types.DirSummaryWithStats, 0, len(summarys))
	for _, summary := range summarys {
		ds := statsByLeaf[summary.Id]
		if ds == nil {
			if subIDs, e := s.folderListSubtreeIDs(ctx, summary.Id); e == nil && len(subIDs) > 0 {
				ds = aggregateDirStatsForIDs(allFilms, subIDs, true)
			}
			if ds == nil {
				ds = newDirStats(true)
			}
		}
		if ds.FilmCount <= 0 {
			continue
		}

		children = append(children, &types.DirSummaryWithStats{
			Summary: summary,
			Stats:   []*types.DirStats{ds},
		})
	}

	return &types.DirPageResult{
		Detail:   detail,
		Children: children,
		Total:    total,
	}, nil
}

func (s *DirectoryService) buildDirectoryRecursiveStats(ctx context.Context, dirID int64) (*types.DirStats, error) {
	dirIDs, err := s.folderListSubtreeIDs(ctx, dirID)
	if err != nil {
		return nil, err
	}
	if len(dirIDs) == 0 {
		dirIDs = []int64{dirID}
	}

	allFilms, _, _, err := s.mediaListByDirectories(ctx, dirIDs, 1, 1, consts.OrderByMediaBirthTime)
	if err != nil {
		return nil, err
	}
	return buildDirStatsFromAll(allFilms, true), nil
}
