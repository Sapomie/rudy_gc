package sc

import (
	"context"
	"errors"
	"math/rand"
	"rudy_gc/internal/types"
)

// 单来源：从 req 指定的候选集中按规则抽取 n 个
func (l *ScService) PickFromSource(ctx context.Context, req *types.ListMovieFullRequest, n int) ([]*types.MovieType, error) {
	if n <= 0 {
		return nil, errors.New("n must be > 0")
	}
	if req == nil {
		return nil, errors.New("req is nil")
	}

	// 读取历史并构建映射
	movieLastWatch, actorLastWatch, now, err := l.loadWatchMaps(ctx)
	if err != nil {
		return nil, err
	}

	// 拉取候选与基础权重
	ensureReqDefaults(req)
	candidates, err := l.fetchCandidates(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	if n > len(candidates) {
		n = len(candidates)
	}
	baseW := computeBaseWeights(candidates, movieLastWatch, actorLastWatch, now)

	// 选择 n 个（单组）
	seed := effectiveSeed(pickCfg.RandomSeed)
	rnd := rand.New(rand.NewSource(seed))
	selected, _ := selectNFromCandidates(candidates, baseW, n, nil, nil, rnd)
	return selected, nil
}

// 从一批候选中选 n 个（用于单来源）
func selectNFromCandidates(
	cands []*types.MovieType,
	baseW []float64,
	n int,
	selectedActors map[string]struct{},
	selectedMovie map[string]struct{},
	rnd *rand.Rand,
) ([]*types.MovieType, int) {

	if selectedActors == nil {
		selectedActors = make(map[string]struct{})
	}
	if selectedMovie == nil {
		selectedMovie = make(map[string]struct{})
	}

	alive := make([]int, len(cands))
	for i := range alive {
		alive[i] = i
	}

	selected := make([]*types.MovieType, 0, n)

	for len(selected) < n && len(alive) > 0 {
		// 权重
		cur := make([]float64, len(alive))
		var sum float64
		for j, idx := range alive {
			m := cands[idx]
			if m != nil && m.JavId != "" {
				if _, ok := selectedMovie[m.JavId]; ok {
					cur[j] = 0
					continue
				}
			}
			overlap := overlapActorCount(m, selectedActors)
			diversity := 1.0 / (1.0 + pickCfg.DiversityBeta*float64(overlap))
			w := baseW[idx] * diversity
			if w < pickCfg.MinWeight {
				w = pickCfg.MinWeight
			}
			cur[j] = w
			sum += w
		}
		if sum <= 0 {
			for j := range cur {
				cur[j] = 1
			}
			sum = float64(len(cur))
		}

		// 抽 1 个
		r := rnd.Float64() * sum
		var pos int
		acc := 0.0
		for j, w := range cur {
			acc += w
			if r <= acc {
				pos = j
				break
			}
		}
		idx := alive[pos]
		picked := cands[idx]
		selected = append(selected, picked)

		if picked != nil && picked.JavId != "" {
			selectedMovie[picked.JavId] = struct{}{}
		}
		for _, c := range picked.Cast {
			name := canonicalActorName(c)
			if name != "" {
				selectedActors[name] = struct{}{}
			}
		}
		alive = append(alive[:pos], alive[pos+1:]...)
	}
	return selected, len(selected)
}

// 请求兜底
func ensureReqDefaults(req *types.ListMovieFullRequest) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 100
	}
}
