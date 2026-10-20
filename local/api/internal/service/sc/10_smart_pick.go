package sc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/types"
)

type SmartPickOptions struct {
	CastSelectedScBlockDays        int     `json:"castSelectedScBlockDays"`
	CastParticipatedScBlockDays    int     `json:"castParticipatedScBlockDays"`
	Rank20Min                      int     `json:"rank20Min"`
	Rank100Min                     int     `json:"rank100Min"`
	Rank500Min                     int     `json:"rank500Min"`
	CastSelectedScPenaltyAlpha     float64 `json:"castSelectedScPenaltyAlpha"`
	CastParticipatedScPenaltyAlpha float64 `json:"castParticipatedScPenaltyAlpha"`
	CastOwnedScRatioPenaltyAlpha   float64 `json:"castOwnedScRatioPenaltyAlpha"`
	MovieHasSelectedScPenaltyAlpha float64 `json:"movieHasSelectedScPenaltyAlpha"`
	RandomSeed                     int64   `json:"randomSeed,omitempty"`
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
	LastSelectedScTime     int64
	LastParticipatedScTime int64
	OwnedWMediaNumber      int64
	OwnedScMovieNumber     int64
}

type SmartPickInfo struct {
	RawCandidateMovieCount              int    `json:"raw_candidate_movie_count"`
	CastCount                           int    `json:"cast_count"`
	AfterBlockCastCount                 int    `json:"after_block_cast_count"`
	AfterBlockMovieCount                int    `json:"after_block_movie_count"`
	CastSelectedScBlockedCastCount      int    `json:"cast_selected_sc_blocked_cast_count"`
	CastParticipatedScBlockedCastCount  int    `json:"cast_participated_sc_blocked_cast_count"`
	CastSelectedScBlockedMovieCount     int    `json:"cast_selected_sc_blocked_movie_count"`
	CastParticipatedScBlockedMovieCount int    `json:"cast_participated_sc_blocked_movie_count"`
	SelectedActorSkippedMovieCount      int    `json:"selected_actor_skipped_movie_count"`
	TotalSizeGB                         string `json:"total_size_gb,omitempty"`
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

func (l *Service) SmartPickCopyFromRequests(ctx context.Context, reqs []PickRequestWithWeight, n int, opt SmartPickOptions, source string) ([]*types.MovieType, error) {
	return l.SmartPickFromRequests(ctx, reqs, n, opt, source)
}

func (l *Service) SmartPickFromRequests(ctx context.Context, reqs []PickRequestWithWeight, n int, opt SmartPickOptions, source string) ([]*types.MovieType, error) {
	result, err := l.SmartPickWithInfoFromRequests(ctx, reqs, n, opt, source)
	if err != nil {
		return nil, err
	}
	return result.Movies, nil
}

func (l *Service) SmartPickWithInfoFromRequests(ctx context.Context, reqs []PickRequestWithWeight, n int, opt SmartPickOptions, source string) (*SmartPickResult, error) {
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
	if opt.CastSelectedScBlockDays < 0 {
		return opt, errors.New("castSelectedScBlockDays must be >= 0")
	}
	if opt.CastParticipatedScBlockDays < 0 {
		return opt, errors.New("castParticipatedScBlockDays must be >= 0")
	}
	if opt.Rank20Min < 0 || opt.Rank100Min < 0 || opt.Rank500Min < 0 {
		return opt, errors.New("rank min values must be >= 0")
	}
	if opt.Rank20Min > opt.Rank100Min || opt.Rank100Min > opt.Rank500Min || opt.Rank500Min > n {
		return opt, errors.New("require 0 <= rank20Min <= rank100Min <= rank500Min <= pickN")
	}
	if opt.CastSelectedScPenaltyAlpha < 0 || opt.CastParticipatedScPenaltyAlpha < 0 || opt.CastOwnedScRatioPenaltyAlpha < 0 || opt.MovieHasSelectedScPenaltyAlpha < 0 {
		return opt, errors.New("penalty alpha must be >= 0")
	}
	return opt, nil
}

func (l *Service) buildSmartPickGroups(ctx context.Context, reqs []PickRequestWithWeight, source string) ([]*smartPickGroup, error) {
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
		sort.SliceStable(cands, func(left, right int) bool {
			if cands[left] == nil {
				return false
			}
			if cands[right] == nil {
				return true
			}
			return cands[left].JavId < cands[right].JavId
		})

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

func normalizeSmartPickReq(req *types.ListMovieFullRequest, _ string) {
	req.MediaOwned = consts.OwnedAllNotRemoved
	req.Page = 1
	req.PageSize = 0
}

func (l *Service) loadSmartCastStats(ctx context.Context, groups []*smartPickGroup) (map[string]smartCastStat, error) {
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
			LastSelectedScTime:     row.LastSelectedScTime,
			LastParticipatedScTime: row.LastParticipatedScTime,
			OwnedWMediaNumber:      row.OwnedWMediaNumber,
			OwnedScMovieNumber:     personOwnedScMap[row.Id],
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
			LastSelectedScTime:     row.LastSelectedScTime,
			LastParticipatedScTime: row.LastParticipatedScTime,
			OwnedWMediaNumber:      row.OwnedWMediaNumber,
			OwnedScMovieNumber:     ownedScMap[row.Name],
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
		movieHasScFactor = 1.0 - opt.MovieHasSelectedScPenaltyAlpha
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
		if opt.CastSelectedScBlockDays > 0 && stat.LastSelectedScTime > 0 && ageInDays(stat.LastSelectedScTime, now) < float64(opt.CastSelectedScBlockDays) {
			return 0.0001
		}
		if opt.CastParticipatedScBlockDays > 0 && stat.LastParticipatedScTime > 0 && ageInDays(stat.LastParticipatedScTime, now) < float64(opt.CastParticipatedScBlockDays) {
			return 0.0001
		}

		recentPenalty := 0.0
		if stat.LastSelectedScTime > 0 {
			daysSince := ageInDays(stat.LastSelectedScTime, now)
			shifted := daysSince - float64(opt.CastSelectedScBlockDays)
			if shifted < 0 {
				shifted = 0
			}
			recentPenalty = math.Exp(-shifted / 90.0)
		}

		participatedScPenalty := 0.0
		if stat.LastParticipatedScTime > 0 {
			daysSince := ageInDays(stat.LastParticipatedScTime, now)
			shifted := daysSince - float64(opt.CastParticipatedScBlockDays)
			if shifted < 0 {
				shifted = 0
			}
			participatedScPenalty = math.Exp(-shifted / 90.0)
		}

		ownedScRatio := 0.0
		if stat.OwnedWMediaNumber > 0 {
			ownedScRatio = float64(stat.OwnedScMovieNumber) / float64(stat.OwnedWMediaNumber)
		}

		factor := (1.0 - opt.CastSelectedScPenaltyAlpha*recentPenalty) *
			(1.0 - opt.CastParticipatedScPenaltyAlpha*participatedScPenalty) *
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
		group    *smartPickGroup
		quota    int
		capacity int
		order    int
	}

	rows := make([]quotaRow, 0, len(groups))
	sum := 0.0
	totalEligible := 0
	for groupIndex, g := range groups {
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
		rows = append(rows, quotaRow{group: g, capacity: eligibleCount, order: groupIndex})
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
		if rows[i].quota > rows[i].capacity {
			rows[i].quota = rows[i].capacity
		}
		out[rows[i].group] = rows[i].quota
	}
	assigned := 0
	for _, row := range rows {
		assigned += row.quota
	}
	for assigned < target {
		candidate := -1
		for i := range rows {
			if rows[i].quota >= rows[i].capacity {
				continue
			}
			if candidate < 0 {
				candidate = i
				continue
			}
			currentWeight := rows[i].group.weight
			if currentWeight <= 0 {
				currentWeight = 1
			}
			candidateWeight := rows[candidate].group.weight
			if candidateWeight <= 0 {
				candidateWeight = 1
			}
			currentShare := float64(rows[i].quota+1) / currentWeight
			candidateShare := float64(rows[candidate].quota+1) / candidateWeight
			if currentShare < candidateShare || (currentShare == candidateShare && rows[i].order < rows[candidate].order) {
				candidate = i
			}
		}
		if candidate < 0 {
			return nil, errors.New("bucket total eligible candidates insufficient after allocation")
		}
		rows[candidate].quota++
		out[rows[candidate].group]++
		assigned++
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
	type actorMovieChoice struct {
		pos    int
		weight float64
	}
	type actorBucket struct {
		actorWeight    float64
		movieChoices   []actorMovieChoice
		movieWeightSum float64
	}

	actorBuckets := make([]actorBucket, 0, len(group.alive))
	actorBucketIndex := make(map[string]int, len(group.alive))
	actorWeightSum := 0.0

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

		actorKeys := primaryActorKeys(movie)
		if len(actorKeys) == 0 {
			actorKeys = []string{"m:" + movie.JavId}
		}
		actorContribution := w / float64(len(actorKeys))
		for _, actorKey := range actorKeys {
			bucketIdx, ok := actorBucketIndex[actorKey]
			if !ok {
				bucketIdx = len(actorBuckets)
				actorBucketIndex[actorKey] = bucketIdx
				actorBuckets = append(actorBuckets, actorBucket{})
			}
			actorBuckets[bucketIdx].actorWeight += actorContribution
			actorBuckets[bucketIdx].movieChoices = append(actorBuckets[bucketIdx].movieChoices, actorMovieChoice{
				pos:    pos,
				weight: w,
			})
			actorBuckets[bucketIdx].movieWeightSum += w
			actorWeightSum += actorContribution
		}
	}

	if len(actorBuckets) == 0 {
		return nil, errors.New("group has no eligible movie")
	}

	r := rnd.Float64() * actorWeightSum
	acc := 0.0
	chosenActorIdx := len(actorBuckets) - 1
	for i := range actorBuckets {
		acc += actorBuckets[i].actorWeight
		if r <= acc {
			chosenActorIdx = i
			break
		}
	}

	chosenActor := actorBuckets[chosenActorIdx]
	r = rnd.Float64() * chosenActor.movieWeightSum
	acc = 0.0
	chosenPos := chosenActor.movieChoices[len(chosenActor.movieChoices)-1].pos
	for _, choice := range chosenActor.movieChoices {
		acc += choice.weight
		if r <= acc {
			chosenPos = choice.pos
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
		if opt.CastSelectedScBlockDays > 0 && stat.LastSelectedScTime > 0 && ageInDays(stat.LastSelectedScTime, now) < float64(opt.CastSelectedScBlockDays) {
			return false
		}
		if opt.CastParticipatedScBlockDays > 0 && stat.LastParticipatedScTime > 0 && ageInDays(stat.LastParticipatedScTime, now) < float64(opt.CastParticipatedScBlockDays) {
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
	blockedBySelectedScActors := make(map[string]struct{}, 64)
	blockedByParticipatedScActors := make(map[string]struct{}, 64)
	blockedBySelectedScMovies := make(map[string]struct{}, 64)
	blockedByParticipatedScMovies := make(map[string]struct{}, 64)

	for _, movie := range movies {
		actorKeys := primaryActorKeys(movie)
		for _, key := range actorKeys {
			actorSeen[key] = struct{}{}
		}

		blockedBySelectedSc := false
		blockedByParticipatedSc := false
		for _, key := range actorKeys {
			stat := castStats[key]
			if opt.CastSelectedScBlockDays > 0 && stat.LastSelectedScTime > 0 && ageInDays(stat.LastSelectedScTime, now) < float64(opt.CastSelectedScBlockDays) {
				blockedBySelectedSc = true
				blockedBySelectedScActors[key] = struct{}{}
			}
			if opt.CastParticipatedScBlockDays > 0 && stat.LastParticipatedScTime > 0 && ageInDays(stat.LastParticipatedScTime, now) < float64(opt.CastParticipatedScBlockDays) {
				blockedByParticipatedSc = true
				blockedByParticipatedScActors[key] = struct{}{}
			}
		}

		if blockedBySelectedSc {
			blockedBySelectedScMovies[movie.JavId] = struct{}{}
		}
		if blockedByParticipatedSc {
			blockedByParticipatedScMovies[movie.JavId] = struct{}{}
		}
		if blockedBySelectedSc || blockedByParticipatedSc {
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
	info.CastSelectedScBlockedCastCount = len(blockedBySelectedScActors)
	info.CastParticipatedScBlockedCastCount = len(blockedByParticipatedScActors)
	info.CastSelectedScBlockedMovieCount = len(blockedBySelectedScMovies)
	info.CastParticipatedScBlockedMovieCount = len(blockedByParticipatedScMovies)
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
