package contracts

import "testing"

func TestLegacyAPIRoutesAreMigrationBaseline(t *testing.T) {
	routes := LegacyAPIRoutes()
	if len(routes) < 60 {
		t.Fatalf("legacy api baseline is incomplete: got %d routes", len(routes))
	}

	required := []LegacyRoute{
		{Method: "POST", Path: "/api/movie/:movie/downloadlater"},
		{Method: "POST", Path: "/api/triggers/media/commit"},
		{Method: "POST", Path: "/api/agg/w-media/backfill"},
		{Method: "GET", Path: "/api/crawler/jobs/stream"},
		{Method: "POST", Path: "/api/crawler/jobs/:jobID/stop"},
	}
	for _, route := range required {
		if !hasRoute(routes, route.Method, route.Path) {
			t.Fatalf("missing legacy api baseline route %s %s", route.Method, route.Path)
		}
	}
}

func TestLegacyPageRoutesAreScreenshotBaseline(t *testing.T) {
	routes := LegacyPageRoutes()
	if len(routes) < 45 {
		t.Fatalf("legacy page baseline is incomplete: got %d routes", len(routes))
	}

	required := []LegacyRoute{
		{Method: "GET", Path: "/cardstoday"},
		{Method: "GET", Path: "/movie/:movie"},
		{Method: "GET", Path: "/triggers/media"},
		{Method: "GET", Path: "/sc-pick-smart-media"},
		{Method: "GET", Path: "/movie-agg-all/release-bucket-list"},
	}
	for _, route := range required {
		if !hasRoute(routes, route.Method, route.Path) {
			t.Fatalf("missing legacy page baseline route %s %s", route.Method, route.Path)
		}
	}
}

func TestLegacyBaselineDoesNotUseGenericPageAction(t *testing.T) {
	for _, route := range LegacyAPIRoutes() {
		if route.Path == "/api/gc/v2/page-actions" || route.Path == "/api/page-actions" {
			t.Fatalf("generic page action must not be treated as legacy baseline: %#v", route)
		}
	}
}

func hasRoute(routes []LegacyRoute, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
