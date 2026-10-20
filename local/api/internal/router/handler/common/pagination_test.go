package common

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
)

func TestBuildSortQueryTogglesActiveDirection(t *testing.T) {
	query := BuildSortQuery(sortTestContext("/listcard?od=rd&cn=test"), consts.OrderByReleasingDate)

	if !query.ByReleasing.Active || !query.ByReleasing.Desc {
		t.Fatalf("expected releasing sort to be active descending: %+v", query.ByReleasing)
	}
	assertSortHref(t, query.ByReleasing.Href, consts.OrderByReleasingDate, "asc", "test")
	assertSortHref(t, query.ByViewer.Href, consts.OrderByViewerWatched, "desc", "test")
}

func TestBuildSortQueryUsesFieldDefaultAndExplicitDirection(t *testing.T) {
	highestRank := BuildSortQuery(sortTestContext("/listcard?od=hrk"), consts.OrderByHighestRank)
	if !highestRank.ByHighestRank.Active || highestRank.ByHighestRank.Desc {
		t.Fatalf("expected highest rank to default to ascending: %+v", highestRank.ByHighestRank)
	}
	assertSortHref(t, highestRank.ByHighestRank.Href, consts.OrderByHighestRank, "desc", "")

	releasingAsc := BuildSortQuery(
		sortTestContext("/listcard?od=rd&order=asc"),
		consts.OrderByReleasingDate,
	)
	if !releasingAsc.ByReleasing.Active || releasingAsc.ByReleasing.Desc {
		t.Fatalf("expected explicit ascending direction: %+v", releasingAsc.ByReleasing)
	}
	assertSortHref(t, releasingAsc.ByReleasing.Href, consts.OrderByReleasingDate, "desc", "")
}

func TestBuildSortQueryProvidesDirectionForEveryMovieCardSort(t *testing.T) {
	query := BuildSortQuery(sortTestContext("/listcard?od=rd&order=desc&cn=test"), consts.OrderByReleasingDate)

	links := map[string]SortLink{
		consts.OrderByDetailUpdateTime:   query.ByDetailUpdate,
		consts.OrderByReleasingDate:      query.ByReleasing,
		consts.OrderByViewerWatched:      query.ByViewer,
		consts.OrderByDaysInRank:         query.ByDaysInRank,
		consts.OrderByHighestRank:        query.ByHighestRank,
		consts.OrderByRankDate:           query.ByRankDate,
		consts.OrderByScTimes:            query.ByScTimes,
		consts.OrderByLastSelectedScTime: query.ByLastSelectedScTime,
		consts.OrderByBirthTime:          query.ByBirthTime,
		consts.OrderByMediaBirthTime:     query.ByMediaBirthTime,
		consts.OrderByCastAgeDesc:        query.ByCastAgeDesc,
		consts.OrderByCastAgeAsc:         query.ByCastAgeAsc,
		consts.OrderByComeTimes:          query.ByComeTimes,
	}

	for orderBy, link := range links {
		wantOrder := movieCardDefaultOrder(orderBy)
		if orderBy == consts.OrderByReleasingDate {
			wantOrder = "asc"
		}
		assertSortHref(t, link.Href, orderBy, wantOrder, "test")
	}

	if !query.ByReleasing.Active || !query.ByReleasing.Desc {
		t.Fatal("当前上映日期排序应标记为倒序")
	}
}

func sortTestContext(target string) *gin.Context {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("GET", target, nil)
	return context
}

func assertSortHref(t *testing.T, href, orderBy, order, castName string) {
	t.Helper()
	parsed, err := url.Parse(href)
	if err != nil {
		t.Fatalf("parse sort href %q: %v", href, err)
	}
	query := parsed.Query()
	if query.Get("od") != orderBy || query.Get("order") != order || query.Get("p") != "1" {
		t.Fatalf("unexpected sort href query: %s", query.Encode())
	}
	if query.Get("cn") != castName {
		t.Fatalf("expected preserved cast filter %q, got %q", castName, query.Get("cn"))
	}
}
