package sc

import (
	"testing"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/types"
)

func TestNormalizeSmartPickReqUsesUnboundedCandidateQuery(t *testing.T) {
	req := types.ListMovieFullRequest{Page: 9, PageSize: 100000}
	normalizeSmartPickReq(&req, SmartPickSourceWMedia)
	if req.Page != 1 {
		t.Fatalf("Page = %d", req.Page)
	}
	if req.PageSize != 0 {
		t.Fatalf("PageSize = %d, want unbounded 0", req.PageSize)
	}
	if req.MediaOwned != consts.OwnedAllNotRemoved {
		t.Fatalf("MediaOwned = %d", req.MediaOwned)
	}
}

func TestAllocateSmartGroupQuotasRedistributesUnavailableCapacity(t *testing.T) {
	limited := &smartPickGroup{
		weight:     10,
		candidates: []*types.MovieType{{JavId: "A"}},
		alive:      []int{0},
	}
	available := &smartPickGroup{
		weight: 1,
		candidates: []*types.MovieType{
			{JavId: "B"}, {JavId: "C"}, {JavId: "D"}, {JavId: "E"},
		},
		alive: []int{0, 1, 2, 3},
	}
	quotas, err := allocateSmartGroupQuotas(
		[]*smartPickGroup{limited, available},
		smartBucketFree,
		5,
		map[string]struct{}{},
		map[string]struct{}{},
		map[string]smartCastStat{},
		SmartPickOptions{},
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if quotas[limited] != 1 || quotas[available] != 4 {
		t.Fatalf("quotas = limited:%d available:%d", quotas[limited], quotas[available])
	}
}
