package sc

import (
	"encoding/json"
	"strings"
	"testing"

	service "rudy_gc/local/api/internal/service/sc"
	"rudy_gc/local/api/internal/types"
)

func TestBuildPickV2MoviesUsesMovieCardContractAndKeepsOrder(t *testing.T) {
	movies := []*types.MovieType{
		{
			Name:                "SECOND-2",
			FilmBirthDateWMedia: "2026-02-02",
			VideoUrlWMedia:      "/media/second.mp4",
			OwnedWMedia:         5,
			Cast:                []*types.CastInfo{{PersonId: 7, Name: "raw", NameShow: "显示名"}},
		},
		{Name: "FIRST-1", FilmBirthDateWMedia: "2025-01-01"},
	}

	got := buildPickV2Movies(movies, service.SmartPickSourceWMedia)
	if len(got) != 2 || got[0].Name != "SECOND-2" || got[0].PickedOrder != 0 || got[1].PickedOrder != 1 {
		t.Fatalf("unexpected ordered result: %#v", got)
	}
	if got[0].Cast[0].Name != "显示名" {
		t.Fatalf("cast display name = %q", got[0].Cast[0].Name)
	}
	raw, err := json.Marshal(got[0])
	if err != nil {
		t.Fatal(err)
	}
	jsonText := string(raw)
	for _, key := range []string{"film_birth_date_w_media", "video_url_w_media", "owned_w_media", "picked_order"} {
		if !strings.Contains(jsonText, `"`+key+`"`) {
			t.Fatalf("missing JSON key %s in %s", key, jsonText)
		}
	}
}
