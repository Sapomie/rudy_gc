package moviex

import (
	"testing"

	"rudy_gc/internal/types"
)

func TestMissingGScStatMatches(t *testing.T) {
	zero := int64(0)
	negative := int64(-1)
	tests := []struct {
		name string
		req  *types.ListMovieFullRequest
		want bool
	}{
		{name: "empty request", req: &types.ListMovieFullRequest{}, want: true},
		{name: "zero maximum includes missing stat", req: &types.ListMovieFullRequest{ScTimesMax: &zero}, want: true},
		{name: "positive minimum excludes missing stat", req: &types.ListMovieFullRequest{ScTimesMin: 1}, want: false},
		{name: "date minimum excludes missing stat", req: &types.ListMovieFullRequest{LastScTimeMin: "2026-01-01"}, want: false},
		{name: "negative maximum excludes missing stat", req: &types.ListMovieFullRequest{ComeTimesMax: &negative}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := missingGScStatMatches(tt.req); got != tt.want {
				t.Fatalf("missingGScStatMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
