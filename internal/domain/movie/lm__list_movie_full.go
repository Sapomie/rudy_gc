package movie

import (
	"context"
	"math/rand"
	"rudy_gc/internal/types"
	"sort"
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

	// 1) 先探测总量
	probe := *r
	probe.Page, probe.PageSize = 1, 1

	_, total, err := s.deps.MovieListRepo.ListFull(ctx, &probe)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &types.ListMovieResponse{List: []*types.MovieType{}, Total: 0, JavIds: []string{}}, nil
	}
	if total <= n {
		// 候选不足 n：直接取前 n 条（保持与原逻辑一致）
		target := *r
		target.Page, target.PageSize = 1, n
		return s.ListMovieFull(ctx, &target)
	}

	// 2) 在 [1..total] 中随机选择 n 个不重复位置
	//    用 map 去重，再排序
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	posSet := make(map[int64]struct{}, n)
	for int64(len(posSet)) < n {
		// positions are 1-based
		p := rng.Int63n(total) + 1
		posSet[p] = struct{}{}
	}
	positions := make([]int64, 0, n)
	for p := range posSet {
		positions = append(positions, p)
	}
	sort.Slice(positions, func(i, j int) bool { return positions[i] < positions[j] })

	// 3) 把离散位置合并为少量“连续区间”
	type seg struct{ start, length int64 } // 1-based page, length items
	segs := make([]seg, 0, n)
	start := positions[0]
	prev := positions[0]
	for i := 1; i < len(positions); i++ {
		if positions[i] == prev+1 {
			prev = positions[i]
			continue
		}
		segs = append(segs, seg{start: start, length: prev - start + 1})
		start, prev = positions[i], positions[i]
	}
	segs = append(segs, seg{start: start, length: prev - start + 1})

	// 4) 分批查询（每段 1 次），聚合结果
	//    这里利用已有 ListMovieFull 的分页语义：Page = start, PageSize = length
	allItems := make([]*types.MovieType, 0, n)
	allIDs := make([]string, 0, n)
	collected := int64(0)

	for _, g := range segs {
		if collected >= n {
			break
		}
		// 对于超大 total，极端情况下段会很散；这里每段只取需要的长度，不额外放大
		target := *r
		target.Page = g.start
		target.PageSize = g.length

		resp, err := s.ListMovieFull(ctx, &target)
		if err != nil {
			return nil, err
		}
		if len(resp.List) == 0 {
			continue
		}
		for i := 0; i < len(resp.List) && collected < n; i++ {
			allItems = append(allItems, resp.List[i])
			allIDs = append(allIDs, resp.JavIds[i])
			collected++
		}
	}

	// 5) 打散顺序（避免还带有原排序的“局部性”）
	rng.Shuffle(len(allItems), func(i, j int) {
		allItems[i], allItems[j] = allItems[j], allItems[i]
		allIDs[i], allIDs[j] = allIDs[j], allIDs[i]
	})

	return &types.ListMovieResponse{
		List:   allItems,
		Total:  total, // 返回候选全集大小
		JavIds: allIDs,
	}, nil
}
