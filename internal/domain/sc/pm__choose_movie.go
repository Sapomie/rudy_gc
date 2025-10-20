package sc

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"rudy_gc/internal/types"
	"sort"
	"time"
)

// ================= 可调参数 =================

type PickConfig struct {
	// 候选集请求参数
	PageSize int64 // ListMovieFull 的 PageSize（默认 100000）
	Owned    int64 // 片源过滤（默认 3）

	// 多样性：减少演员重复（越大惩罚越强）
	DiversityBeta float64 // 默认 0.75

	// 观看历史惩罚（仅演员保留软惩罚）
	ActorPenaltyAlpha float64 // 最近看过演员惩罚强度（0~1，默认 0.6）

	// “近期”的时间尺度（半衰期，单位：天；越小越看重最近）
	HalfLifeMovieDays float64 // 电影用作硬阈值参考（BlockMovieDays），保留字段以备扩展（默认 14）
	HalfLifeActorDays float64 // 演员软惩罚的半衰期（默认 7）

	// 随机数种子（>0 可复现；<=0 使用当前时间）
	RandomSeed int64 // 默认 0

	// 权重下限（避免为负或过小）
	MinWeight float64 // 默认 0

	// 硬阈值：直接屏蔽最近看过
	BlockMovieDays int // 最近 X 天看过的电影直接权重=0（默认 0 表示不开）
	BlockActorDays int // 最近 X 天看过的演员出场则权重=0（默认 0 表示不开）
}

func DefaultPickConfig() PickConfig {
	return PickConfig{
		PageSize:          100000,
		Owned:             3,
		DiversityBeta:     100,
		ActorPenaltyAlpha: 0.6,
		HalfLifeMovieDays: 14,
		HalfLifeActorDays: 7,
		RandomSeed:        0,
		MinWeight:         0,
		BlockMovieDays:    0,
		BlockActorDays:    30,
	}
}

// 全局配置（可在外部初始化时覆盖）
var pickCfg = DefaultPickConfig()

// 可选：在启动时调用覆盖配置
func SetPickConfig(c PickConfig) {
	pickCfg = c
}

// ===========================================

type scMovies struct {
	MovieType   *types.MovieType
	WatchedTime int64
}

type requestWithWeight struct {
	req *types.ListMovieFullRequest
	w   int64 //权重
}

// -------- 外部主函数（多请求按比例抽取）：从多个 reqs 中按权重比重随机选 n 个 --------
func (l *ScService) PickMovie(ctx context.Context, reqs []*requestWithWeight, n int) ([]*types.MovieType, error) {
	if n <= 0 {
		return nil, errors.New("n must be > 0")
	}
	if len(reqs) == 0 {
		return nil, errors.New("reqs is empty")
	}

	// 1) 读取观影历史（只做一次）
	history, err := l.allScMovies(ctx)
	if err != nil {
		return nil, err
	}
	movieLastWatch, actorLastWatch := buildWatchMaps(history)
	now := time.Now()

	// 2) 拉取候选集 + 计算基础权重
	type group struct {
		req        *types.ListMovieFullRequest
		weight     float64
		candidates []*types.MovieType
		baseW      []float64 // 每个电影的基础权重
		alive      []int     // 活跃索引
		target     int       // 该组分配的抽取数量
	}

	groups := make([]*group, 0, len(reqs))
	totalWeight := 0.0

	for _, rw := range reqs {
		if rw == nil || rw.req == nil {
			continue
		}
		if rw.w <= 0 {
			rw.w = 1
		}
		req := rw.req
		if req.Page == 0 {
			req.Page = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 100
		}

		resp, err := l.movieSvc.ListMovieFull(ctx, req)
		if err != nil {
			return nil, err
		}
		cands := resp.List
		l.deps.Log.Info("找到movies  Number---------", len(resp.List))
		if len(cands) == 0 {
			continue
		}

		base := make([]float64, len(cands))
		for j, m := range cands {
			w := scoreMovie(m, movieLastWatch, actorLastWatch, now)
			if w < pickCfg.MinWeight {
				w = pickCfg.MinWeight
			}
			base[j] = w
		}
		alive := make([]int, len(cands))
		for j := range alive {
			alive[j] = j
		}

		g := &group{
			req:        req,
			weight:     float64(rw.w),
			candidates: cands,
			baseW:      base,
			alive:      alive,
		}
		groups = append(groups, g)
		totalWeight += g.weight
	}

	if len(groups) == 0 {
		return nil, nil
	}

	// 3) 按权重分配各组抽取份额
	alloc := make([]int, len(groups))
	if totalWeight <= 0 {
		for i := range groups {
			alloc[i] = n / len(groups)
		}
		rem := n % len(groups)
		for i := 0; i < rem; i++ {
			alloc[i]++
		}
	} else {
		type frac struct {
			i    int
			part float64
		}
		fracs := make([]frac, 0, len(groups))
		sum := 0
		for i, g := range groups {
			exact := float64(n) * g.weight / totalWeight
			floor := int(math.Floor(exact))
			alloc[i] = floor
			sum += floor
			fracs = append(fracs, frac{i: i, part: exact - float64(floor)})
		}
		rem := n - sum
		sort.Slice(fracs, func(a, b int) bool { return fracs[a].part > fracs[b].part })
		for i := 0; i < rem && i < len(fracs); i++ {
			alloc[fracs[i].i]++
		}
	}
	for i := range groups {
		groups[i].target = alloc[i]
	}

	// 4) 全局随机选择 + 多样性去重
	selected := make([]*types.MovieType, 0, n)
	selectedActors := make(map[string]struct{})
	selectedMovie := make(map[string]struct{})

	seed := pickCfg.RandomSeed
	if seed <= 0 {
		seed = time.Now().UnixNano()
	}
	rnd := rand.New(rand.NewSource(seed))

	// 轮询式多组抽取
	for {
		progress := false
		for _, g := range groups {
			if g.target <= 0 || len(g.alive) == 0 {
				continue
			}

			curWeights := make([]float64, len(g.alive))
			var sum float64
			for j, idx := range g.alive {
				m := g.candidates[idx]
				// 避免重复电影
				if _, ok := selectedMovie[m.JavId]; ok {
					curWeights[j] = 0
					continue
				}
				overlap := overlapActorCount(m, selectedActors)
				diversityFactor := 1.0 / (1.0 + pickCfg.DiversityBeta*float64(overlap))
				w := g.baseW[idx] * diversityFactor
				if w < pickCfg.MinWeight {
					w = pickCfg.MinWeight
				}
				curWeights[j] = w
				sum += w
			}
			if sum <= 0 {
				for j := range curWeights {
					curWeights[j] = 1
				}
				sum = float64(len(curWeights))
			}

			r := rnd.Float64() * sum
			var pickPos int
			acc := 0.0
			for j, w := range curWeights {
				acc += w
				if r <= acc {
					pickPos = j
					break
				}
			}

			pickIdx := g.alive[pickPos]
			picked := g.candidates[pickIdx]
			selected = append(selected, picked)
			selectedMovie[picked.JavId] = struct{}{}
			for _, c := range picked.Cast {
				name := canonicalActorName(c)
				if name != "" {
					selectedActors[name] = struct{}{}
				}
			}

			// 无放回
			g.alive = append(g.alive[:pickPos], g.alive[pickPos+1:]...)
			g.target--
			progress = true

			if len(selected) >= n {
				return selected, nil
			}
		}
		if !progress {
			break
		}
	}

	return selected, nil
}

// 单请求：从 req 指定的候选集中按规则抽取 n 个
func (l *ScService) PickMovieOnce(ctx context.Context, req *types.ListMovieFullRequest, n int) ([]*types.MovieType, error) {
	if n <= 0 {
		return nil, errors.New("n must be > 0")
	}
	if req == nil {
		return nil, errors.New("req is nil")
	}

	// 读取观影历史并构建“最近观看映射”
	history, err := l.allScMovies(ctx)
	if err != nil {
		return nil, err
	}
	movieLastWatch, actorLastWatch := buildWatchMaps(history)
	now := time.Now()

	// 请求兜底
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 100
	}

	// 拉取候选集
	resp, err := l.movieSvc.ListMovieFull(ctx, req)
	if err != nil {
		return nil, err
	}
	candidates := resp.List
	if len(candidates) == 0 {
		return nil, nil
	}
	if n > len(candidates) {
		n = len(candidates)
	}

	// 计算基础权重
	baseWeights := make([]float64, len(candidates))
	for i, m := range candidates {
		w := scoreMovie(m, movieLastWatch, actorLastWatch, now)
		if w < pickCfg.MinWeight {
			w = pickCfg.MinWeight
		}
		baseWeights[i] = w
	}

	// 选择（组内多样性、无放回；仅保证本次选择内不重复）
	selected := make([]*types.MovieType, 0, n)
	selectedActors := make(map[string]struct{})
	selectedMovie := make(map[string]struct{})

	seed := pickCfg.RandomSeed
	if seed <= 0 {
		seed = time.Now().UnixNano()
	}
	rnd := rand.New(rand.NewSource(seed))

	alive := make([]int, len(candidates))
	for i := range alive {
		alive[i] = i
	}

	for len(selected) < n && len(alive) > 0 {
		// 当前权重 = 基础权重 × 多样性因子
		curWeights := make([]float64, len(alive))
		var sum float64
		for j, idx := range alive {
			m := candidates[idx]
			// 避免同一次结果内的电影重复（同 JavId）
			if m != nil && m.JavId != "" {
				if _, ok := selectedMovie[m.JavId]; ok {
					curWeights[j] = 0
					continue
				}
			}
			overlap := overlapActorCount(m, selectedActors)
			diversityFactor := 1.0 / (1.0 + pickCfg.DiversityBeta*float64(overlap))
			w := baseWeights[idx] * diversityFactor
			if w < pickCfg.MinWeight {
				w = pickCfg.MinWeight
			}
			curWeights[j] = w
			sum += w
		}

		// 权重全为 0 时退化为等概率
		if sum <= 0 {
			for j := range curWeights {
				curWeights[j] = 1
			}
			sum = float64(len(curWeights))
		}

		// 加权随机
		r := rnd.Float64() * sum
		var pickPos int
		acc := 0.0
		for j, w := range curWeights {
			acc += w
			if r <= acc {
				pickPos = j
				break
			}
		}

		// 记录与无放回
		pickIdx := alive[pickPos]
		picked := candidates[pickIdx]
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
		alive = append(alive[:pickPos], alive[pickPos+1:]...)
	}

	return selected, nil
}

// 你已有声明：在别处实现实际读取逻辑
func (l *ScService) allScMovies(ctx context.Context) ([]*scMovies, error) {
	glists, err := l.deps.GListRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	scs := make([]*scMovies, 0, len(glists))
	for _, gl := range glists {
		movieType, err := l.movieSvc.GetMovieType(ctx, gl.MovieJavId)
		if err != nil {
			return nil, err
		}
		gSc, err := l.deps.ScRepo.FindOneByName(ctx, gl.ScName)
		if err != nil {
			return nil, err
		}
		scMovie := &scMovies{
			MovieType:   movieType,
			WatchedTime: gSc.ScTime,
		}
		scs = append(scs, scMovie)
	}

	return scs, nil
}

// ======= 评分与惩罚相关 =======

// 为电影打分（越大越容易被选中）
// 注意：已移除“电影近期软惩罚”，仅保留：
//   - 硬阈值 BlockMovieDays / BlockActorDays
//   - 演员近期软惩罚（ActorPenaltyAlpha * exp(-ageDays/HalfLifeActorDays))
func scoreMovie(m *types.MovieType, movieLast map[string]int64, actorLast map[string]int64, now time.Time) float64 {
	if m == nil {
		return 0
	}

	// --------- 硬阈值（先挡掉“太近”的）---------
	// 电影硬阈值
	if pickCfg.BlockMovieDays > 0 {
		if last, ok := movieLast[m.JavId]; ok && last > 0 {
			if ageInDays(last, now) < float64(pickCfg.BlockMovieDays) {
				return 0
			}
		}
	}
	// 演员硬阈值：只要片中任一演员在 X 天内出现过，直接 0
	if pickCfg.BlockActorDays > 0 {
		for _, c := range m.Cast {
			name := canonicalActorName(c)
			if name == "" {
				continue
			}
			if last, ok := actorLast[name]; ok && last > 0 {
				if ageInDays(last, now) < float64(pickCfg.BlockActorDays) {
					return 0
				}
			}
		}
	}

	// --------- 软惩罚（仅对演员）---------
	score := 1.0

	// 演员近期惩罚（取本片演员中 recentness 最大的那位）
	var maxActorRecentness float64
	for _, c := range m.Cast {
		name := canonicalActorName(c)
		if name == "" {
			continue
		}
		if last, ok := actorLast[name]; ok && last > 0 {
			ageDays := ageInDays(last, now)
			r := math.Exp(-ageDays / pickCfg.HalfLifeActorDays)
			if r > maxActorRecentness {
				maxActorRecentness = r
			}
		}
	}
	if maxActorRecentness > 0 {
		score *= (1.0 - pickCfg.ActorPenaltyAlpha*maxActorRecentness)
	}

	if score < pickCfg.MinWeight {
		return pickCfg.MinWeight
	}
	return score
}

// 统计这部电影与“已选演员集合”的重叠数
func overlapActorCount(m *types.MovieType, selectedActors map[string]struct{}) int {
	if len(selectedActors) == 0 || m == nil || len(m.Cast) == 0 {
		return 0
	}
	cnt := 0
	for _, c := range m.Cast {
		name := canonicalActorName(c)
		if name == "" {
			continue
		}
		if _, ok := selectedActors[name]; ok {
			cnt++
		}
	}
	return cnt
}

// 从观影历史构建：电影 -> 最近观看时间；演员 -> 最近观看时间（取 max）
func buildWatchMaps(history []*scMovies) (map[string]int64, map[string]int64) {
	movieLast := make(map[string]int64, len(history))
	actorLast := make(map[string]int64)

	for _, h := range history {
		if h == nil || h.MovieType == nil {
			continue
		}
		mv := h.MovieType
		wt := h.WatchedTime

		// 电影
		if mv.JavId != "" {
			if old, ok := movieLast[mv.JavId]; !ok || wt > old {
				movieLast[mv.JavId] = wt
			}
		}

		// 演员（以该电影的观看时间作为“看过该演员”的时间）
		for _, c := range mv.Cast {
			name := canonicalActorName(c)
			if name == "" {
				continue
			}
			if old, ok := actorLast[name]; !ok || wt > old {
				actorLast[name] = wt
			}
		}
	}
	return movieLast, actorLast
}

func ageInDays(ts int64, now time.Time) float64 {
	if ts <= 0 {
		return 36500 // 非法时间就当很久以前
	}
	age := now.Sub(time.Unix(ts, 0))
	if age < 0 {
		age = 0
	}
	return age.Hours() / 24.0
}

func canonicalActorName(c *types.CastInfo) string {
	if c == nil {
		return ""
	}
	// 统一用 Name，有需要可回退 NameShow
	if c.Name != "" {
		return c.Name
	}
	return c.NameShow
}
