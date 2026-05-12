package sc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

type SmartPickOptions struct {
	CastLastScBlockDays          int     `json:"castLastScBlockDays"`
	LastScEventBlockDays         int     `json:"lastScEventBlockDays"`
	Rank20Min                    int     `json:"rank20Min"`
	Rank100Min                   int     `json:"rank100Min"`
	Rank500Min                   int     `json:"rank500Min"`
	CastLastScPenaltyAlpha       float64 `json:"castLastScPenaltyAlpha"`
	LastScEventPenaltyAlpha      float64 `json:"lastScEventPenaltyAlpha"`
	CastOwnedScRatioPenaltyAlpha float64 `json:"castOwnedScRatioPenaltyAlpha"`
	MovieHasScPenaltyAlpha       float64 `json:"movieHasScPenaltyAlpha"`
	RandomSeed                   int64   `json:"randomSeed,omitempty"`
}

const (
	SmartPickSourceWMedia = "wmedia"
)

type smartPickGroup struct {
	req        *types.ListMovieFullRequest
	weight     float64
	candidates []*types.MovieType
	baseW      []float64
	alive      []int
}

type smartCastStat struct {
	LastScTime         int64
	LastScEventTime    int64
	OwnedWMediaNumber  int64
	OwnedScMovieNumber int64
}

type SmartPickInfo struct {
	RawCandidateMovieCount         int    `json:"raw_candidate_movie_count"`
	CastCount                      int    `json:"cast_count"`
	AfterBlockCastCount            int    `json:"after_block_cast_count"`
	AfterBlockMovieCount           int    `json:"after_block_movie_count"`
	CastLastScBlockedCastCount     int    `json:"cast_last_sc_blocked_cast_count"`
	LastScEventBlockedCastCount    int    `json:"last_sc_event_blocked_cast_count"`
	CastLastScBlockedMovieCount    int    `json:"cast_last_sc_blocked_movie_count"`
	LastScEventBlockedMovieCount   int    `json:"last_sc_event_blocked_movie_count"`
	SelectedActorSkippedMovieCount int    `json:"selected_actor_skipped_movie_count"`
	TotalSizeGB                    string `json:"total_size_gb,omitempty"`
}

type SmartPickResult struct {
	Movies []*types.MovieType `json:"movies"`
	Info   SmartPickInfo      `json:"info"`
}

type smartRankBucket int

const (
	smartBucket20 smartRankBucket = iota + 1
	smartBucket100
	smartBucket500
	smartBucketFree
)

type smartBucketPlan struct {
	bucket smartRankBucket
	target int
	label  string
}

func NormalizeSmartPickSource(source string) string {
	return SmartPickSourceWMedia
}

func (l *ScService) SmartPickCopyFromRequests(ctx context.Context, reqs []PickRequestWithWeight, n int, opt SmartPickOptions, source string) ([]*types.MovieType, error) {
	return l.SmartPickFromRequests(ctx, reqs, n, opt, source)
}

func (l *ScService) SmartPickFromRequests(ctx context.Context, reqs []PickRequestWithWeight, n int, opt SmartPickOptions, source string) ([]*types.MovieType, error) {
	result, err := l.SmartPickWithInfoFromRequests(ctx, reqs, n, opt, source)
	if err != nil {
		return nil, err
	}
	return result.Movies, nil
}

func (l *ScService) SmartPickWithInfoFromRequests(ctx context.Context, reqs []PickRequestWithWeight, n int, opt SmartPickOptions, source string) (*SmartPickResult, error) {
	if len(reqs) == 0 {
		return nil, errors.New("reqs is empty")
	}
	if n <= 0 {
		return nil, errors.New("pickN must be > 0")
	}

	opt, err := normalizeSmartPickOptions(n, opt)
	if err != nil {
		return nil, err
	}

	source = NormalizeSmartPickSource(source)

	groups, err := l.buildSmartPickGroups(ctx, reqs, source)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, errors.New("no candidates found")
	}

	castStats, err := l.loadSmartCastStats(ctx, groups)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	info := analyzeSmartPickInfo(groups, castStats, opt, now)

	for _, g := range groups {
		g.baseW = computeSmartBaseWeights(g.candidates, castStats, opt)
	}

	plans := buildSmartBucketPlans(n, opt)
	selected := make([]*types.MovieType, 0, n)
	selectedActors := make(map[string]struct{})
	selectedMovies := make(map[string]struct{})
	rnd := rand.New(rand.NewSource(effectiveSeed(opt.RandomSeed)))

	for _, plan := range plans {
		quotas, err := allocateSmartGroupQuotas(groups, plan.bucket, plan.target, selectedActors, selectedMovies, castStats, opt, now)
		if err != nil {
			return nil, fmt.Errorf("%s candidates insufficient: %w", plan.label, err)
		}
		for _, g := range groups {
			quota := quotas[g]
			for i := 0; i < quota; i++ {
				picked, err := pickSmartMovieFromGroup(g, plan.bucket, selectedActors, selectedMovies, castStats, opt, now, rnd)
				if err != nil {
					return nil, fmt.Errorf("%s group quota unmet: %w", plan.label, err)
				}
				selected = append(selected, picked)
				for _, name := range primaryActorKeys(picked) {
					selectedActors[name] = struct{}{}
				}
				if picked != nil && picked.JavId != "" {
					selectedMovies[picked.JavId] = struct{}{}
					pruneMovieFromSmartGroups(groups, picked.JavId)
				}
			}
		}
	}

	l.LogPicksBySource(selected, source)
	info.SelectedActorSkippedMovieCount = countSmartSelectedActorSkippedMovies(groups, selected)
	return &SmartPickResult{
		Movies: selected,
		Info:   info,
	}, nil
}

func normalizeSmartPickOptions(n int, opt SmartPickOptions) (SmartPickOptions, error) {
	if opt.CastLastScBlockDays < 0 {
		return opt, errors.New("castLastScBlockDays must be >= 0")
	}
	if opt.LastScEventBlockDays < 0 {
		return opt, errors.New("lastScEventBlockDays must be >= 0")
	}
	if opt.Rank20Min < 0 || opt.Rank100Min < 0 || opt.Rank500Min < 0 {
		return opt, errors.New("rank min values must be >= 0")
	}
	if opt.Rank20Min > opt.Rank100Min || opt.Rank100Min > opt.Rank500Min || opt.Rank500Min > n {
		return opt, errors.New("require 0 <= rank20Min <= rank100Min <= rank500Min <= pickN")
	}
	if opt.CastLastScPenaltyAlpha < 0 || opt.LastScEventPenaltyAlpha < 0 || opt.CastOwnedScRatioPenaltyAlpha < 0 || opt.MovieHasScPenaltyAlpha < 0 {
		return opt, errors.New("penalty alpha must be >= 0")
	}
	return opt, nil
}

func (l *ScService) buildSmartPickGroups(ctx context.Context, reqs []PickRequestWithWeight, source string) ([]*smartPickGroup, error) {
	groups := make([]*smartPickGroup, 0, len(reqs))
	for i := range reqs {
		req := reqs[i].Req
		normalizeSmartPickReq(&req, source)

		cands, err := l.fetchCandidates(ctx, &req)
		if err != nil {
			return nil, err
		}
		if len(cands) == 0 {
			continue
		}

		alive := make([]int, len(cands))
		for idx := range alive {
			alive[idx] = idx
		}

		weight := float64(reqs[i].Weight)
		if weight <= 0 {
			weight = 1
		}

		groups = append(groups, &smartPickGroup{
			req:        &req,
			weight:     weight,
			candidates: cands,
			alive:      alive,
		})
	}
	return groups, nil
}

func normalizeSmartPickReq(req *types.ListMovieFullRequest, source string) {
	req.MediaOwned = consts.MovieAll
	req.MediaOwned = consts.OwnedAllNotRemoved
	req.Page = 1
	req.PageSize = 100000
	ensureReqDefaults(req)
}

func (l *ScService) loadSmartCastStats(ctx context.Context, groups []*smartPickGroup) (map[string]smartCastStat, error) {
	personIDs := make([]int64, 0, 128)
	names := make([]string, 0, 128)
	for _, g := range groups {
		if g == nil {
			continue
		}
		for _, movie := range g.candidates {
			for _, cast := range primaryActorInfos(movie) {
				if cast == nil {
					continue
				}
				if cast.PersonId > 0 {
					personIDs = append(personIDs, cast.PersonId)
					continue
				}
				if name := canonicalActorName(cast); name != "" {
					names = append(names, name)
				}
			}
		}
	}

	stats := make(map[string]smartCastStat, len(personIDs)+len(names))

	personRows, err := l.personFindByIDs(ctx, personIDs)
	if err != nil {
		return nil, err
	}
	personOwnedScMap, err := l.personCountOwnedScMovieNumbersByIDs(ctx, personIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range personRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		stats["p:"+fmt.Sprintf("%d", row.Id)] = smartCastStat{
			LastScTime:         row.LastScTime,
			LastScEventTime:    row.LastScEventTime,
			OwnedWMediaNumber:  row.OwnedWMediaNumber,
			OwnedScMovieNumber: personOwnedScMap[row.Id],
		}
	}

	rows, err := l.castFindByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	ownedScMap, err := l.castCountOwnedScMovieNumbersByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil || row.Name == "" {
			continue
		}
		stats["n:"+row.Name] = smartCastStat{
			LastScTime:         row.LastScTime,
			LastScEventTime:    row.LastScEventTime,
			OwnedWMediaNumber:  row.OwnedWMediaNumber,
			OwnedScMovieNumber: ownedScMap[row.Name],
		}
	}

	for _, personID := range personIDs {
		key := "p:" + fmt.Sprintf("%d", personID)
		if _, ok := stats[key]; ok {
			continue
		}
		stats[key] = smartCastStat{}
	}
	for _, name := range names {
		key := "n:" + name
		if _, ok := stats[key]; ok {
			continue
		}
		stats[key] = smartCastStat{}
	}
	return stats, nil
}

func computeSmartBaseWeights(cands []*types.MovieType, castStats map[string]smartCastStat, opt SmartPickOptions) []float64 {
	out := make([]float64, len(cands))
	now := time.Now()
	for i, movie := range cands {
		out[i] = smartMovieBaseWeight(movie, castStats, opt, now)
	}
	return out
}

func smartMovieBaseWeight(movie *types.MovieType, castStats map[string]smartCastStat, opt SmartPickOptions, now time.Time) float64 {
	movieHasScFactor := 1.0
	if movie != nil && movie.ScTimes > 0 {
		movieHasScFactor = 1.0 - opt.MovieHasScPenaltyAlpha
		if movieHasScFactor < 0.0001 {
			movieHasScFactor = 0.0001
		}
	}

	actors := primaryActorKeys(movie)
	if len(actors) == 0 {
		return movieHasScFactor
	}

	total := 0.0
	count := 0.0
	for _, name := range actors {
		stat := castStats[name]
		if opt.CastLastScBlockDays > 0 && stat.LastScTime > 0 && ageInDays(stat.LastScTime, now) < float64(opt.CastLastScBlockDays) {
			return 0.0001
		}
		if opt.LastScEventBlockDays > 0 && stat.LastScEventTime > 0 && ageInDays(stat.LastScEventTime, now) < float64(opt.LastScEventBlockDays) {
			return 0.0001
		}

		recentPenalty := 0.0
		if stat.LastScTime > 0 {
			daysSince := ageInDays(stat.LastScTime, now)
			shifted := daysSince - float64(opt.CastLastScBlockDays)
			if shifted < 0 {
				shifted = 0
			}
			recentPenalty = math.Exp(-shifted / 90.0)
		}

		lastScEventPenalty := 0.0
		if stat.LastScEventTime > 0 {
			daysSince := ageInDays(stat.LastScEventTime, now)
			shifted := daysSince - float64(opt.LastScEventBlockDays)
			if shifted < 0 {
				shifted = 0
			}
			lastScEventPenalty = math.Exp(-shifted / 90.0)
		}

		ownedScRatio := 0.0
		if stat.OwnedWMediaNumber > 0 {
			ownedScRatio = float64(stat.OwnedScMovieNumber) / float64(stat.OwnedWMediaNumber)
		}

		factor := (1.0 - opt.CastLastScPenaltyAlpha*recentPenalty) *
			(1.0 - opt.LastScEventPenaltyAlpha*lastScEventPenalty) *
			(1.0 - opt.CastOwnedScRatioPenaltyAlpha*ownedScRatio)
		if factor < 0.0001 {
			factor = 0.0001
		}
		total += factor
		count++
	}

	if count == 0 {
		return movieHasScFactor
	}
	return (total / count) * movieHasScFactor
}

func buildSmartBucketPlans(n int, opt SmartPickOptions) []smartBucketPlan {
	return []smartBucketPlan{
		{bucket: smartBucket20, target: opt.Rank20Min, label: "rank20"},
		{bucket: smartBucket100, target: opt.Rank100Min - opt.Rank20Min, label: "rank100"},
		{bucket: smartBucket500, target: opt.Rank500Min - opt.Rank100Min, label: "rank500"},
		{bucket: smartBucketFree, target: n - opt.Rank500Min, label: "free"},
	}
}

func allocateSmartGroupQuotas(
	groups []*smartPickGroup,
	bucket smartRankBucket,
	target int,
	selectedActors map[string]struct{},
	selectedMovies map[string]struct{},
	castStats map[string]smartCastStat,
	opt SmartPickOptions,
	now time.Time,
) (map[*smartPickGroup]int, error) {
	out := make(map[*smartPickGroup]int, len(groups))
	if target <= 0 {
		return out, nil
	}
	type quotaRow struct {
		group     *smartPickGroup
		quota     int
		remainder float64
	}

	rows := make([]quotaRow, 0, len(groups))
	sum := 0.0
	totalEligible := 0
	for _, g := range groups {
		if g == nil || len(g.alive) == 0 {
			continue
		}
		eligibleCount := countEligibleMovies(g, bucket, selectedActors, selectedMovies, castStats, opt, now)
		if eligibleCount <= 0 {
			continue
		}
		w := g.weight
		if w <= 0 {
			w = 1
		}
		rows = append(rows, quotaRow{group: g})
		sum += w
		totalEligible += eligibleCount
	}
	if len(rows) == 0 {
		return nil, errors.New("no group can satisfy current bucket")
	}
	if totalEligible < target {
		return nil, errors.New("bucket total eligible candidates insufficient")
	}
	for i := range rows {
		w := rows[i].group.weight
		if w <= 0 {
			w = 1
		}
		exact := float64(target) * w / sum
		rows[i].quota = int(math.Floor(exact))
		rows[i].remainder = exact - float64(rows[i].quota)
		out[rows[i].group] = rows[i].quota
	}
	assigned := 0
	for _, row := range rows {
		assigned += row.quota
	}
	left := target - assigned
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].remainder == rows[j].remainder {
			return rows[i].group.weight > rows[j].group.weight
		}
		return rows[i].remainder > rows[j].remainder
	})
	for i := 0; i < left; i++ {
		row := rows[i%len(rows)]
		out[row.group]++
	}
	for _, row := range rows {
		eligibleCount := countEligibleMovies(row.group, bucket, selectedActors, selectedMovies, castStats, opt, now)
		if out[row.group] > eligibleCount {
			return nil, fmt.Errorf("group weight quota %d exceeds eligible %d", out[row.group], eligibleCount)
		}
	}
	return out, nil
}

func countEligibleMovies(
	group *smartPickGroup,
	bucket smartRankBucket,
	selectedActors map[string]struct{},
	selectedMovies map[string]struct{},
	castStats map[string]smartCastStat,
	opt SmartPickOptions,
	now time.Time,
) int {
	count := 0
	for _, idx := range group.alive {
		if idx < 0 || idx >= len(group.candidates) {
			continue
		}
		if smartMovieEligible(group.candidates[idx], bucket, selectedActors, selectedMovies, castStats, opt, now) {
			count++
		}
	}
	return count
}

func pickSmartMovieFromGroup(
	group *smartPickGroup,
	bucket smartRankBucket,
	selectedActors map[string]struct{},
	selectedMovies map[string]struct{},
	castStats map[string]smartCastStat,
	opt SmartPickOptions,
	now time.Time,
	rnd *rand.Rand,
) (*types.MovieType, error) {
	eligiblePos := make([]int, 0, len(group.alive))
	weights := make([]float64, 0, len(group.alive))
	sum := 0.0

	for pos, idx := range group.alive {
		if idx < 0 || idx >= len(group.candidates) {
			continue
		}
		movie := group.candidates[idx]
		if !smartMovieEligible(movie, bucket, selectedActors, selectedMovies, castStats, opt, now) {
			continue
		}
		w := group.baseW[idx]
		if w < 0.0001 {
			w = 0.0001
		}
		eligiblePos = append(eligiblePos, pos)
		weights = append(weights, w)
		sum += w
	}

	if len(eligiblePos) == 0 {
		return nil, errors.New("group has no eligible movie")
	}

	r := rnd.Float64() * sum
	acc := 0.0
	chosenPos := eligiblePos[len(eligiblePos)-1]
	for i, pos := range eligiblePos {
		acc += weights[i]
		if r <= acc {
			chosenPos = pos
			break
		}
	}

	idx := group.alive[chosenPos]
	movie := group.candidates[idx]
	group.alive = append(group.alive[:chosenPos], group.alive[chosenPos+1:]...)
	return movie, nil
}

func smartMovieEligible(
	movie *types.MovieType,
	bucket smartRankBucket,
	selectedActors map[string]struct{},
	selectedMovies map[string]struct{},
	castStats map[string]smartCastStat,
	opt SmartPickOptions,
	now time.Time,
) bool {
	if movie == nil || movie.JavId == "" {
		return false
	}
	if _, ok := selectedMovies[movie.JavId]; ok {
		return false
	}
	if !smartBucketAllowsMovie(bucket, movie) {
		return false
	}

	for _, name := range primaryActorKeys(movie) {
		if _, ok := selectedActors[name]; ok {
			return false
		}
		stat := castStats[name]
		if opt.CastLastScBlockDays > 0 && stat.LastScTime > 0 && ageInDays(stat.LastScTime, now) < float64(opt.CastLastScBlockDays) {
			return false
		}
		if opt.LastScEventBlockDays > 0 && stat.LastScEventTime > 0 && ageInDays(stat.LastScEventTime, now) < float64(opt.LastScEventBlockDays) {
			return false
		}
	}
	return true
}

func smartBucketAllowsMovie(bucket smartRankBucket, movie *types.MovieType) bool {
	if bucket == smartBucketFree {
		return true
	}
	if movie == nil {
		return false
	}
	rank := movie.HighestRank
	switch bucket {
	case smartBucket20:
		return rank > 0 && rank <= 20
	case smartBucket100:
		return rank >= 21 && rank <= 100
	case smartBucket500:
		return rank >= 101 && rank <= 500
	default:
		return false
	}
}

func primaryActorKeys(movie *types.MovieType) []string {
	infos := primaryActorInfos(movie)
	out := make([]string, 0, len(infos))
	for _, cast := range infos {
		key := canonicalActorKey(cast)
		if key != "" {
			out = append(out, key)
		}
	}
	return out
}

func pruneMovieFromSmartGroups(groups []*smartPickGroup, javID string) {
	if javID == "" {
		return
	}
	for _, g := range groups {
		if g == nil || len(g.alive) == 0 {
			continue
		}
		dst := g.alive[:0]
		for _, idx := range g.alive {
			if idx < 0 || idx >= len(g.candidates) {
				continue
			}
			movie := g.candidates[idx]
			if movie != nil && movie.JavId == javID {
				continue
			}
			dst = append(dst, idx)
		}
		g.alive = dst
	}
}

func analyzeSmartPickInfo(groups []*smartPickGroup, castStats map[string]smartCastStat, opt SmartPickOptions, now time.Time) SmartPickInfo {
	info := SmartPickInfo{}
	movies := collectSmartUniqueMovies(groups)
	actorSeen := make(map[string]struct{}, 128)
	afterBlockActorSeen := make(map[string]struct{}, 128)
	blockedByLastScActors := make(map[string]struct{}, 64)
	blockedByLastScEventActors := make(map[string]struct{}, 64)
	blockedByLastScMovies := make(map[string]struct{}, 64)
	blockedByLastScEventMovies := make(map[string]struct{}, 64)

	for _, movie := range movies {
		actorKeys := primaryActorKeys(movie)
		for _, key := range actorKeys {
			actorSeen[key] = struct{}{}
		}

		blockedByLastSc := false
		blockedByLastScEvent := false
		for _, key := range actorKeys {
			stat := castStats[key]
			if opt.CastLastScBlockDays > 0 && stat.LastScTime > 0 && ageInDays(stat.LastScTime, now) < float64(opt.CastLastScBlockDays) {
				blockedByLastSc = true
				blockedByLastScActors[key] = struct{}{}
			}
			if opt.LastScEventBlockDays > 0 && stat.LastScEventTime > 0 && ageInDays(stat.LastScEventTime, now) < float64(opt.LastScEventBlockDays) {
				blockedByLastScEvent = true
				blockedByLastScEventActors[key] = struct{}{}
			}
		}

		if blockedByLastSc {
			blockedByLastScMovies[movie.JavId] = struct{}{}
		}
		if blockedByLastScEvent {
			blockedByLastScEventMovies[movie.JavId] = struct{}{}
		}
		if blockedByLastSc || blockedByLastScEvent {
			continue
		}

		info.AfterBlockMovieCount++
		for _, key := range actorKeys {
			afterBlockActorSeen[key] = struct{}{}
		}
	}

	info.RawCandidateMovieCount = len(movies)
	info.CastCount = len(actorSeen)
	info.AfterBlockCastCount = len(afterBlockActorSeen)
	info.CastLastScBlockedCastCount = len(blockedByLastScActors)
	info.LastScEventBlockedCastCount = len(blockedByLastScEventActors)
	info.CastLastScBlockedMovieCount = len(blockedByLastScMovies)
	info.LastScEventBlockedMovieCount = len(blockedByLastScEventMovies)
	return info
}

func collectSmartUniqueMovies(groups []*smartPickGroup) []*types.MovieType {
	seen := make(map[string]*types.MovieType, 256)
	for _, g := range groups {
		if g == nil {
			continue
		}
		for _, movie := range g.candidates {
			if movie == nil || movie.JavId == "" {
				continue
			}
			if _, ok := seen[movie.JavId]; ok {
				continue
			}
			seen[movie.JavId] = movie
		}
	}
	out := make([]*types.MovieType, 0, len(seen))
	for _, movie := range seen {
		out = append(out, movie)
	}
	return out
}

func countSmartSelectedActorSkippedMovies(groups []*smartPickGroup, selected []*types.MovieType) int {
	if len(selected) == 0 {
		return 0
	}
	selectedActors := make(map[string]struct{}, 64)
	selectedMovies := make(map[string]struct{}, len(selected))
	for _, movie := range selected {
		if movie == nil || movie.JavId == "" {
			continue
		}
		selectedMovies[movie.JavId] = struct{}{}
		for _, key := range primaryActorKeys(movie) {
			selectedActors[key] = struct{}{}
		}
	}

	count := 0
	seen := make(map[string]struct{}, 128)
	for _, g := range groups {
		if g == nil {
			continue
		}
		for _, movie := range g.candidates {
			if movie == nil || movie.JavId == "" {
				continue
			}
			if _, ok := selectedMovies[movie.JavId]; ok {
				continue
			}
			if _, ok := seen[movie.JavId]; ok {
				continue
			}
			for _, key := range primaryActorKeys(movie) {
				if _, ok := selectedActors[key]; ok {
					seen[movie.JavId] = struct{}{}
					count++
					break
				}
			}
		}
	}
	return count
}

func sortSmartPickedMoviesByBirth(movies []*types.MovieType) {
	sort.Slice(movies, func(i, j int) bool {
		ti := parseDate(movies[i].FilmBirthDateWMedia)
		tj := parseDate(movies[j].FilmBirthDateWMedia)
		return ti.Before(tj)
	})
}
