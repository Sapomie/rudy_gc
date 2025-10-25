package vfilm

import (
	"context"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/types"
)

func (s *DirectoryService) GetDirectoryPage(ctx context.Context, in *types.DirPageRequest) (*types.DirPageResult, error) {
	// —— 1) 定位目录（root 或 指定 id）+ 统计 & 面包屑
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

	// —— 2) 子目录（固定取一页）
	children, _, _ := s.ListChildren(ctx, dirID, in.ChildrenPage, in.ChildrenSize, film_repo.DirSortName, true, true)

	// —— 3) 影片列表（直属/递归 + 分页 + 排序）
	listReq := &types.ListDirFilmsRequest{
		DirID:     dirID,
		Page:      in.Page,
		PageSize:  in.PageSize,
		OrderBy:   in.SortField,
		Recursive: in.Recursive,
	}
	mts, stats, total, err := s.ListFilmsForDirPage(ctx, listReq)
	if err == nil {
		detail.MovieTypes = mts
		detail.Stats = stats
	}

	// —— 4) 汇总结果
	return &types.DirPageResult{
		Detail:   detail,
		Children: children,
		Total:    total,
	}, nil
}
