package dailybest

import "testing"

func TestComputeDailyBestSnapshotChecksumStableOrdering(t *testing.T) {
	left := DailyBestSnapshotResponse{
		DailyBestSnapshotManifest: DailyBestSnapshotManifest{BusinessDate: "2026-07-19", Category: 1},
		Pages: []DailyBestSnapshotPage{
			{PageNumber: 2, SourceURL: "page-2", Content: "two"},
			{PageNumber: 1, SourceURL: "page-1", Content: "one"},
		},
		Items: []DailyBestSnapshotItem{
			{RankPos: 2, SourcePage: 1, SourceIndex: 2, JavID: "b", MovieName: "B", MovieCode: "B"},
			{RankPos: 1, SourcePage: 1, SourceIndex: 1, JavID: "a", MovieName: "A", MovieCode: "A"},
		},
	}
	right := left
	right.Pages = []DailyBestSnapshotPage{left.Pages[1], left.Pages[0]}
	right.Items = []DailyBestSnapshotItem{left.Items[1], left.Items[0]}
	if got, want := ComputeDailyBestSnapshotChecksum(left), ComputeDailyBestSnapshotChecksum(right); got != want {
		t.Fatalf("checksum ordering mismatch: %s != %s", got, want)
	}
}
