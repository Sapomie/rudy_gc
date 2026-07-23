package dailybestremotex

import (
	"testing"

	"rudy_gc/remote/api/internal/model/modelg"
)

func TestRankMappingsIncludeMovieBasic(t *testing.T) {
	movie := modelg.DailyBestMovie{
		MovieJavID: "jav-1",
		MovieName:  "ABC-001",
		MovieCode:  "ABC-001",
		Title:      "ABC-001 完整标题",
		CoverURL:   "https://example.com/cover.jpg",
	}
	movies := map[string]modelg.DailyBestMovie{movie.MovieJavID: movie}

	ranks := mapRanks([]modelg.Rank{{Id: 1, MovieJavID: movie.MovieJavID}}, movies)
	if len(ranks) != 1 || ranks[0].Movie.Title != movie.Title || ranks[0].Movie.CoverURL != movie.CoverURL {
		t.Fatalf("rank movie mismatch: %+v", ranks)
	}
}
