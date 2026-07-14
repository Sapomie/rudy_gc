package sc

import (
	"testing"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
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
