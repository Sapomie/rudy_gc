package sc

import (
	"testing"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func TestAggregateFC2MovieScInfo(t *testing.T) {
	movieIDs := []int64{10, 20, 30}
	rows := []*types.GList{
		{FC2MovieId: 10, ScName: "event-a", IsSc: consts.GListIsSc, IsCome: consts.GListIsCome},
		{FC2MovieId: 10, ScName: "event-b", IsSc: 1, IsCome: 1},
		{FC2MovieId: 20, ScName: "event-a", IsSc: consts.GListIsSc, IsCome: 1},
		{FC2MovieId: 20, ScName: "event-c", IsSc: consts.GListIsSc, IsCome: consts.GListIsCome},
	}
	scTimes := map[string]int64{
		"event-a": 100,
		"event-b": 200,
		"event-c": 300,
	}

	got := aggregateFC2MovieScInfo(movieIDs, rows, scTimes)
	if info := got[10]; info.ScEventTimes != 2 || info.ScTimes != 1 || info.ComeTimes != 1 || info.LastSelectedScTime != 100 || info.LastParticipatedScTime != 200 {
		t.Fatalf("movie 10 stats = %#v", info)
	}
	if info := got[20]; info.ScEventTimes != 2 || info.ScTimes != 2 || info.ComeTimes != 1 || info.LastSelectedScTime != 300 || info.LastParticipatedScTime != 300 {
		t.Fatalf("movie 20 stats = %#v", info)
	}
	if info := got[30]; info != (movieScInfo{}) {
		t.Fatalf("movie 30 stats should be cleared: %#v", info)
	}
}

func TestSummarizeScEventMovies(t *testing.T) {
	rows := []*types.GList{
		{MovieType: "jav", IsSc: consts.GListIsSc},
		{MovieType: "jav", IsSc: consts.GListIsNotSc},
		{MovieType: "fc2ppv", IsSc: consts.GListIsSc, IsCome: consts.GListIsCome},
		{MovieType: "fc2ppv", IsSc: consts.GListIsNotSc},
		nil,
	}

	got := summarizeScEventMovies(rows)
	if got.TotalMovieCount != 4 || got.ScMovieCount != 2 || got.NoScMovieCount != 2 {
		t.Fatalf("SC event counts = %#v", got)
	}
	if got.JavMovieCount != 2 || got.FC2MovieCount != 2 || got.ComeMovieType != "fc2ppv" {
		t.Fatalf("SC event sources = %#v", got)
	}
}

func TestEventDetailMovieSourceFilters(t *testing.T) {
	rows := []*types.GList{
		{Id: 1, MovieType: "jav"},
		{Id: 2, MovieType: "fc2ppv"},
		{Id: 3, MovieType: "jav"},
	}

	javRows := filterEventDetailMoviesByMovieType(rows, "jav")
	if len(javRows) != 2 || javRows[0].Id != 1 || javRows[1].Id != 3 {
		t.Fatalf("JAV rows = %#v", javRows)
	}
	fc2Rows := filterEventDetailMoviesByMovieType(rows, "fc2ppv")
	if len(fc2Rows) != 1 || fc2Rows[0].Id != 2 {
		t.Fatalf("FC2 rows = %#v", fc2Rows)
	}
	if got := normalizeEventDetailMovieFilter("jav"); got != eventDetailMovieFilterJAV {
		t.Fatalf("JAV filter = %q", got)
	}
	if got := normalizeEventDetailMovieFilter("fc2"); got != eventDetailMovieFilterFC2 {
		t.Fatalf("FC2 filter = %q", got)
	}
}
