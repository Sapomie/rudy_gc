package movie

import (
	"context"
	"rudy_gc/internal/types"
)

func (s *MovieService) ListMovieFull(ctx context.Context, r *types.ListMovieFullRequest) (*types.ListMovieResponse, error) {

	// 交给 Repo 做：分表筛选 -> 交集 -> 指定表排序分页 -> 返回 javId 列表 + total
	rows, total, err := s.deps.MovieListRepo.ListFull(ctx, r)
	if err != nil {
		return nil, err
	}

	// 聚合 MovieType（走你已有缓存/聚合链路）
	out := make([]*types.MovieType, 0, len(rows))
	javIds := make([]string, 0, len(rows))
	for _, mv := range rows {
		mt, err := s.GetMovieType(ctx, mv.JavId)
		if err != nil {
			return nil, err
		}
		out = append(out, mt)
		javIds = append(javIds, mv.JavId)
	}

	return &types.ListMovieResponse{
		List:   out,
		Total:  total,
		JavIds: javIds,
	}, nil
}
