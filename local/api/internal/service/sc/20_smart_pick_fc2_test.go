package sc

import (
	"strings"
	"testing"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

func TestNormalizeFC2SmartPickRequestBuildsInclusiveDateRanges(t *testing.T) {
	req, err := normalizeFC2SmartPickRequest(FC2SmartPickRequest{
		PostTimeStart:       "2026-07-01",
		PostTimeEnd:         "2026-07-02",
		MediaBirthTimeStart: "2026-06-01",
		MediaBirthTimeEnd:   "2026-06-30",
	})
	if err != nil {
		t.Fatalf("normalize request: %v", err)
	}
	start, _ := time.ParseInLocation("2006-01-02", "2026-07-01", time.Local)
	end, _ := time.ParseInLocation("2006-01-02", "2026-07-03", time.Local)
	if req.postTimeStart != start.Unix() || req.postTimeEnd != end.Add(-time.Second).Unix() {
		t.Fatalf("post range = %d..%d", req.postTimeStart, req.postTimeEnd)
	}
	if req.mediaFilter.BirthTimeStart <= 0 || req.mediaFilter.BirthTimeEnd <= req.mediaFilter.BirthTimeStart {
		t.Fatalf("media range = %d..%d", req.mediaFilter.BirthTimeStart, req.mediaFilter.BirthTimeEnd)
	}
}

func TestMatchesFC2SmartPickRequestUsesMovieAndStatFields(t *testing.T) {
	scMin, scMax := int64(1), int64(3)
	comeMin, comeMax := int64(2), int64(4)
	normalized, err := normalizeFC2SmartPickRequest(FC2SmartPickRequest{
		FC2Key:        "4556413",
		Title:         "sample",
		PerformerText: "演员",
		MaskType:      "mosaic",
		ScTimesMin:    &scMin,
		ScTimesMax:    &scMax,
		ComeTimesMin:  &comeMin,
		ComeTimesMax:  &comeMax,
	})
	if err != nil {
		t.Fatal(err)
	}
	movie := &moviex.FC2Movie{ID: 1, FC2Key: "FC2PPV-4556413-1", Title: "Sample Title", PerformerText: "演员甲", MaskType: "MOSAIC"}
	stat := &moviex.FC2ScStat{FC2MovieID: 1, ScTimes: 2, ComeTimes: 3}
	if !matchesFC2SmartPickRequest(normalized, movie, stat) {
		t.Fatal("matching FC2 movie was filtered out")
	}
	stat.ScTimes = 4
	if matchesFC2SmartPickRequest(normalized, movie, stat) {
		t.Fatal("SC max filter was ignored")
	}
}

func TestPickFC2SmartMoviesBlocksRecentAndDeduplicatesAcrossGroups(t *testing.T) {
	now := time.Now()
	shared := testFC2SmartPickMovie(1, 0)
	second := testFC2SmartPickMovie(2, 0)
	third := testFC2SmartPickMovie(3, 0)
	blockedBySelectedSc := testFC2SmartPickMovie(4, now.Add(-24*time.Hour).Unix())
	blockedByParticipatedSc := testFC2SmartPickMovieWithTimes(5, 0, now.Add(-24*time.Hour).Unix())
	groups := []*fc2SmartPickGroup{
		testFC2SmartPickGroup(2, shared, second, blockedBySelectedSc),
		testFC2SmartPickGroup(1, shared, third, blockedByParticipatedSc),
	}
	selected, err := pickFC2SmartMovies(groups, 3, FC2SmartPickOptions{MovieSelectedScBlockDays: 7, MovieParticipatedScBlockDays: 7, RandomSeed: 7}, now)
	if err != nil {
		t.Fatalf("pick FC2 movies: %v", err)
	}
	if len(selected) != 3 {
		t.Fatalf("selected = %d", len(selected))
	}
	seen := make(map[int64]struct{}, len(selected))
	for _, movie := range selected {
		if movie.ID == blockedBySelectedSc.ID || movie.ID == blockedByParticipatedSc.ID {
			t.Fatal("recently SC blocked movie was selected")
		}
		if _, ok := seen[movie.ID]; ok {
			t.Fatalf("duplicate movie selected: %d", movie.ID)
		}
		seen[movie.ID] = struct{}{}
	}
}

func TestFC2SmartMovieBlockRules(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.Local)
	recent := now.Add(-24 * time.Hour).Unix()
	old := now.Add(-31 * 24 * time.Hour).Unix()

	tests := []struct {
		name  string
		movie *FC2SmartPickMovie
		opt   FC2SmartPickOptions
		want  bool
	}{
		{name: "selected SC blocks independently", movie: testFC2SmartPickMovieWithTimes(1, recent, old), opt: FC2SmartPickOptions{MovieSelectedScBlockDays: 30}, want: true},
		{name: "participated SC blocks independently", movie: testFC2SmartPickMovieWithTimes(2, old, recent), opt: FC2SmartPickOptions{MovieParticipatedScBlockDays: 30}, want: true},
		{name: "zero disables both blocks", movie: testFC2SmartPickMovieWithTimes(3, recent, recent), opt: FC2SmartPickOptions{}, want: false},
		{name: "old timestamps remain eligible", movie: testFC2SmartPickMovieWithTimes(4, old, old), opt: FC2SmartPickOptions{MovieSelectedScBlockDays: 30, MovieParticipatedScBlockDays: 30}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fc2SmartMovieBlocked(tt.movie, tt.opt, now); got != tt.want {
				t.Fatalf("blocked = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalyzeFC2SmartPickInfoCountsBlockReasonsAndUnion(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.Local)
	recent := now.Add(-24 * time.Hour).Unix()
	old := now.Add(-31 * 24 * time.Hour).Unix()
	groups := []*fc2SmartPickGroup{testFC2SmartPickGroup(1,
		testFC2SmartPickMovieWithTimes(1, recent, old),
		testFC2SmartPickMovieWithTimes(2, old, recent),
		testFC2SmartPickMovieWithTimes(3, recent, recent),
		testFC2SmartPickMovieWithTimes(4, old, old),
	)}

	info := analyzeFC2SmartPickInfo(groups, FC2SmartPickOptions{MovieSelectedScBlockDays: 30, MovieParticipatedScBlockDays: 30}, now)
	if info.RawCandidateMovieCount != 4 || info.AfterBlockMovieCount != 1 || info.SelectedScBlockedMovieCount != 2 || info.ParticipatedScBlockedMovieCount != 2 {
		t.Fatalf("pick info = %#v", info)
	}
}

func TestValidateFC2SmartPickOptionsRejectsNegativeBlockDays(t *testing.T) {
	for _, opt := range []FC2SmartPickOptions{
		{MovieSelectedScBlockDays: -1},
		{MovieParticipatedScBlockDays: -1},
	} {
		if err := validateFC2SmartPickOptions(opt); err == nil {
			t.Fatalf("negative block option accepted: %#v", opt)
		}
	}
}

func TestSmartPickFC2CopyFileName(t *testing.T) {
	movie := &FC2SmartPickMovie{FC2Movie: &moviex.FC2Movie{FC2Key: "FC2PPV-1", PerformerText: "演员/甲", Title: "标题"}}
	name := smartPickFC2CopyFileName(movie, "/media/source.mp4")
	if name != "FC2PPV-1_演员甲_标题.mp4" {
		t.Fatalf("copy file name = %q", name)
	}
	if strings.ContainsAny(name, "/\\") {
		t.Fatalf("copy file name contains path separator: %q", name)
	}
}

func testFC2SmartPickMovie(id, lastSelectedScTime int64) *FC2SmartPickMovie {
	return testFC2SmartPickMovieWithTimes(id, lastSelectedScTime, 0)
}

func testFC2SmartPickMovieWithTimes(id, lastSelectedScTime, lastParticipatedScTime int64) *FC2SmartPickMovie {
	return &FC2SmartPickMovie{
		FC2Movie:               &moviex.FC2Movie{ID: id, FC2Key: "FC2PPV"},
		LastSelectedScTime:     lastSelectedScTime,
		LastParticipatedScTime: lastParticipatedScTime,
	}
}

func testFC2SmartPickGroup(weight float64, movies ...*FC2SmartPickMovie) *fc2SmartPickGroup {
	alive := make([]int, len(movies))
	for index := range alive {
		alive[index] = index
	}
	return &fc2SmartPickGroup{weight: weight, candidates: movies, alive: alive}
}
