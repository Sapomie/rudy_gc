package types

import "testing"

func TestBuildRankKeyIncludesCategory(t *testing.T) {
	got := BuildRankKey("2026-07-20", 2, 7)
	if got != "2026-07-20_2_007" {
		t.Fatalf("BuildRankKey() = %q", got)
	}
}
