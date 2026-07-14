package sc

import (
	"strings"
	"testing"
	"time"

	"rudy_gc/internal/repository/moviex"
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
	blocked := testFC2SmartPickMovie(4, now.Add(-24*time.Hour).Unix())
	groups := []*fc2SmartPickGroup{
		testFC2SmartPickGroup(2, shared, second, blocked),
		testFC2SmartPickGroup(1, shared, third),
	}
	selected, err := pickFC2SmartMovies(groups, 3, FC2SmartPickOptions{LastScBlockDays: 7, RandomSeed: 7}, now)
	if err != nil {
		t.Fatalf("pick FC2 movies: %v", err)
	}
	if len(selected) != 3 {
		t.Fatalf("selected = %d", len(selected))
	}
	seen := make(map[int64]struct{}, len(selected))
	for _, movie := range selected {
		if movie.ID == blocked.ID {
			t.Fatal("recently SC movie was selected")
		}
		if _, ok := seen[movie.ID]; ok {
			t.Fatalf("duplicate movie selected: %d", movie.ID)
		}
		seen[movie.ID] = struct{}{}
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

func testFC2SmartPickMovie(id, lastScTime int64) *FC2SmartPickMovie {
	return &FC2SmartPickMovie{FC2Movie: &moviex.FC2Movie{ID: id, FC2Key: "FC2PPV"}, LastScTime: lastScTime}
}

func testFC2SmartPickGroup(weight float64, movies ...*FC2SmartPickMovie) *fc2SmartPickGroup {
	alive := make([]int, len(movies))
	for index := range alive {
		alive[index] = index
	}
	return &fc2SmartPickGroup{weight: weight, candidates: movies, alive: alive}
}
