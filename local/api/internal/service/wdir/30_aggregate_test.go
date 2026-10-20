package wdir

import (
	"testing"

	"rudy_gc/local/api/internal/types"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

func TestBuildDirStatsFromAllClassifiesMediaType(t *testing.T) {
	tests := []struct {
		name      string
		films     []*types.Film
		mediaType string
		count     int64
	}{
		{name: "empty", mediaType: directoryMediaTypeEmpty},
		{
			name: "jav",
			films: []*types.Film{
				{MovieType: moviedomain.SehuatangThreadMovieTypeJAV},
			},
			mediaType: moviedomain.SehuatangThreadMovieTypeJAV,
			count:     1,
		},
		{
			name: "fc2",
			films: []*types.Film{
				{MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV},
			},
			mediaType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
			count:     1,
		},
		{
			name: "mixed",
			films: []*types.Film{
				{MovieType: moviedomain.SehuatangThreadMovieTypeJAV},
				nil,
				{MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV},
			},
			mediaType: directoryMediaTypeMixed,
			count:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := buildDirStatsFromAll(tt.films, true)
			if stats.MediaType != tt.mediaType {
				t.Fatalf("MediaType = %q, want %q", stats.MediaType, tt.mediaType)
			}
			if stats.FilmCount != tt.count {
				t.Fatalf("FilmCount = %d, want %d", stats.FilmCount, tt.count)
			}
		})
	}
}

func TestAggregateDirStatsForIDsClassifiesSelectedSubtree(t *testing.T) {
	stats := aggregateDirStatsForIDs([]*types.Film{
		{DirectoryId: 11, MovieType: moviedomain.SehuatangThreadMovieTypeJAV},
		{DirectoryId: 12, MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV},
		{DirectoryId: 13, MovieType: moviedomain.SehuatangThreadMovieTypeJAV},
	}, []int64{11, 12}, true)

	if stats == nil {
		t.Fatal("stats is nil")
	}
	if stats.MediaType != directoryMediaTypeMixed {
		t.Fatalf("MediaType = %q, want %q", stats.MediaType, directoryMediaTypeMixed)
	}
	if stats.FilmCount != 2 {
		t.Fatalf("FilmCount = %d, want 2", stats.FilmCount)
	}
}
