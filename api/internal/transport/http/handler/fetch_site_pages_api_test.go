package handler

import (
	"strings"
	"testing"
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
