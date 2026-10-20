package movie

import (
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
)

func TestBuildRankDayStatOnlyBuildsRequestedDay(t *testing.T) {
	rows := []*moviex.CRank{
		{Id: 1, DayNumber: 1, Category: 1, RankPos: 2, MovieJavId: "A"},
		{Id: 2, DayNumber: 1, Category: 1, RankPos: 4, MovieJavId: "B"},
		{Id: 3, DayNumber: 3, Category: 1, RankPos: 1, MovieJavId: "A"},
		{Id: 4, DayNumber: 3, Category: 1, RankPos: 3, MovieJavId: "C"},
		{Id: 5, DayNumber: 4, Category: 1, RankPos: 2, MovieJavId: "B"},
		{Id: 6, DayNumber: 4, Category: 1, RankPos: 3, MovieJavId: "C"},
	}

	stat, err := buildRankDayStat(rows, dailybesttypes.RankDayTiming{
		DayNumber: 4,
		FetchedOn: 80,
		SyncedOn:  90,
	}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if stat.DayNumber != 4 || stat.PrevDayNumber != 3 {
		t.Fatalf("unexpected day relation: %+v", stat)
	}
	if stat.ContinuedMovieCount != 1 || stat.ReturningMovieCount != 1 || stat.DroppedMovieCount != 1 {
		t.Fatalf("unexpected movie changes: %+v", stat)
	}
	if stat.SourceChecksum == "" || stat.CalculatedOn != 100 || stat.CreatedOn != 100 || stat.UpdatedOn != 100 {
		t.Fatalf("unexpected immutable metadata: %+v", stat)
	}
	if stat.FetchedOn != 80 || stat.SyncedOn != 90 {
		t.Fatalf("unexpected source timing: %+v", stat)
	}
}

func TestBuildRankDayStatDeduplicatesMoviesAcrossCategories(t *testing.T) {
	rows := []*moviex.CRank{
		{Id: 1, DayNumber: 1, Category: 1, RankPos: 5, MovieJavId: "A"},
		{Id: 2, DayNumber: 1, Category: 2, RankPos: 2, MovieJavId: "A"},
		{Id: 3, DayNumber: 2, Category: 1, RankPos: 3, MovieJavId: "A"},
	}
	first, err := buildRankDayStat(rows, dailybesttypes.RankDayTiming{DayNumber: 1}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if first.RankRowCount != 2 || first.MovieCount != 1 || first.FirstSeenMovieCount != 1 {
		t.Fatalf("unexpected first stat: %+v", first)
	}
	second, err := buildRankDayStat(rows, dailybesttypes.RankDayTiming{DayNumber: 2}, 200)
	if err != nil {
		t.Fatal(err)
	}
	if second.ContinuedMovieCount != 1 || second.PrevDayNumber != 1 {
		t.Fatalf("movie continuity was not preserved: %+v", second)
	}
}

func TestBuildRankDayStatRejectsMissingDay(t *testing.T) {
	if _, err := buildRankDayStat(
		[]*moviex.CRank{{DayNumber: 1, MovieJavId: "A"}},
		dailybesttypes.RankDayTiming{DayNumber: 2},
		100,
	); err == nil {
		t.Fatal("missing rank day must fail")
	}
}

func TestBuildRankDayStatFallsBackToRankStoredOn(t *testing.T) {
	stat, err := buildRankDayStat(
		[]*moviex.CRank{
			{DayNumber: 1, Category: 1, RankPos: 1, MovieJavId: "A", CreatedOn: 70},
			{DayNumber: 1, Category: 1, RankPos: 2, MovieJavId: "B", CreatedOn: 72},
		},
		dailybesttypes.RankDayTiming{DayNumber: 1},
		100,
	)
	if err != nil {
		t.Fatal(err)
	}
	if stat.FetchedOn != 70 || stat.SyncedOn != 70 {
		t.Fatalf("unexpected fallback timing: %+v", stat)
	}
}
