package rank

import "testing"

func TestBuildPeriodHrefKeepsCategoryAndKey(t *testing.T) {
	href := buildPeriodHref("month", 1, "2026-05")
	if href != "/moviecardperiodrank?category=1&key=2026-05&type=month" {
		t.Fatalf("unexpected href: %s", href)
	}
}
