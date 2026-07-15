package sc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/repository/moviex"
)

type FC2SmartPickRequest struct {
	PostTimeStart         string `json:"PostTimeStart"`
	PostTimeEnd           string `json:"PostTimeEnd"`
	MediaBirthTimeStart   string `json:"MediaBirthTimeStart"`
	MediaBirthTimeEnd     string `json:"MediaBirthTimeEnd"`
	LastSelectedScTimeMin string `json:"LastSelectedScTimeMin"`
	LastSelectedScTimeMax string `json:"LastSelectedScTimeMax"`

	ScTimesMin   *int64 `json:"ScTimesMin"`
	ScTimesMax   *int64 `json:"ScTimesMax"`
	ComeTimesMin *int64 `json:"ComeTimesMin"`
	ComeTimesMax *int64 `json:"ComeTimesMax"`

	FC2Key        string `json:"FC2Key"`
	Code          string `json:"Code"`
	Title         string `json:"Title"`
	PerformerText string `json:"PerformerText"`
	MaskType      string `json:"MaskType"`
	Word          string `json:"Word"`
	MediaDir1     string `json:"MediaDir1"`
	MediaDir2     string `json:"MediaDir2"`
	MediaDir3     string `json:"MediaDir3"`
	MediaDir4     string `json:"MediaDir4"`
	HasSub        *int64 `json:"HasSub"`
	SelfMake      *int64 `json:"SelfMake"`
	HasMask       *int64 `json:"HasMask"`
}

type FC2PickRequestWithWeight struct {
	Req    FC2SmartPickRequest `json:"req"`
	Weight int64               `json:"weight"`
}

type FC2SmartPickOptions struct {
	MovieSelectedScBlockDays       int     `json:"movieSelectedScBlockDays"`
	MovieParticipatedScBlockDays   int     `json:"movieParticipatedScBlockDays"`
	MovieSelectedScPenaltyAlpha    float64 `json:"movieSelectedScPenaltyAlpha"`
	MovieHasSelectedScPenaltyAlpha float64 `json:"movieHasSelectedScPenaltyAlpha"`
	RandomSeed                     int64   `json:"randomSeed,omitempty"`
}

type FC2SmartPickInfo struct {
	RawCandidateMovieCount          int    `json:"raw_candidate_movie_count"`
	AfterBlockMovieCount            int    `json:"after_block_movie_count"`
	SelectedScBlockedMovieCount     int    `json:"selected_sc_blocked_movie_count"`
	ParticipatedScBlockedMovieCount int    `json:"participated_sc_blocked_movie_count"`
	SelectedMovieCount              int    `json:"selected_movie_count"`
	TotalSizeGB                     string `json:"total_size_gb,omitempty"`
}

type FC2SmartPickMovie struct {
	*moviex.FC2Movie
	VideoURL               string
	MediaSize              int64
	ScTimes                int64
	ComeTimes              int64
	LastSelectedScTime     int64
	LastParticipatedScTime int64
}

type FC2SmartPickResult struct {
	Movies []*FC2SmartPickMovie `json:"movies"`
	Info   FC2SmartPickInfo     `json:"info"`
}

type fc2SmartPickNormalizedRequest struct {
	req                     FC2SmartPickRequest
	postTimeStart           int64
	postTimeEnd             int64
	lastSelectedScTimeStart int64
	lastSelectedScTimeEnd   int64
	mediaFilter             moviex.FC2SmartPickMediaFilter
}

type fc2SmartPickPreparedGroup struct {
	req       fc2SmartPickNormalizedRequest
	weight    float64
	mediaRows []*moviex.FC2Media
}

type fc2SmartPickGroup struct {
	weight     float64
	candidates []*FC2SmartPickMovie
	alive      []int
}

func (l *ScService) SmartPickFC2WithInfoFromRequests(ctx context.Context, reqs []FC2PickRequestWithWeight, n int, opt FC2SmartPickOptions) (*FC2SmartPickResult, error) {
	if len(reqs) == 0 {
		return nil, errors.New("reqs is empty")
	}
	if n <= 0 {
		return nil, errors.New("pickN must be > 0")
	}
	if err := validateFC2SmartPickOptions(opt); err != nil {
		return nil, err
	}

	groups, err := l.buildFC2SmartPickGroups(ctx, reqs)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, errors.New("no candidates found")
	}

	now := time.Now()
	info := analyzeFC2SmartPickInfo(groups, opt, now)
	selected, err := pickFC2SmartMovies(groups, n, opt, now)
	if err != nil {
		return nil, err
	}
	info.SelectedMovieCount = len(selected)
	return &FC2SmartPickResult{Movies: selected, Info: info}, nil
}

func validateFC2SmartPickOptions(opt FC2SmartPickOptions) error {
	if opt.MovieSelectedScBlockDays < 0 {
		return errors.New("movieSelectedScBlockDays must be >= 0")
	}
	if opt.MovieParticipatedScBlockDays < 0 {
		return errors.New("movieParticipatedScBlockDays must be >= 0")
	}
	if opt.MovieSelectedScPenaltyAlpha < 0 || opt.MovieHasSelectedScPenaltyAlpha < 0 {
		return errors.New("penalty alpha must be >= 0")
	}
	return nil
}

func (l *ScService) buildFC2SmartPickGroups(ctx context.Context, reqs []FC2PickRequestWithWeight) ([]*fc2SmartPickGroup, error) {
	prepared := make([]fc2SmartPickPreparedGroup, 0, len(reqs))
	allMovieIDs := make([]int64, 0)
	for index, input := range reqs {
		normalized, err := normalizeFC2SmartPickRequest(input.Req)
		if err != nil {
			return nil, fmt.Errorf("group %d: %w", index+1, err)
		}
		mediaRows, err := l.deps.FC2MediaModel.ListSmartPickCandidates(ctx, normalized.mediaFilter)
		if err != nil {
			return nil, fmt.Errorf("FC2MediaModel.ListSmartPickCandidates: %w", err)
		}
		weight := float64(input.Weight)
		if weight <= 0 {
			weight = 1
		}
		prepared = append(prepared, fc2SmartPickPreparedGroup{req: normalized, weight: weight, mediaRows: mediaRows})
		for _, mediaRow := range mediaRows {
			if mediaRow != nil && mediaRow.FC2MovieID > 0 {
				allMovieIDs = append(allMovieIDs, mediaRow.FC2MovieID)
			}
		}
	}

	movieRows, err := l.deps.FC2MovieModel.ListByIDs(ctx, allMovieIDs)
	if err != nil {
		return nil, fmt.Errorf("FC2MovieModel.ListByIDs: %w", err)
	}
	statRows, err := l.deps.FC2ScStatModel.ListByMovieIDs(ctx, allMovieIDs)
	if err != nil {
		return nil, fmt.Errorf("FC2ScStatModel.ListByMovieIDs: %w", err)
	}
	movieByID := make(map[int64]*moviex.FC2Movie, len(movieRows))
	for _, row := range movieRows {
		if row != nil && row.ID > 0 {
			movieByID[row.ID] = row
		}
	}
	statByMovieID := make(map[int64]*moviex.FC2ScStat, len(statRows))
	for _, row := range statRows {
		if row != nil && row.FC2MovieID > 0 {
			statByMovieID[row.FC2MovieID] = row
		}
	}

	groups := make([]*fc2SmartPickGroup, 0, len(prepared))
	for _, item := range prepared {
		candidates := make([]*FC2SmartPickMovie, 0, len(item.mediaRows))
		seen := make(map[int64]struct{}, len(item.mediaRows))
		for _, mediaRow := range item.mediaRows {
			if mediaRow == nil || mediaRow.FC2MovieID <= 0 {
				continue
			}
			if _, ok := seen[mediaRow.FC2MovieID]; ok {
				continue
			}
			movieRow := movieByID[mediaRow.FC2MovieID]
			if movieRow == nil {
				continue
			}
			statRow := statByMovieID[mediaRow.FC2MovieID]
			if !matchesFC2SmartPickRequest(item.req, movieRow, statRow) {
				continue
			}
			seen[mediaRow.FC2MovieID] = struct{}{}
			candidates = append(candidates, newFC2SmartPickMovie(movieRow, mediaRow, statRow))
		}
		if len(candidates) == 0 {
			continue
		}
		alive := make([]int, len(candidates))
		for index := range alive {
			alive[index] = index
		}
		groups = append(groups, &fc2SmartPickGroup{weight: item.weight, candidates: candidates, alive: alive})
	}
	return groups, nil
}

func normalizeFC2SmartPickRequest(req FC2SmartPickRequest) (fc2SmartPickNormalizedRequest, error) {
	var out fc2SmartPickNormalizedRequest
	out.req = req
	var err error
	if out.postTimeStart, err = parseFC2SmartPickDate(req.PostTimeStart, false); err != nil {
		return out, fmt.Errorf("PostTimeStart: %w", err)
	}
	if out.postTimeEnd, err = parseFC2SmartPickDate(req.PostTimeEnd, true); err != nil {
		return out, fmt.Errorf("PostTimeEnd: %w", err)
	}
	if out.lastSelectedScTimeStart, err = parseFC2SmartPickDate(req.LastSelectedScTimeMin, false); err != nil {
		return out, fmt.Errorf("LastSelectedScTimeMin: %w", err)
	}
	if out.lastSelectedScTimeEnd, err = parseFC2SmartPickDate(req.LastSelectedScTimeMax, true); err != nil {
		return out, fmt.Errorf("LastSelectedScTimeMax: %w", err)
	}
	if out.mediaFilter.BirthTimeStart, err = parseFC2SmartPickDate(req.MediaBirthTimeStart, false); err != nil {
		return out, fmt.Errorf("MediaBirthTimeStart: %w", err)
	}
	if out.mediaFilter.BirthTimeEnd, err = parseFC2SmartPickDate(req.MediaBirthTimeEnd, true); err != nil {
		return out, fmt.Errorf("MediaBirthTimeEnd: %w", err)
	}
	out.mediaFilter.MediaDir1 = req.MediaDir1
	out.mediaFilter.MediaDir2 = req.MediaDir2
	out.mediaFilter.MediaDir3 = req.MediaDir3
	out.mediaFilter.MediaDir4 = req.MediaDir4
	out.mediaFilter.HasSub = req.HasSub
	out.mediaFilter.SelfMake = req.SelfMake
	out.mediaFilter.HasMask = req.HasMask
	if err := validateFC2SmartPickRange("ScTimes", req.ScTimesMin, req.ScTimesMax); err != nil {
		return out, err
	}
	if err := validateFC2SmartPickRange("ComeTimes", req.ComeTimesMin, req.ComeTimesMax); err != nil {
		return out, err
	}
	return out, nil
}

func parseFC2SmartPickDate(raw string, endOfDay bool) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return 0, errors.New("日期格式必须为 YYYY-MM-DD")
	}
	if endOfDay {
		parsed = parsed.AddDate(0, 0, 1).Add(-time.Second)
	}
	return parsed.Unix(), nil
}

func validateFC2SmartPickRange(name string, minValue, maxValue *int64) error {
	if minValue != nil && *minValue < 0 || maxValue != nil && *maxValue < 0 {
		return fmt.Errorf("%s must be >= 0", name)
	}
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		return fmt.Errorf("%s min must be <= max", name)
	}
	return nil
}

func matchesFC2SmartPickRequest(req fc2SmartPickNormalizedRequest, movie *moviex.FC2Movie, stat *moviex.FC2ScStat) bool {
	if movie == nil {
		return false
	}
	if req.postTimeStart > 0 && movie.PostTime < req.postTimeStart || req.postTimeEnd > 0 && movie.PostTime > req.postTimeEnd {
		return false
	}
	if !containsFold(movie.FC2Key, req.req.FC2Key) || !containsFold(movie.Code, req.req.Code) || !containsFold(movie.Title, req.req.Title) || !containsFold(movie.PerformerText, req.req.PerformerText) {
		return false
	}
	if maskType := strings.TrimSpace(req.req.MaskType); maskType != "" && !strings.EqualFold(strings.TrimSpace(movie.MaskType), maskType) {
		return false
	}
	if word := strings.TrimSpace(req.req.Word); word != "" && !containsFold(movie.FC2Key, word) && !containsFold(movie.Code, word) && !containsFold(movie.Title, word) && !containsFold(movie.PerformerText, word) {
		return false
	}
	scTimes, comeTimes, lastSelectedScTime := int64(0), int64(0), int64(0)
	if stat != nil {
		scTimes, comeTimes, lastSelectedScTime = stat.ScTimes, stat.ComeTimes, stat.LastSelectedScTime
	}
	if !int64InOptionalRange(scTimes, req.req.ScTimesMin, req.req.ScTimesMax) || !int64InOptionalRange(comeTimes, req.req.ComeTimesMin, req.req.ComeTimesMax) {
		return false
	}
	if req.lastSelectedScTimeStart > 0 && lastSelectedScTime < req.lastSelectedScTimeStart || req.lastSelectedScTimeEnd > 0 && lastSelectedScTime > req.lastSelectedScTimeEnd {
		return false
	}
	return true
}

func containsFold(value, query string) bool {
	query = strings.TrimSpace(query)
	return query == "" || strings.Contains(strings.ToLower(value), strings.ToLower(query))
}

func int64InOptionalRange(value int64, minValue, maxValue *int64) bool {
	return (minValue == nil || value >= *minValue) && (maxValue == nil || value <= *maxValue)
}

func newFC2SmartPickMovie(movie *moviex.FC2Movie, media *moviex.FC2Media, stat *moviex.FC2ScStat) *FC2SmartPickMovie {
	item := &FC2SmartPickMovie{
		FC2Movie:  movie,
		VideoURL:  filepath.Join(media.FullDir, media.FileName),
		MediaSize: media.Size,
	}
	if stat != nil {
		item.ScTimes = stat.ScTimes
		item.ComeTimes = stat.ComeTimes
		item.LastSelectedScTime = stat.LastSelectedScTime
		item.LastParticipatedScTime = stat.LastParticipatedScTime
	}
	return item
}

func analyzeFC2SmartPickInfo(groups []*fc2SmartPickGroup, opt FC2SmartPickOptions, now time.Time) FC2SmartPickInfo {
	unique := collectUniqueFC2SmartPickMovies(groups)
	blocked := 0
	selectedScBlocked := 0
	participatedScBlocked := 0
	for _, movie := range unique {
		blockedBySelectedSc := fc2SmartMovieSelectedScBlocked(movie, opt, now)
		blockedByParticipatedSc := fc2SmartMovieParticipatedScBlocked(movie, opt, now)
		if blockedBySelectedSc {
			selectedScBlocked++
		}
		if blockedByParticipatedSc {
			participatedScBlocked++
		}
		if blockedBySelectedSc || blockedByParticipatedSc {
			blocked++
		}
	}
	return FC2SmartPickInfo{
		RawCandidateMovieCount:          len(unique),
		AfterBlockMovieCount:            len(unique) - blocked,
		SelectedScBlockedMovieCount:     selectedScBlocked,
		ParticipatedScBlockedMovieCount: participatedScBlocked,
	}
}

func collectUniqueFC2SmartPickMovies(groups []*fc2SmartPickGroup) []*FC2SmartPickMovie {
	seen := make(map[int64]*FC2SmartPickMovie)
	for _, group := range groups {
		if group == nil {
			continue
		}
		for _, movie := range group.candidates {
			if movie == nil || movie.FC2Movie == nil || movie.ID <= 0 {
				continue
			}
			if _, ok := seen[movie.ID]; !ok {
				seen[movie.ID] = movie
			}
		}
	}
	out := make([]*FC2SmartPickMovie, 0, len(seen))
	for _, movie := range seen {
		out = append(out, movie)
	}
	return out
}

func pickFC2SmartMovies(groups []*fc2SmartPickGroup, n int, opt FC2SmartPickOptions, now time.Time) ([]*FC2SmartPickMovie, error) {
	for _, group := range groups {
		if group == nil {
			continue
		}
		alive := group.alive[:0]
		for _, index := range group.alive {
			if index >= 0 && index < len(group.candidates) && !fc2SmartMovieBlocked(group.candidates[index], opt, now) {
				alive = append(alive, index)
			}
		}
		group.alive = alive
	}
	if countFC2SmartPickEligible(groups) < n {
		return nil, errors.New("FC2 candidates insufficient after filters and participated/selected SC blocks")
	}

	targets := allocateFC2SmartPickTargets(groups, n)
	selected := make([]*FC2SmartPickMovie, 0, n)
	rnd := rand.New(rand.NewSource(effectiveSeed(opt.RandomSeed)))
	for {
		progress := false
		for index, group := range groups {
			if targets[index] <= 0 || group == nil || len(group.alive) == 0 {
				continue
			}
			movie := pickOneFC2SmartMovie(group, opt, now, rnd)
			if movie == nil {
				continue
			}
			selected = append(selected, movie)
			targets[index]--
			pruneFC2SmartMovie(groups, movie.ID)
			progress = true
		}
		if len(selected) >= n || !progress {
			break
		}
	}
	for len(selected) < n {
		group := chooseFC2SmartPickGroup(groups, rnd)
		if group == nil {
			return nil, errors.New("FC2 candidates insufficient after cross-group deduplication")
		}
		movie := pickOneFC2SmartMovie(group, opt, now, rnd)
		if movie == nil {
			continue
		}
		selected = append(selected, movie)
		pruneFC2SmartMovie(groups, movie.ID)
	}
	return selected, nil
}

func fc2SmartMovieBlocked(movie *FC2SmartPickMovie, opt FC2SmartPickOptions, now time.Time) bool {
	return fc2SmartMovieSelectedScBlocked(movie, opt, now) || fc2SmartMovieParticipatedScBlocked(movie, opt, now)
}

func fc2SmartMovieSelectedScBlocked(movie *FC2SmartPickMovie, opt FC2SmartPickOptions, now time.Time) bool {
	return movie != nil && opt.MovieSelectedScBlockDays > 0 && movie.LastSelectedScTime > 0 && ageInDays(movie.LastSelectedScTime, now) < float64(opt.MovieSelectedScBlockDays)
}

func fc2SmartMovieParticipatedScBlocked(movie *FC2SmartPickMovie, opt FC2SmartPickOptions, now time.Time) bool {
	return movie != nil && opt.MovieParticipatedScBlockDays > 0 && movie.LastParticipatedScTime > 0 && ageInDays(movie.LastParticipatedScTime, now) < float64(opt.MovieParticipatedScBlockDays)
}

func fc2SmartMovieWeight(movie *FC2SmartPickMovie, opt FC2SmartPickOptions, now time.Time) float64 {
	if movie == nil {
		return 0
	}
	weight := 1.0
	if movie.ScTimes > 0 {
		weight *= 1 - opt.MovieHasSelectedScPenaltyAlpha
	}
	if movie.LastSelectedScTime > 0 {
		recentness := math.Exp(-ageInDays(movie.LastSelectedScTime, now) / 90.0)
		weight *= 1 - opt.MovieSelectedScPenaltyAlpha*recentness
	}
	if weight < 0.0001 {
		return 0.0001
	}
	return weight
}

func allocateFC2SmartPickTargets(groups []*fc2SmartPickGroup, n int) []int {
	targets := make([]int, len(groups))
	totalWeight := 0.0
	for _, group := range groups {
		if group != nil && len(group.alive) > 0 {
			totalWeight += group.weight
		}
	}
	type remainder struct {
		index int
		value float64
	}
	remainders := make([]remainder, 0, len(groups))
	assigned := 0
	for index, group := range groups {
		if group == nil || len(group.alive) == 0 {
			continue
		}
		exact := float64(n) * group.weight / totalWeight
		targets[index] = int(math.Floor(exact))
		assigned += targets[index]
		remainders = append(remainders, remainder{index: index, value: exact - float64(targets[index])})
	}
	sort.SliceStable(remainders, func(i, j int) bool { return remainders[i].value > remainders[j].value })
	for index := 0; index < n-assigned; index++ {
		targets[remainders[index%len(remainders)].index]++
	}
	return targets
}

func pickOneFC2SmartMovie(group *fc2SmartPickGroup, opt FC2SmartPickOptions, now time.Time, rnd *rand.Rand) *FC2SmartPickMovie {
	if group == nil || len(group.alive) == 0 {
		return nil
	}
	weights := make([]float64, len(group.alive))
	total := 0.0
	for index, candidateIndex := range group.alive {
		weights[index] = fc2SmartMovieWeight(group.candidates[candidateIndex], opt, now)
		total += weights[index]
	}
	selectedPosition := len(group.alive) - 1
	randomValue := rnd.Float64() * total
	accumulated := 0.0
	for index, weight := range weights {
		accumulated += weight
		if randomValue <= accumulated {
			selectedPosition = index
			break
		}
	}
	candidateIndex := group.alive[selectedPosition]
	movie := group.candidates[candidateIndex]
	group.alive = append(group.alive[:selectedPosition], group.alive[selectedPosition+1:]...)
	return movie
}

func pruneFC2SmartMovie(groups []*fc2SmartPickGroup, movieID int64) {
	for _, group := range groups {
		if group == nil {
			continue
		}
		alive := group.alive[:0]
		for _, index := range group.alive {
			movie := group.candidates[index]
			if movie != nil && movie.ID == movieID {
				continue
			}
			alive = append(alive, index)
		}
		group.alive = alive
	}
}

func countFC2SmartPickEligible(groups []*fc2SmartPickGroup) int {
	seen := make(map[int64]struct{})
	for _, group := range groups {
		if group == nil {
			continue
		}
		for _, index := range group.alive {
			movie := group.candidates[index]
			if movie != nil && movie.ID > 0 {
				seen[movie.ID] = struct{}{}
			}
		}
	}
	return len(seen)
}

func chooseFC2SmartPickGroup(groups []*fc2SmartPickGroup, rnd *rand.Rand) *fc2SmartPickGroup {
	total := 0.0
	for _, group := range groups {
		if group != nil && len(group.alive) > 0 {
			total += group.weight
		}
	}
	if total <= 0 {
		return nil
	}
	randomValue := rnd.Float64() * total
	accumulated := 0.0
	for _, group := range groups {
		if group == nil || len(group.alive) == 0 {
			continue
		}
		accumulated += group.weight
		if randomValue <= accumulated {
			return group
		}
	}
	return nil
}

func FC2SmartPickTotalSize(movies []*FC2SmartPickMovie) int64 {
	var total int64
	for _, movie := range movies {
		if movie != nil {
			total += movie.MediaSize
		}
	}
	return total
}
