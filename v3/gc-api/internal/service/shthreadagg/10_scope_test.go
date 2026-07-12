package shthreadagg

import (
	"testing"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func TestBuildScopesIncludesAllLevels(t *testing.T) {
	postTime := time.Date(2026, time.July, 12, 15, 30, 0, 0, time.Local).Unix()
	scopes := buildScopes(postTime)
	if len(scopes) != 5 {
		t.Fatalf("scope count = %d, want 5", len(scopes))
	}
	wantKeys := map[string]bool{
		"day:2026-07-12":  true,
		"week:2026-W28":   true,
		"month:2026-07":   true,
		"quarter:2026-Q3": true,
		"year:2026":       true,
	}
	for _, scope := range scopes {
		if !wantKeys[scope.Key] {
			t.Fatalf("unexpected scope key %q", scope.Key)
		}
		if scope.StartUnix > postTime || scope.EndUnix < postTime {
			t.Fatalf("scope %q does not contain post time", scope.Key)
		}
	}
}

func TestBuildScopesUsesISOWeekYear(t *testing.T) {
	postTime := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.Local).Unix()
	scopes := buildScopes(postTime)
	for _, scope := range scopes {
		if scope.Level != levelWeek {
			continue
		}
		if scope.Key != "week:2020-W53" || scope.Year != 2020 || scope.Week != 53 {
			t.Fatalf("week scope = %+v, want ISO week 2020-W53", scope)
		}
		start := time.Unix(scope.StartUnix, 0).In(time.Local)
		end := time.Unix(scope.EndUnix, 0).In(time.Local)
		if start.Weekday() != time.Monday || end.Weekday() != time.Sunday {
			t.Fatalf("week range = %s to %s, want Monday to Sunday", start, end)
		}
		return
	}
	t.Fatal("week scope missing")
}

func TestBuildBucketWorksDeduplicatesScopesAndKeepsForumScopes(t *testing.T) {
	day := time.Date(2026, time.July, 12, 10, 0, 0, 0, time.Local)
	works := buildBucketWorks([]moviex.ShThreadAggTarget{
		{PostTime: day.Unix(), SourceForumID: "36"},
		{PostTime: day.Add(time.Hour).Unix(), SourceForumID: "36"},
		{PostTime: day.Add(2 * time.Hour).Unix(), SourceForumID: "103"},
	})
	if len(works) != 15 {
		t.Fatalf("work count = %d, want 15", len(works))
	}
	counts := map[string]int{}
	for _, work := range works {
		counts[work.AggScope]++
	}
	for _, aggScope := range []string{AggScopeAll, "forum:36", "forum:103"} {
		if counts[aggScope] != 5 {
			t.Fatalf("scope %q work count = %d, want 5", aggScope, counts[aggScope])
		}
	}
}
