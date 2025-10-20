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

//
// ======================= 配置（可调） =======================
//

type PickConfig struct {
	// 候选集请求参数
	PageSize int64
	Owned    int64

	// 多样性：减少演员重复（越大惩罚越强）
	DiversityBeta float64

	// 观看历史惩罚（仅演员保留软惩罚）
	ActorPenaltyAlpha float64

	// 半衰期（天）
	HalfLifeActorDays float64

	// 随机数种子（>0 可复现；<=0 使用当前时间）
	RandomSeed int64

	// 权重下限（避免为负或过小）
	MinWeight float64

	// 硬阈值：直接屏蔽最近看过
	BlockMovieDays int
	BlockActorDays int
}

func DefaultPickConfig() PickConfig {
	return PickConfig{
		PageSize:          100000,
		Owned:             3,
		DiversityBeta:     100, // 你当前使用的默认
		ActorPenaltyAlpha: 0.6,
		HalfLifeActorDays: 45,
		BlockActorDays:    30,
		RandomSeed:        0,
		MinWeight:         0,
		BlockMovieDays:    0,
	}
}

// 全局配置（可在外部初始化时覆盖）
var pickCfg = DefaultPickConfig()

func SetPickConfig(c PickConfig) { pickCfg = c }

//
// ======================= 对外类型 =======================
//

type requestWithWeight struct {
	req *types.ListMovieFullRequest
	w   int64 // 权重
}

// 统一的“组”类型，避免匿名结构体造成的类型不匹配
type pickGroup struct {
	req        *types.ListMovieFullRequest
	weight     float64
	candidates []*types.MovieType
	baseW      []float64
	alive      []int // 活跃索引
	target     int   // 分配的名额
}

// 多来源：从多个 reqs 中按权重比例抽取 n 个（全局共享多样性/去重）
func (l *ScService) PickFromSources(ctx context.Context, reqs []*requestWithWeight, n int) ([]*types.MovieType, error) {
	if n <= 0 {
		return nil, errors.New("n must be > 0")
	}
	if len(reqs) == 0 {
		return nil, errors.New("reqs is empty")
	}

	// 读取历史并构建映射
	movieLastWatch, actorLastWatch, now, err := l.loadWatchMaps(ctx)
	if err != nil {
		return nil, err
	}

	// 组装各组
	groups := make([]*pickGroup, 0, len(reqs))
	totalWeight := 0.0

	for _, rw := range reqs {
		if rw == nil || rw.req == nil {
			continue
		}
		if rw.w <= 0 {
			rw.w = 1
		}
		req := rw.req
		ensureReqDefaults(req)

		cands, err := l.fetchCandidates(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(cands) == 0 {
			continue
		}

		baseW := computeBaseWeights(cands, movieLastWatch, actorLastWatch, now)
		alive := make([]int, len(cands))
		for i := range alive {
			alive[i] = i
		}

		groups = append(groups, &pickGroup{
			req:        req,
			weight:     float64(rw.w),
			candidates: cands,
			baseW:      baseW,
			alive:      alive,
		})
		totalWeight += float64(rw.w)
	}

	if len(groups) == 0 {
		return nil, nil
	}

	// 最大余数法分配 target
	targets := allocateTargets(totalWeight, n, groups)
	for i := range groups {
		groups[i].target = targets[i]
	}

	// 全局抽样（共享演员/电影去重）
	selected := make([]*types.MovieType, 0, n)
	selectedActors := make(map[string]struct{})
	selectedMovie := make(map[string]struct{})
	rnd := rand.New(rand.NewSource(effectiveSeed(pickCfg.RandomSeed)))

	for {
		madeProgress := false
		for _, g := range groups {
			if g.target <= 0 || len(g.alive) == 0 {
				continue
			}
			// 该组抽 1 个
			picked, ok := pickOneFromGroup(g, selectedActors, selectedMovie, rnd)
			if !ok {
				continue
			}
			selected = append(selected, picked)
			g.target--
			madeProgress = true

			// 填满就返回
			if len(selected) >= n {
				return selected, nil
			}
		}
		if !madeProgress {
			break
		}
	}
	return selected, nil
}

//
// ======================= 复用的内部逻辑/工具 =======================
//

// 单组内抽 1 个（加权 + 多样性 + 无放回），成功返回 true
func pickOneFromGroup(
	g *pickGroup,
	selectedActors map[string]struct{},
	selectedMovie map[string]struct{},
	rnd *rand.Rand,
) (*types.MovieType, bool) {

	// 计算当前权重
	cur := make([]float64, len(g.alive))
	var sum float64
	for j, idx := range g.alive {
		m := g.candidates[idx]
		// 同电影去重
		if m != nil && m.JavId != "" {
			if _, ok := selectedMovie[m.JavId]; ok {
				cur[j] = 0
				continue
			}
		}
		overlap := overlapActorCount(m, selectedActors)
		diversity := 1.0 / (1.0 + pickCfg.DiversityBeta*float64(overlap))
		w := g.baseW[idx] * diversity
		if w < pickCfg.MinWeight {
			w = pickCfg.MinWeight
		}
		cur[j] = w
		sum += w
	}
	// 全 0 退化为等概率
	if sum <= 0 {
		for j := range cur {
			cur[j] = 1
		}
		sum = float64(len(cur))
	}
	// 加权随机
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
	// 记录 + 无放回
	idx := g.alive[pos]
	picked := g.candidates[idx]
	if picked != nil && picked.JavId != "" {
		selectedMovie[picked.JavId] = struct{}{}
	}
	for _, c := range picked.Cast {
		name := canonicalActorName(c)
		if name != "" {
			selectedActors[name] = struct{}{}
		}
	}
	g.alive = append(g.alive[:pos], g.alive[pos+1:]...)
	return picked, true
}

// 加载历史映射
func (l *ScService) loadWatchMaps(ctx context.Context) (movieLast, actorLast map[string]int64, now time.Time, err error) {
	history, err := l.allScMovies(ctx)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	movieLast, actorLast = buildWatchMaps(history)
	return movieLast, actorLast, time.Now(), nil
}

// 拉取候选集
func (l *ScService) fetchCandidates(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.MovieType, error) {
	resp, err := l.movieSvc.ListMovieFull(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.List, nil
}

// 计算基础权重
func computeBaseWeights(cands []*types.MovieType, movieLast, actorLast map[string]int64, now time.Time) []float64 {
	base := make([]float64, len(cands))
	for i, m := range cands {
		w := scoreMovie(m, movieLast, actorLast, now)
		if w < pickCfg.MinWeight {
			w = pickCfg.MinWeight
		}
		base[i] = w
	}
	return base
}

// 最大余数法分配
func allocateTargets(totalWeight float64, n int, groups []*pickGroup) []int {
	targets := make([]int, len(groups))
	if totalWeight <= 0 {
		base := n / len(groups)
		rem := n % len(groups)
		for i := range groups {
			targets[i] = base
			if i < rem {
				targets[i]++
			}
		}
		return targets
	}
	type frac struct {
		i    int
		part float64
	}
	fracs := make([]frac, 0, len(groups))
	sum := 0
	for i, g := range groups {
		exact := float64(n) * g.weight / totalWeight
		floor := int(math.Floor(exact))
		targets[i] = floor
		sum += floor
		fracs = append(fracs, frac{i: i, part: exact - float64(floor)})
	}
	rem := n - sum
	sort.Slice(fracs, func(a, b int) bool { return fracs[a].part > fracs[b].part })
	for i := 0; i < rem && i < len(fracs); i++ {
		targets[fracs[i].i]++
	}
	return targets
}

func effectiveSeed(seed int64) int64 {
	if seed > 0 {
		return seed
	}
	return time.Now().UnixNano()
}

// 为电影打分（越大越容易被选中）
// - 硬阈值：BlockMovieDays / BlockActorDays
// - 演员近期软惩罚：ActorPenaltyAlpha * exp(-ageDays / HalfLifeActorDays)
func scoreMovie(m *types.MovieType, movieLast map[string]int64, actorLast map[string]int64, now time.Time) float64 {
	if m == nil {
		return 0
	}

	// 硬拦：最近看过电影直接 0
	if pickCfg.BlockMovieDays > 0 {
		if last, ok := movieLast[m.JavId]; ok && last > 0 {
			if ageInDays(last, now) < float64(pickCfg.BlockMovieDays) {
				return 0
			}
		}
	}
	// 硬拦：最近看过演员直接 0
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

	// 软惩：演员最近性（取该片演员中 recentness 最大的那位）
	score := 1.0
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

func buildWatchMaps(history []*scMovies) (map[string]int64, map[string]int64) {
	movieLast := make(map[string]int64, len(history))
	actorLast := make(map[string]int64)
	for _, h := range history {
		if h == nil || h.MovieType == nil {
			continue
		}
		mv := h.MovieType
		wt := h.WatchedTime
		if mv.JavId != "" {
			if old, ok := movieLast[mv.JavId]; !ok || wt > old {
				movieLast[mv.JavId] = wt
			}
		}
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
		return 36500 // 非法时间当作很久以前
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
	if c.Name != "" {
		return c.Name
	}
	return c.NameShow
}
