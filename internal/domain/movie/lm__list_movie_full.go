package movie

import (
	"context"
	"math"
	"math/rand"
	"rudy_gc/internal/types"
	"time"
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

// in: internal/service/movie_service.go  (或你现有的 service 文件里)

func (s *MovieService) ListMovieFullRandom(ctx context.Context, r *types.ListMovieFullRequest, n int64) (*types.ListMovieResponse, error) {
	if n <= 0 {
		n = 18
	}
	if n > 200 {
		n = 200
	}

	// 先探测总量：取 1 条即可拿到 total
	probe := *r
	probe.Page = 1
	probe.PageSize = 1

	_, total, err := s.deps.MovieListRepo.ListFull(ctx, &probe)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &types.ListMovieResponse{
			List:   []*types.MovieType{},
			Total:  0,
			JavIds: []string{},
		}, nil
	}

	// 如果总量不超过 n，就直接取第一页 n 条；否则随机选一“页块”
	target := *r
	target.PageSize = n

	if total <= n {
		target.Page = 1
	} else {
		// 以“页大小 = n”为步长，随机挑一个页码
		pages := int64(math.Ceil(float64(total) / float64(n)))
		if pages < 1 {
			pages = 1
		}
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		target.Page = rnd.Int63n(pages) + 1 // [1, pages]
	}

	// 复用已有聚合逻辑
	return s.ListMovieFull(ctx, &target)
}
