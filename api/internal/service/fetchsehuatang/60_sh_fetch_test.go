package fetchsehuatang

import (
	"context"
	"testing"

	"rudy_gc/internal/taskctx"
)

func TestBuildSHProgressStatsIncludesListRunMetrics(t *testing.T) {
	result := &SHFetchResult{
		RunID:              12,
		ExistingPolicy:     SHExistingPolicySkip,
		TargetForumIDs:     "103",
		StartPage:          3,
		EndPage:            4,
		TotalPageCount:     2,
		CompletedPageCount: 1,
		CurrentPage:        4,
		CurrentPageText:    "forum=103 page=4",
		CurrentForumID:     "103",
		SuccessPageCount:   1,
		FailedPageCount:    0,
		SeenCount:          40,
		MatchedCount:       38,
		InsertedCount:      8,
		ExistingCount:      29,
		SkippedTopCount:    3,
		PersistFailedCount: 1,
	}

	stats := BuildSHProgressStats(result)
	wantNumbers := map[string]int64{
		"run_id":               12,
		"start_page":           3,
		"end_page":             4,
		"total_page_count":     2,
		"completed_page_count": 1,
		"current_page":         4,
		"success_page_count":   1,
		"failed_page_count":    0,
		"seen_count":           40,
		"matched_count":        38,
		"inserted_count":       8,
		"existing_count":       29,
		"skipped_top_count":    3,
		"persist_failed_count": 1,
	}
	for key, want := range wantNumbers {
		if got := stats[key]; got != want {
			t.Fatalf("stats[%q] = %d, want %d", key, got, want)
		}
	}

	statsText := BuildSHProgressStatsText(result)
	wantText := map[string]string{
		"current_page_text": "forum=103 page=4",
		"current_forum_id":  "103",
		"target_forum_ids":  "103",
		"existing_policy":   SHExistingPolicySkip,
	}
	for key, want := range wantText {
		if got := statsText[key]; got != want {
			t.Fatalf("statsText[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestNormalizeSHExistingPolicy(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "default", raw: "", want: SHExistingPolicyUpsert},
		{name: "upsert", raw: " UPSERT ", want: SHExistingPolicyUpsert},
		{name: "skip", raw: " Skip ", want: SHExistingPolicySkip},
		{name: "invalid", raw: "ignore", want: "ignore"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizeSHExistingPolicy(test.raw); got != test.want {
				t.Fatalf("NormalizeSHExistingPolicy(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}

func TestValidateSHFetchRequestRejectsInvalidExistingPolicy(t *testing.T) {
	req := normalizeSHFetchRequest(SHFetchRequest{
		FetchMode:      SHFetchModeForum,
		ExistingPolicy: "ignore",
		StartPage:      1,
		EndPage:        1,
	})
	if err := validateSHFetchRequest(req); err == nil {
		t.Fatal("validateSHFetchRequest() error = nil, want invalid existing_policy error")
	}
}

func TestSHListTopicNeedsDetail(t *testing.T) {
	tests := []struct {
		name     string
		existed  bool
		status   string
		expected bool
	}{
		{name: "new", existed: false, status: "", expected: true},
		{name: "pending", existed: true, status: SHDetailStatusPending, expected: true},
		{name: "failed", existed: true, status: SHDetailStatusFailed, expected: true},
		{name: "fetched", existed: true, status: SHDetailStatusFetched, expected: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shListTopicNeedsDetail(test.existed, test.status); got != test.expected {
				t.Fatalf("shListTopicNeedsDetail(%v, %q) = %v, want %v", test.existed, test.status, got, test.expected)
			}
		})
	}
}

func TestSHListTopicShouldSkip(t *testing.T) {
	tests := []struct {
		name     string
		policy   string
		existed  bool
		expected bool
	}{
		{name: "skip existing", policy: SHExistingPolicySkip, existed: true, expected: true},
		{name: "skip new", policy: SHExistingPolicySkip, existed: false, expected: false},
		{name: "upsert existing", policy: SHExistingPolicyUpsert, existed: true, expected: false},
		{name: "default existing", policy: "", existed: true, expected: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shListTopicShouldSkip(test.policy, test.existed); got != test.expected {
				t.Fatalf("shListTopicShouldSkip(%q, %v) = %v, want %v", test.policy, test.existed, got, test.expected)
			}
		})
	}
}

func TestReportSHDetailProgressKeepsListStats(t *testing.T) {
	var reported taskctx.Progress
	ctx := taskctx.WithProgressReporter(context.Background(), func(progress taskctx.Progress) {
		reported = progress
	})
	reportSHDetailProgress(ctx, &SHFetchResult{
		MatchedCount: 3,
		FetchedCount: 2,
		FailedCount:  1,
	}, SHDetailFetchRequest{
		BaseStats:     map[string]int64{"inserted_count": 8},
		BaseStatsText: map[string]string{"current_page_text": "forum=36 page=2"},
	}, taskctx.Progress{Stage: "detail"})

	if got := reported.Stats["inserted_count"]; got != 8 {
		t.Fatalf("inserted_count = %d, want 8", got)
	}
	if got := reported.Stats["detail_matched_count"]; got != 3 {
		t.Fatalf("detail_matched_count = %d, want 3", got)
	}
	if got := reported.Stats["detail_fetched_count"]; got != 2 {
		t.Fatalf("detail_fetched_count = %d, want 2", got)
	}
	if got := reported.Stats["detail_failed_count"]; got != 1 {
		t.Fatalf("detail_failed_count = %d, want 1", got)
	}
	if got := reported.StatsText["current_page_text"]; got != "forum=36 page=2" {
		t.Fatalf("current_page_text = %q", got)
	}
}
