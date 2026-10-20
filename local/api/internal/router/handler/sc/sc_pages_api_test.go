package sc

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildScEventListFilterIncludesOptionIDs(t *testing.T) {
	filter, err := buildScEventListFilter(scEventListQuery{
		KindOptionID: 3, FGOptionID: 4, VesselOptionID: 5,
	}, 100, 200)
	if err != nil {
		t.Fatalf("buildScEventListFilter: %v", err)
	}
	if filter.StartTime != 100 || filter.EndTime != 200 || filter.KindOptionID != 3 || filter.FGOptionID != 4 || filter.VesselOptionID != 5 {
		t.Fatalf("filter = %#v", filter)
	}
}

func TestBuildScEventListFilterRejectsNegativeOptionID(t *testing.T) {
	if _, err := buildScEventListFilter(scEventListQuery{FGOptionID: -1}, 0, 0); err == nil {
		t.Fatal("negative option id must be rejected")
	}
}

func TestScEventSortAndViewLinksPreserveOptionFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/scevents?kind_option_id=3&fg_option_id=4&vessel_option_id=5", nil)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	sortLinks := buildScEventSortQuery(ctx, "sc_time", "desc")
	assertScEventOptionFilterQuery(t, sortLinks.ByKind.Href)
	viewLinks := buildScEventViewQuery(ctx)
	assertScEventOptionFilterQuery(t, viewLinks.CardHref)
}

func assertScEventOptionFilterQuery(t *testing.T, href string) {
	t.Helper()
	parsed, err := url.Parse(href)
	if err != nil {
		t.Fatalf("parse href %q: %v", href, err)
	}
	query := parsed.Query()
	if query.Get("kind_option_id") != "3" || query.Get("fg_option_id") != "4" || query.Get("vessel_option_id") != "5" {
		t.Fatalf("href filters = %q", href)
	}
}
