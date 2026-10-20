package spider

import (
	"reflect"
	"testing"

	"rudy_gc/local/api/internal/types"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func TestBestinvAndRemoteRanksNormalizeToSameBatch(t *testing.T) {
	dayNumber, err := contracts.RankDayNumber("2026-07-21")
	if err != nil {
		t.Fatal(err)
	}
	bestinvs := []*types.Bestinv{
		{
			Id: 2, Date: "2026-07-21", DayNumber: dayNumber, Category: 1, Page: 2,
			Content: bestinvRankTestHTML(
				bestinvRankTestItem{javID: "CCC-003", title: "影片 C"},
			),
		},
		{
			Id: 1, Date: "2026-07-21", DayNumber: dayNumber, Category: 1, Page: 1,
			Content: bestinvRankTestHTML(
				bestinvRankTestItem{javID: "AAA-001", title: "影片 A"},
				bestinvRankTestItem{javID: "BLU-001", title: "影片（ブルーレイディスク）"},
				bestinvRankTestItem{javID: "BLU-002", title: "影片蓝光版"},
				bestinvRankTestItem{javID: "BBB-002", title: "影片 B"},
			),
		},
	}

	pending, err := buildPendingBestinvRankBatches(bestinvs)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending batches=%d", len(pending))
	}
	localBatches, err := normalizeDailyBestRankBatches([]dailybesttypes.RankBatch{pending[0].Batch})
	if err != nil {
		t.Fatal(err)
	}
	remoteBatches, err := normalizeDailyBestRankBatches([]dailybesttypes.RankBatch{{
		BusinessDate: "2026-07-21",
		DayNumber:    dayNumber,
		Category:     1,
		Ranks: []contracts.DailyBestRank{
			{MovieJavID: "CCC-003", DayNumber: dayNumber, RankPos: 21, Category: 1},
			{MovieJavID: "BBB-002", DayNumber: dayNumber, RankPos: 2, Category: 1},
			{MovieJavID: "AAA-001", DayNumber: dayNumber, RankPos: 1, Category: 1},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(localBatches, remoteBatches) {
		t.Fatalf("local=%+v remote=%+v", localBatches, remoteBatches)
	}
}

func TestBuildPendingBestinvRankBatchesRejectsMissingPage(t *testing.T) {
	dayNumber, err := contracts.RankDayNumber("2026-07-21")
	if err != nil {
		t.Fatal(err)
	}
	_, err = buildPendingBestinvRankBatches([]*types.Bestinv{{
		Id: 2, Date: "2026-07-21", DayNumber: dayNumber, Category: 1, Page: 2,
		Content: bestinvRankTestHTML(bestinvRankTestItem{javID: "AAA-001", title: "影片 A"}),
	}})
	if err == nil {
		t.Fatal("missing first page must fail")
	}
}

func TestNormalizeDailyBestRankBatchesRejectsDuplicateRank(t *testing.T) {
	dayNumber, err := contracts.RankDayNumber("2026-07-21")
	if err != nil {
		t.Fatal(err)
	}
	_, err = normalizeDailyBestRankBatches([]dailybesttypes.RankBatch{{
		BusinessDate: "2026-07-21",
		DayNumber:    dayNumber,
		Category:     1,
		Ranks: []contracts.DailyBestRank{
			{MovieJavID: "AAA-001", RankPos: 1},
			{MovieJavID: "BBB-002", RankPos: 1},
		},
	}})
	if err == nil {
		t.Fatal("duplicate rank position must fail")
	}
}

type bestinvRankTestItem struct {
	javID string
	title string
}

func bestinvRankTestHTML(items ...bestinvRankTestItem) string {
	html := `<div class="videos">`
	for _, item := range items {
		html += `<div class="video"><div class="title">` + item.title + `</div><div class="toolbar"><a id="` + item.javID + `"></a></div></div>`
	}
	return html + `</div>`
}
