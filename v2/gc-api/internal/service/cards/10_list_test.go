package cards

import (
	"testing"

	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"
)

func TestNormalizeRequestKeepsExplicitPositivePageSize(t *testing.T) {
	req := &types.CardsListRequest{View: "cards", PageSize: 30000, Page: 2}
	normalizeRequest(req)

	if req.PageSize != 30000 {
		t.Fatalf("page size should stay explicit, got %d", req.PageSize)
	}
	if req.Page != 2 {
		t.Fatalf("page should stay explicit, got %d", req.Page)
	}
}

func TestNormalizeRequestAppliesLegacyViewDefaults(t *testing.T) {
	req := &types.CardsListRequest{View: "cardsneeddownload"}
	normalizeRequest(req)

	if req.Page != 1 || req.PageSize != 18 {
		t.Fatalf("unexpected pagination defaults: page=%d pageSize=%d", req.Page, req.PageSize)
	}
	if req.DaysInRankMin != nil {
		t.Fatalf("need-download view should not set days-in-rank default, got %d", *req.DaysInRankMin)
	}
	if req.NeedDownload != consts.MovieNeedDownloadOK {
		t.Fatalf("need-download view should default need_download ok, got %d", req.NeedDownload)
	}
}

func TestNormalizeRequestKeepsExplicitZeroDaysInRank(t *testing.T) {
	zero := int64(0)
	req := &types.CardsListRequest{View: "cardshasrank", DaysInRankMin: &zero}
	normalizeRequest(req)

	if req.DaysInRankMin == nil || *req.DaysInRankMin != 0 {
		t.Fatalf("explicit drkmin=0 should be preserved, got %#v", req.DaysInRankMin)
	}
}
