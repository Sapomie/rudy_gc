package dailybest

import "testing"

func TestComputeDailyBestRankChecksumStableOrdering(t *testing.T) {
	left := []DailyBestRank{{RankPos: 2, MovieJavID: "b"}, {RankPos: 1, MovieJavID: "a"}}
	right := []DailyBestRank{left[1], left[0]}
	if got, want := ComputeDailyBestRankChecksum("2026-07-19", 1296, 1, left), ComputeDailyBestRankChecksum("2026-07-19", 1296, 1, right); got != want {
		t.Fatalf("checksum ordering mismatch: %s != %s", got, want)
	}
}

func TestComputeDailyBestRankChecksumIgnoresMovieBasicChanges(t *testing.T) {
	left := []DailyBestRank{{
		RankPos: 1, MovieJavID: "a",
		Movie: DailyBestMovieBasic{MovieName: "A", Title: "旧标题", CoverURL: "https://example.com/old.jpg"},
	}}
	right := []DailyBestRank{{
		RankPos: 1, MovieJavID: "a",
		Movie: DailyBestMovieBasic{MovieName: "A", Title: "新标题", CoverURL: "https://example.com/new.jpg"},
	}}
	if got, want := ComputeDailyBestRankChecksum("2026-07-19", 1296, 1, left), ComputeDailyBestRankChecksum("2026-07-19", 1296, 1, right); got != want {
		t.Fatalf("movie basic changes affected rank checksum: %s != %s", got, want)
	}
}

func TestRankDateRoundTrip(t *testing.T) {
	day, err := RankDayNumber("2026-07-19")
	if err != nil {
		t.Fatal(err)
	}
	if got := RankDate(day); got != "2026-07-19" {
		t.Fatalf("rank date=%s", got)
	}
}
