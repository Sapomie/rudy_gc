package wdir

import (
	"context"
	"errors"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) folderFindOneByID(ctx context.Context, id int64) (*types.Directory, error) {
	row, err := s.deps.WFolderModel.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if row == nil || row.SourceType != consts.WFolderSourceNative {
		return nil, nil
	}
	return mapFolderModelToTypes(row), nil
}

func (s *Service) folderListRoots(ctx context.Context, page, size int) ([]*types.DirSummary, int64, error) {
	return s.folderListByParent(ctx, 0, page, size)
}

func (s *Service) folderListChildren(ctx context.Context, parentID int64, page, size int) ([]*types.DirSummary, int64, error) {
	return s.folderListByParent(ctx, parentID, page, size)
}

func (s *Service) folderListByParent(ctx context.Context, parentID int64, page, size int) ([]*types.DirSummary, int64, error) {
	rows, total, err := s.deps.WFolderModel.ListByParentSourceType(ctx, parentID, consts.WFolderSourceNative, page, size)
	if err != nil {
		return nil, 0, err
	}

	out := make([]*types.DirSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, &types.DirSummary{
			Id:        row.Id,
			ParentId:  row.ParentId,
			Name:      row.Name,
			Depth:     row.Depth,
			Path:      row.Path,
			UpdatedOn: row.UpdatedOn,
		})
	}
	return out, total, nil
}

func (s *Service) folderBuildBreadcrumbs(ctx context.Context, id int64) ([]types.Breadcrumb, error) {
	cur, err := s.deps.WFolderModel.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return []types.Breadcrumb{}, nil
		}
		return nil, err
	}
	if cur == nil || cur.SourceType != consts.WFolderSourceNative {
		return []types.Breadcrumb{}, nil
	}

	seen := map[int64]struct{}{}
	out := make([]types.Breadcrumb, 0, 8)
	for cur != nil {
		if _, ok := seen[cur.Id]; ok {
			break
		}
		seen[cur.Id] = struct{}{}
		out = append(out, types.Breadcrumb{
			Id:   cur.Id,
			Name: cur.Name,
			Path: cur.Path,
		})
		if cur.ParentId <= 0 {
			break
		}

		parent, err := s.deps.WFolderModel.FindOne(ctx, cur.ParentId)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				break
			}
			return nil, err
		}
		if parent == nil || parent.SourceType != consts.WFolderSourceNative {
			break
		}
		cur = parent
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (s *Service) folderListSubtreeIDs(ctx context.Context, id int64) ([]int64, error) {
	row, err := s.deps.WFolderModel.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}
	if row == nil || row.SourceType != consts.WFolderSourceNative {
		return []int64{}, nil
	}

	ids, err := s.deps.WFolderModel.ListSubtreeIDsByPathSourceType(ctx, row.Path, consts.WFolderSourceNative)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []int64{id}, nil
	}
	return ids, nil
}

func (s *Service) mediaListByDirectories(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*types.Film, total int64, err error) {
	if len(dirIDs) == 0 {
		return []*types.Film{}, []*types.Film{}, 0, nil
	}

	rowsAll, rowsPaged, total, err := s.deps.WMediaModel.ListByDirectoryIDs(ctx, dirIDs, page, size, orderBy)
	if err != nil {
		return nil, nil, 0, err
	}

	all = make([]*types.Film, 0, len(rowsAll))
	for _, row := range rowsAll {
		all = append(all, mapMediaModelToFilmTypes(row))
	}
	paged = make([]*types.Film, 0, len(rowsPaged))
	for _, row := range rowsPaged {
		paged = append(paged, mapMediaModelToFilmTypes(row))
	}
	return all, paged, total, nil
}

func mapFolderModelToTypes(v *moviex.WFolder) *types.Directory {
	if v == nil {
		return nil
	}
	return &types.Directory{
		Id:        v.Id,
		ParentId:  v.ParentId,
		Name:      v.Name,
		Depth:     v.Depth,
		Path:      v.Path,
		CreatedOn: v.CreatedOn,
		UpdatedOn: v.UpdatedOn,
	}
}

func mapMediaModelToFilmTypes(v *moviex.WMedia) *types.Film {
	if v == nil {
		return nil
	}
	return &types.Film{
		Id:            v.Id,
		MovieJavId:    v.MovieJavId,
		MovieName:     v.MovieName,
		FileName:      v.FileName,
		DirectoryId:   v.DirectoryId,
		RootDir:       v.RootDir,
		FullDir:       v.FullDir,
		Alias:         v.Alias,
		Size:          v.Size,
		Width:         v.Width,
		Height:        v.Height,
		BitRate:       v.BitRate,
		Duration:      v.Duration,
		FrameAverage:  v.FrameAverage,
		HasSub:        v.HasSub,
		SelfMake:      v.SelfMake,
		HasMask:       v.HasMask,
		NeedScanMeta:  v.NeedScanMeta,
		IsRemoved:     v.IsRemoved,
		RemoveTime:    v.RemoveTime,
		BirthTime:     v.BirthTime,
		ReleasingDate: v.ReleasingDate,
		CreatedOn:     v.CreatedOn,
		UpdatedOn:     v.UpdatedOn,
	}
}
