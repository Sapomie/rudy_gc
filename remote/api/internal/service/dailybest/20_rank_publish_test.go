package dailybest

import (
	"context"
	"testing"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type rankPublishRepositoryStub struct {
	Repository
	ranks          []contracts.DailyBestRank
	published      bool
	noChangeMarked bool
	publishedStat  contracts.DailyBestRankDayStat
}

func (r *rankPublishRepositoryStub) ListAllRanks(context.Context) ([]contracts.DailyBestRank, error) {
	return append([]contracts.DailyBestRank(nil), r.ranks...), nil
}

func (r *rankPublishRepositoryStub) MarkFetchRunNoChange(
	context.Context,
	int64,
	int64,
	int64,
	int64,
	string,
	[]dailybesttypes.MovieDraft,
) error {
	r.noChangeMarked = true
	return nil
}

func (r *rankPublishRepositoryStub) PublishRanks(
	_ context.Context,
	_ int64,
	_ dailybesttypes.RankDraft,
	stat contracts.DailyBestRankDayStat,
	_ int64,
) error {
	r.published = true
	r.publishedStat = stat
	return nil
}

func TestBuildRankDayStatOnlyReturnsRequestedDay(t *testing.T) {
	dayOne, _ := contracts.RankDayNumber("2026-07-19")
	dayTwo, _ := contracts.RankDayNumber("2026-07-20")
	ranks := []contracts.DailyBestRank{
		{MovieJavID: "movie-b", DayNumber: dayOne, RankPos: 1, Category: 1, CreatedOn: 10},
		{MovieJavID: "movie-a", DayNumber: dayOne, RankPos: 2, Category: 1, CreatedOn: 10},
		{MovieJavID: "movie-a", DayNumber: dayTwo, RankPos: 1, Category: 1, CreatedOn: 20},
		{MovieJavID: "movie-c", DayNumber: dayTwo, RankPos: 2, Category: 1, CreatedOn: 20},
	}
	stat, err := buildRankDayStat(ranks, dayTwo, 25, 28, 30)
	if err != nil {
		t.Fatal(err)
	}
	if stat.DayNumber != dayTwo || stat.PrevDayNumber != dayOne {
		t.Fatalf("unexpected day stat: %+v", stat)
	}
	if stat.ContinuedMovieCount != 1 || stat.FirstSeenMovieCount != 1 || stat.ReturningMovieCount != 0 || stat.DroppedMovieCount != 1 {
		t.Fatalf("unexpected latest day stat: %+v", stat)
	}
	if stat.CalculatedOn != 30 || stat.FetchedAt != 25 || stat.SyncedAt != 28 || stat.CreatedOn != 30 || stat.UpdatedOn != 30 {
		t.Fatalf("unexpected timestamps: %+v", stat)
	}
}

func TestBuildRankDayStatFallsBackToRankStoredOn(t *testing.T) {
	dayNumber, _ := contracts.RankDayNumber("2026-07-20")
	stat, err := buildRankDayStat([]contracts.DailyBestRank{{
		MovieJavID: "movie-a", DayNumber: dayNumber, RankPos: 1, Category: 1, CreatedOn: 20,
	}}, dayNumber, 0, 0, 30)
	if err != nil {
		t.Fatal(err)
	}
	if stat.FetchedAt != 20 || stat.SyncedAt != 20 {
		t.Fatalf("unexpected fallback timestamps: %+v", stat)
	}
}

func TestPublishRanksKeepsExistingDayImmutable(t *testing.T) {
	dayNumber, _ := contracts.RankDayNumber("2026-07-20")
	existing := []contracts.DailyBestRank{{
		MovieJavID: "movie-a", DayNumber: dayNumber, RankPos: 1, Category: 1,
	}}
	repository := &rankPublishRepositoryStub{ranks: existing}
	service := &Service{repository: repository}
	draft := dailybesttypes.RankDraft{
		BusinessDate: "2026-07-20",
		DayNumber:    dayNumber,
		Category:     1,
		PageCount:    1,
		Ranks:        append([]contracts.DailyBestRank(nil), existing...),
	}
	draft.Checksum = contracts.ComputeDailyBestRankChecksum(
		draft.BusinessDate,
		draft.DayNumber,
		draft.Category,
		draft.Ranks,
	)
	if err := service.publishRanks(context.Background(), 1, draft, 100); err != nil {
		t.Fatal(err)
	}
	if !repository.noChangeMarked || repository.published {
		t.Fatalf("no_change=%v published=%v", repository.noChangeMarked, repository.published)
	}

	draft.Ranks[0].MovieJavID = "movie-b"
	draft.Checksum = contracts.ComputeDailyBestRankChecksum(
		draft.BusinessDate,
		draft.DayNumber,
		draft.Category,
		draft.Ranks,
	)
	if err := service.publishRanks(context.Background(), 2, draft, 200); err == nil {
		t.Fatal("expected immutable rank conflict")
	}
	if repository.published {
		t.Fatal("conflicting existing day must not be published")
	}
}

func TestPublishRanksCreatesOnlyNewDayStat(t *testing.T) {
	dayOne, _ := contracts.RankDayNumber("2026-07-19")
	dayTwo, _ := contracts.RankDayNumber("2026-07-20")
	repository := &rankPublishRepositoryStub{ranks: []contracts.DailyBestRank{{
		MovieJavID: "movie-a", DayNumber: dayOne, RankPos: 1, Category: 1,
	}}}
	service := &Service{repository: repository}
	draft := dailybesttypes.RankDraft{
		BusinessDate: "2026-07-20",
		DayNumber:    dayTwo,
		Category:     1,
		PageCount:    1,
		Ranks: []contracts.DailyBestRank{
			{MovieJavID: "movie-a", DayNumber: dayTwo, RankPos: 1, Category: 1},
			{MovieJavID: "movie-b", DayNumber: dayTwo, RankPos: 2, Category: 1},
		},
	}
	draft.Checksum = contracts.ComputeDailyBestRankChecksum(
		draft.BusinessDate,
		draft.DayNumber,
		draft.Category,
		draft.Ranks,
	)
	if err := service.publishRanks(context.Background(), 2, draft, 200); err != nil {
		t.Fatal(err)
	}
	if !repository.published || repository.noChangeMarked {
		t.Fatalf("published=%v no_change=%v", repository.published, repository.noChangeMarked)
	}
	if repository.publishedStat.DayNumber != dayTwo || repository.publishedStat.PrevDayNumber != dayOne {
		t.Fatalf("unexpected published stat: %+v", repository.publishedStat)
	}
	if repository.publishedStat.FetchedAt != 200 || repository.publishedStat.SyncedAt != 200 {
		t.Fatalf("unexpected published timing: %+v", repository.publishedStat)
	}
}
