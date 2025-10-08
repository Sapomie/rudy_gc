package movie

import (
	"context"
	"errors"
	"rudy_gc/internal/types"
)

func (s *Service) ListMovieLite(ctx context.Context, r *types.ListMovieLiteRequest) (*types.ListMovieLiteResponse, error) {
	if r == nil {
		return nil, errors.New("nil ListMovieLiteRequest")
	}

	// page/size/orderBy 由上层（router）决定，这里只做位移计算
	offset := (r.Page - 1) * r.PageSize
	limit := r.PageSize
	orderKey := r.OrderBy

	// 分页 + 总数：在 repo 中完成（后续加筛选条件时，同步到 COUNT 与 SELECT）
	rows, total, err := s.deps.MovieRepo.ListPage(ctx, offset, limit, orderKey)
	if err != nil {
		return nil, err
	}

	// 组装 MovieType（需要完整聚合就用现有的 buildMovieTypeFromRepos）
	out := make([]*types.MovieType, 0, len(rows))
	for _, mv := range rows {
		mt, err := s.GetMovieType(ctx, mv.JavId)
		if err != nil {
			return nil, err
		}
		out = append(out, mt)
	}

	return &types.ListMovieLiteResponse{
		List:  out,
		Total: total,
	}, nil
}
