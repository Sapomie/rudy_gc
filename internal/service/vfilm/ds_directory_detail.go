package vfilm

import (
	"context"
	"rudy_gc/internal/types"
)

func (s *DirectoryService) GetDirectoryPage(ctx context.Context, in *types.DirPageRequest) (*types.DirPageResult, error) {
	// 1) 定位目录
	var (
		detail *types.DirDetail
		err    error
		dirID  int64
	)
	if in.UseRoot {
		detail, err = s.GetRootDetail(ctx)
		if err != nil {
			return nil, err
		}
		if detail != nil && detail.Directory != nil {
			dirID = detail.Directory.Id
		}
	} else {
		detail, err = s.GetDirDetail(ctx, in.DirID)
		if err != nil {
			return nil, err
		}
		if detail == nil || detail.Directory == nil {
			return &types.DirPageResult{Detail: &types.DirDetail{}}, nil
		}
		dirID = detail.Directory.Id
	}

	// 2) 子目录 summary（不带聚合）
	summarys, _, _ := s.directoryListChildren(ctx, dirID, int(in.ChildrenPage), int(in.ChildrenSize))

	// 3) 影片列表（只查一次；返回 allFilms 给我们聚合子目录）
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

	// 4) 用 allFilms（一次查询的全量数据）聚合每个子目录的“直属统计”

	statsByLeaf := aggregateDirStatsByLeaf(allFilms)

	children := make([]*types.DirSummaryWithStats, 0, len(summarys))
	for _, ssum := range summarys {
		// 先用直属（叶子）统计
		ds := statsByLeaf[ssum.Id]
		if ds == nil {
			// 没有直属影片 ⇒ 做“子树聚合”
			if subIDs, e := s.directoryListSubtreeIDs(ctx, ssum.Id); e == nil && len(subIDs) > 0 {
				ds = aggregateDirStatsForIDs(allFilms, subIDs, true) // 递归子树
			}
			// 子树也没有影片 ⇒ 返回零统计
			if ds == nil {
				ds = &types.DirStats{Recursive: true}
			}
		}

		children = append(children, &types.DirSummaryWithStats{
			Summary: ssum,
			Stats:   []*types.DirStats{ds},
		})
	}

	return &types.DirPageResult{
		Detail:   detail,
		Children: children, // DirSummaryWithStats
		Total:    total,
	}, nil
}

// 叶子层“直属统计”：按影片的 DirectoryId 分组
func aggregateDirStatsByLeaf(all []*types.Film) map[int64]*types.DirStats {
	res := make(map[int64]*types.DirStats, 64)
	for _, f := range all {
		if f == nil {
			continue
		}
		ds := res[f.DirectoryId]
		if ds == nil {
			ds = &types.DirStats{Recursive: false}
			res[f.DirectoryId] = ds
		}
		ds.FilmCount++
		ds.TotalSize += f.Size
		if f.BirthTime > ds.LastFilmBirth {
			ds.LastFilmBirth = f.BirthTime
		}
		if f.UpdatedOn > ds.LastUpdatedOn {
			ds.LastUpdatedOn = f.UpdatedOn
		}
	}
	return res
}

// 给定一组目录ID（通常是“某子目录的整棵子树”的所有叶子/目录ID），对 allFilms 做聚合
func aggregateDirStatsForIDs(all []*types.Film, ids []int64, recursive bool) *types.DirStats {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	var ds *types.DirStats
	for _, f := range all {
		if f == nil {
			continue
		}
		if _, ok := set[f.DirectoryId]; !ok {
			continue
		}
		if ds == nil {
			ds = &types.DirStats{Recursive: recursive}
		}
		ds.FilmCount++
		ds.TotalSize += f.Size
		if f.BirthTime > ds.LastFilmBirth {
			ds.LastFilmBirth = f.BirthTime
		}
		if f.UpdatedOn > ds.LastUpdatedOn {
			ds.LastUpdatedOn = f.UpdatedOn
		}
	}
	return ds
}
