package shthreadagg

import (
	"net/url"
	"testing"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func TestSHBucketDisplay(t *testing.T) {
	tests := []struct {
		row  moviex.ShThreadBucketStat
		want string
	}{
		{row: moviex.ShThreadBucketStat{Level: levelYear, Year: 2026}, want: "2026"},
		{row: moviex.ShThreadBucketStat{Level: levelQuarter, Year: 2026, Quarter: 3}, want: "2026-Q3"},
		{row: moviex.ShThreadBucketStat{Level: levelMonth, Year: 2026, Month: 7}, want: "2026-07"},
		{row: moviex.ShThreadBucketStat{Level: levelWeek, Year: 2026, Week: 28}, want: "2026-W28"},
		{row: moviex.ShThreadBucketStat{Level: levelDay, Year: 2026, Month: 7, Day: 12}, want: "2026-07-12"},
	}
	for _, test := range tests {
		if got := shBucketDisplay(&test.row); got != test.want {
			t.Fatalf("display = %q, want %q", got, test.want)
		}
	}
}

func TestSHBucketViewHrefIncludesRangeAndForum(t *testing.T) {
	start := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, time.July, 12, 23, 59, 59, 0, time.Local)
	href := shBucketViewHref(&moviex.ShThreadBucketStat{
		AggScope: "forum:36", BucketStart: start.Unix(), BucketEnd: end.Unix(),
	})
	parsed, err := url.Parse(href)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/shpool" {
		t.Fatalf("path = %q, want /shpool", parsed.Path)
	}
	query := parsed.Query()
	if query.Get("post_time_from") != "2026-07-06" || query.Get("post_time_to") != "2026-07-12" {
		t.Fatalf("date range = %q to %q", query.Get("post_time_from"), query.Get("post_time_to"))
	}
	if query.Get("source_forum_ids") != "36" {
		t.Fatalf("forum = %q, want 36", query.Get("source_forum_ids"))
	}
}
