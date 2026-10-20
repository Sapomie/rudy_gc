package dailybestsync

import (
	"context"
	"testing"

	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type fakeClient struct {
	day   contracts.DailyBestRankDay
	ranks []contracts.DailyBestRank
}

func (f *fakeClient) ListRankDays(context.Context, dailybesttypes.RankDayQuery) (contracts.DailyBestRankDayListResponse, error) {
	return contracts.DailyBestRankDayListResponse{Days: []contracts.DailyBestRankDay{f.day}}, nil
}
func (f *fakeClient) ListRanks(context.Context, int64, int64) (contracts.DailyBestRankListResponse, error) {
	return contracts.DailyBestRankListResponse{Ranks: f.ranks}, nil
}

type fakeStore struct {
	imported  []int64
	processed []int64
	state     dailybesttypes.LocalRankCategoryState
}

func (f *fakeStore) ListLocalRankCategoryStates(
	_ context.Context,
	queries []dailybesttypes.LocalRankCategoryQuery,
) (map[dailybesttypes.RankCategoryKey]dailybesttypes.LocalRankCategoryState, error) {
	states := make(map[dailybesttypes.RankCategoryKey]dailybesttypes.LocalRankCategoryState, len(queries))
	for _, query := range queries {
		states[dailybesttypes.RankCategoryKey{DayNumber: query.DayNumber, Category: query.Category}] = f.state
	}
	return states, nil
}
func (f *fakeStore) SaveImported(_ context.Context, manifest dailybesttypes.SyncManifest) error {
	f.imported = append(f.imported, manifest.FetchRunID)
	return nil
}
func (f *fakeStore) MarkProcessed(_ context.Context, manifests []dailybesttypes.SyncManifest) error {
	for _, manifest := range manifests {
		f.processed = append(f.processed, manifest.FetchRunID)
	}
	return nil
}
func (*fakeStore) MarkFailed(context.Context, dailybesttypes.SyncManifest, string) error {
	return nil
}

func TestSyncValidatesImportsAndCompletesBatch(t *testing.T) {
	dayNumber, _ := contracts.RankDayNumber("2026-07-19")
	ranks := []contracts.DailyBestRank{{RankKey: "2026-07-19_1_001", MovieJavID: "jav-1", DayNumber: dayNumber, RankPos: 1, Category: 1}}
	checksum := contracts.ComputeDailyBestRankChecksum("2026-07-19", dayNumber, 1, ranks)
	day := contracts.DailyBestRankDay{
		BusinessDate: "2026-07-19", DayNumber: dayNumber, RankCount: 1,
		Categories: []contracts.DailyBestRankDayCategory{{
			FetchRunID: 7, Category: 1, PageCount: 1, ItemCount: 1, RankCount: 1, Checksum: checksum,
		}},
	}
	store := &fakeStore{}
	service := New(&fakeClient{day: day, ranks: ranks}, store)
	applied := 0
	result, err := service.Sync(context.Background(), []dailybesttypes.SyncTarget{{DayNumber: dayNumber, Category: 1}}, func(_ context.Context, got Batch) error {
		applied++
		if got.Manifest.FetchRunID != 7 {
			t.Fatalf("fetch_run_id=%d", got.Manifest.FetchRunID)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if applied != 1 || len(store.imported) != 1 || len(result.Batches) != 1 {
		t.Fatalf("unexpected sync result: applied=%d imported=%v result=%+v", applied, store.imported, result)
	}
	if err := service.MarkProcessed(context.Background(), result); err != nil {
		t.Fatal(err)
	}
	if len(store.processed) != 1 || store.processed[0] != 7 {
		t.Fatalf("processed=%v", store.processed)
	}
}

func TestListRankDaysMarksMissingLocalRankAsSyncable(t *testing.T) {
	dayNumber, _ := contracts.RankDayNumber("2026-07-19")
	day := contracts.DailyBestRankDay{
		BusinessDate: "2026-07-19", DayNumber: dayNumber, RankCount: 465,
		Categories: []contracts.DailyBestRankDayCategory{{
			FetchRunID: 7, Category: 1, RankCount: 465, Checksum: "remote-checksum",
		}},
	}
	service := New(&fakeClient{day: day}, &fakeStore{})
	result, err := service.ListRankDays(context.Background(), dailybesttypes.RankDayQuery{Page: 1, PageSize: 24})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Days) != 1 || !result.Days[0].NeedsSync || result.Days[0].Status != dailybesttypes.RankDayStatusMissing {
		t.Fatalf("result=%+v", result)
	}
}

func TestSyncSkipsSelectedDayWhenLocalRanksMatch(t *testing.T) {
	dayNumber, _ := contracts.RankDayNumber("2026-07-19")
	ranks := []contracts.DailyBestRank{{MovieJavID: "jav-1", DayNumber: dayNumber, RankPos: 1, Category: 1}}
	checksum := contracts.ComputeDailyBestRankChecksum("2026-07-19", dayNumber, 1, ranks)
	day := contracts.DailyBestRankDay{
		BusinessDate: "2026-07-19", DayNumber: dayNumber, RankCount: 1,
		Categories: []contracts.DailyBestRankDayCategory{{Category: 1, RankCount: 1, Checksum: checksum}},
	}
	store := &fakeStore{state: dailybesttypes.LocalRankCategoryState{RankCount: 1, Checksum: checksum}}
	service := New(&fakeClient{day: day, ranks: ranks}, store)
	result, err := service.Sync(
		context.Background(),
		[]dailybesttypes.SyncTarget{{DayNumber: dayNumber, Category: 1}},
		func(context.Context, Batch) error {
			t.Fatal("matching rank day must be skipped")
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Batches) != 0 || len(store.imported) != 0 {
		t.Fatalf("result=%+v imported=%v", result, store.imported)
	}
}

func TestSyncDoesNotOverwriteExistingRanksWhenRemoteChecksumDiffers(t *testing.T) {
	dayNumber, _ := contracts.RankDayNumber("2026-07-19")
	ranks := []contracts.DailyBestRank{{
		MovieJavID: "jav-remote",
		DayNumber:  dayNumber,
		RankPos:    1,
		Category:   1,
	}}
	day := contracts.DailyBestRankDay{
		BusinessDate: "2026-07-19",
		DayNumber:    dayNumber,
		RankCount:    1,
		Categories: []contracts.DailyBestRankDayCategory{{
			Category:  1,
			RankCount: 1,
			Checksum:  contracts.ComputeDailyBestRankChecksum("2026-07-19", dayNumber, 1, ranks),
		}},
	}
	store := &fakeStore{state: dailybesttypes.LocalRankCategoryState{
		RankCount: 1,
		Checksum:  "local-immutable-checksum",
	}}
	service := New(&fakeClient{day: day, ranks: ranks}, store)
	result, err := service.Sync(
		context.Background(),
		[]dailybesttypes.SyncTarget{{DayNumber: dayNumber, Category: 1}},
		func(context.Context, Batch) error {
			t.Fatal("existing rank day must never be overwritten")
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Batches) != 0 || len(store.imported) != 0 {
		t.Fatalf("result=%+v imported=%v", result, store.imported)
	}
}
