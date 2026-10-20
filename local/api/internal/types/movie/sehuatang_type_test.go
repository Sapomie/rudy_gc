package movie

import (
	"reflect"
	"testing"
)

func TestResolveSehuatangThreadMovieType(t *testing.T) {
	tests := []struct {
		name       string
		movieName  string
		movieJavID string
		want       string
	}{
		{name: "jav", movieName: "ABP-123", movieJavID: "123", want: SehuatangThreadMovieTypeJAV},
		{name: "fc2", movieName: "FC2PPV-4556413", want: SehuatangThreadMovieTypeFC2PPV},
		{name: "fc2 precedence", movieName: "fc2ppv- 4556413", movieJavID: "123", want: SehuatangThreadMovieTypeFC2PPV},
		{name: "unclassified", movieName: "unknown", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveSehuatangThreadMovieType(tt.movieName, tt.movieJavID); got != tt.want {
				t.Fatalf("ResolveSehuatangThreadMovieType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSehuatangThreadMovieTypes(t *testing.T) {
	want := []string{SehuatangThreadMovieTypeJAV, SehuatangThreadMovieTypeFC2PPV}
	got := NormalizeSehuatangThreadMovieTypes(" JAV ", "invalid", "fc2ppv", "jav")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeSehuatangThreadMovieTypes() = %#v, want %#v", got, want)
	}
}
