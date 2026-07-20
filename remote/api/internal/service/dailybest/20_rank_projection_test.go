package dailybest

import (
	"testing"
	"time"

	contracts "rudy_gc/shared/dailybest"
)

func TestBuildRankProjectionRebuildsDayAndPeriodData(t *testing.T) {
	dayOne, _ := contracts.RankDayNumber("2026-07-19")
	dayTwo, _ := contracts.RankDayNumber("2026-07-20")
	ranks := []contracts.DailyBestRank{
		{MovieJavID: "movie-b", DayNumber: dayOne, RankPos: 1, Category: 1, CreatedOn: 10},
		{MovieJavID: "movie-a", DayNumber: dayOne, RankPos: 2, Category: 1, CreatedOn: 10},
		{MovieJavID: "movie-a", DayNumber: dayTwo, RankPos: 1, Category: 1, CreatedOn: 20},
		{MovieJavID: "movie-c", DayNumber: dayTwo, RankPos: 2, Category: 1, CreatedOn: 20},
	}
	projection := buildRankProjection(ranks, 30, time.UTC)
	if len(projection.DayStats) != 2 {
		t.Fatalf("day stats=%d", len(projection.DayStats))
	}
	latest := projection.DayStats[1]
	if latest.ContinuedMovieCount != 1 || latest.FirstSeenMovieCount != 1 || latest.DroppedMovieCount != 1 || latest.RiseMovieCount != 1 {
		t.Fatalf("unexpected latest day stat: %+v", latest)
	}
	if len(projection.Periods) != 5 {
		t.Fatalf("period projections=%d", len(projection.Periods))
	}
	for _, period := range projection.Periods {
		if len(period.Items) == 0 || period.Stat.PeriodMovieCount != int64(len(period.Items)) || period.Stat.SourceChecksum == "" {
			t.Fatalf("invalid period projection: %+v", period)
		}
	}
}
