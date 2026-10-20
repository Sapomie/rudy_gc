package movie

import (
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

func TestBuildRankPeriodStat(t *testing.T) {
	spec := rankStatPeriodSpec{
		PeriodType:     2,
		PeriodKey:      "2026-07",
		StartDayNumber: 1,
		EndDayNumber:   3,
	}
	sourceRows := []*moviex.CRank{
		{DayNumber: 1, Category: 1, RankPos: 1, MovieJavId: "A", CreatedOn: 10, UpdatedOn: 11},
		{DayNumber: 3, Category: 2, RankPos: 2, MovieJavId: "A", CreatedOn: 20, UpdatedOn: 21},
		{DayNumber: 3, Category: 1, RankPos: 3, MovieJavId: "B", CreatedOn: 15, UpdatedOn: 16},
	}
	dayStats := []*moviex.CRankDayStat{
		{DayNumber: 1, RankRowCount: 2, MovieCount: 1, FirstSeenMovieCount: 1, SourceChecksum: "day-1"},
		{DayNumber: 3, RankRowCount: 3, MovieCount: 2, FirstSeenMovieCount: 1, ReturningMovieCount: 1, SourceChecksum: "day-3"},
	}

	stat := buildRankPeriodStat(spec, sourceRows, dayStats, 30)
	if stat.PeriodType != spec.PeriodType || stat.PeriodKey != spec.PeriodKey || stat.StartDayNumber != 1 || stat.EndDayNumber != 3 {
		t.Fatalf("unexpected period dimensions: %+v", stat)
	}
	if stat.SourceRankRowCount != 5 || stat.SourceRankDayCount != 2 || stat.SourceMovieCount != 2 || stat.MissingRankDayCount != 1 {
		t.Fatalf("unexpected source stats: %+v", stat)
	}
	if stat.PeriodMovieCount != 3 || stat.NewEntryMovieCount != 2 {
		t.Fatalf("unexpected period movie stats: %+v", stat)
	}
	if stat.SourceFetchedStartOn != 10 || stat.SourceFetchedEndOn != 20 || stat.CalculatedOn != 30 || stat.SourceChecksum == "" {
		t.Fatalf("unexpected metadata: %+v", stat)
	}
}

func TestRankStatPeriodSpecsForDaysOnlySelectsAffectedPeriods(t *testing.T) {
	specs := []rankStatPeriodSpec{
		{PeriodType: 2, PeriodKey: "old", StartDayNumber: 1, EndDayNumber: 10},
		{PeriodType: 2, PeriodKey: "current", StartDayNumber: 11, EndDayNumber: 20},
		{PeriodType: 3, PeriodKey: "wide", StartDayNumber: 1, EndDayNumber: 20},
	}
	selected := rankStatPeriodSpecsForDays(specs, []int64{15})
	if len(selected) != 2 || selected[0].PeriodKey != "current" || selected[1].PeriodKey != "wide" {
		t.Fatalf("unexpected selected specs: %+v", selected)
	}
}

func TestRankPeriodSpecContainsAnyDay(t *testing.T) {
	spec := rankPeriodSpec{StartDayNumber: 11, EndDayNumber: 20}
	if !rankPeriodSpecContainsAnyDay(spec, []int64{5, 15, 25}) {
		t.Fatal("affected day in period was not detected")
	}
	if rankPeriodSpecContainsAnyDay(spec, []int64{5, 25}) {
		t.Fatal("unrelated historical period must not be refreshed")
	}
}
