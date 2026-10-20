package sc

import (
	"testing"

	"rudy_gc/local/api/internal/service/media"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

func TestBuildScMediaMoveRowKeepsFC2Identity(t *testing.T) {
	row := buildScMediaMoveRow(&media.ScMediaMoveItem{
		Status:    "pass",
		MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
		FC2Key:    " FC2PPV-1234567-1 ",
		MovieName: "FC2影片",
	})
	if row.MovieType != moviedomain.SehuatangThreadMovieTypeFC2PPV {
		t.Fatalf("影片类型 = %q", row.MovieType)
	}
	if row.FC2Key != "FC2PPV-1234567-1" {
		t.Fatalf("fc2_key = %q", row.FC2Key)
	}
	if row.MovieJavId != "" {
		t.Fatalf("FC2响应包含 movie_jav_id: %q", row.MovieJavId)
	}
}
