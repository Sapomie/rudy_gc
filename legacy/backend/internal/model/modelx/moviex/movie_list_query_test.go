package moviex

import (
	"testing"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func TestShouldPickFromGScStat(t *testing.T) {
	t.Run("media birth filter should not force gsc stat", func(t *testing.T) {
		req := &types.ListMovieFullRequest{
			MediaOwned:          consts.OwnedAllNotRemoved,
			MediaBirthTimeStart: "2026-03-01",
			MediaBirthTimeEnd:   "2026-03-31",
			OrderBy:             consts.OrderByReleasingDate,
		}
		if shouldPickFromGScStat(req) {
			t.Fatalf("expected false, got true")
		}
	})

	t.Run("sc filters should use gsc stat", func(t *testing.T) {
		req := &types.ListMovieFullRequest{
			ScTimesMin: 1,
			OrderBy:    consts.OrderByReleasingDate,
		}
		if !shouldPickFromGScStat(req) {
			t.Fatalf("expected true, got false")
		}
	})

	t.Run("sc ordering should use gsc stat", func(t *testing.T) {
		req := &types.ListMovieFullRequest{
			OrderBy: consts.OrderByScTimes,
		}
		if !shouldPickFromGScStat(req) {
			t.Fatalf("expected true, got false")
		}
	})
}
