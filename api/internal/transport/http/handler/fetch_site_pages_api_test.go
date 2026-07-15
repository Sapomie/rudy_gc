package handler

import (
	"strings"
	"testing"

	"rudy_gc/internal/repository/moviex"
)

func TestNormalizeSHSehuatangThreadSortField(t *testing.T) {
	for _, field := range []string{"thread_tid", "movie_name", "post_time", "detail_status", "last_seen_time", "created_on", "updated_on"} {
		if got := normalizeSHSehuatangThreadSortField(field); got != field {
			t.Fatalf("normalizeSHSehuatangThreadSortField(%q) = %q", field, got)
		}
	}
	if got := normalizeSHSehuatangThreadSortField("invalid"); got != "post_time" {
		t.Fatalf("invalid sort field = %q", got)
	}
}

func TestBuildSHSehuatangThreadOrderBy(t *testing.T) {
	for _, field := range []string{"movie_name", "detail_status"} {
		orderBy := buildSHSehuatangThreadOrderBy(field, "asc")
		if !strings.HasPrefix(orderBy, "`"+field+"` ASC") {
			t.Fatalf("order by %s = %q", field, orderBy)
		}
	}
}

func TestApplySHThreadFC2Keys(t *testing.T) {
	rows := []*moviex.ShSehuatangThread{
		{ThreadTID: "3619676", FC2MovieID: 96, FC2Key: ""},
		{ThreadTID: "jav", MovieJavId: "ABP-123"},
	}
	movieRows := []*moviex.FC2Movie{
		{ID: 96, FC2Key: " FC2PPV-4819241-1 "},
	}

	applySHThreadFC2Keys(rows, movieRows)

	if got := rows[0].FC2Key; got != "FC2PPV-4819241-1" {
		t.Fatalf("FC2Key = %q", got)
	}
	if got := rows[1].FC2Key; got != "" {
		t.Fatalf("JAV row FC2Key = %q", got)
	}
}
