package list

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCrawlRecordFilterTypeAll(t *testing.T) {
	if got := crawlRecordFilterType(crawlRecordTypeAll); got != "" {
		t.Fatalf("crawlRecordFilterType(%q) = %q, want empty filter", crawlRecordTypeAll, got)
	}

	const recordType = "DailyBest"
	if got := crawlRecordFilterType(recordType); got != recordType {
		t.Fatalf("crawlRecordFilterType(%q) = %q, want %q", recordType, got, recordType)
	}
}

func TestBuildCrawlRecordTypeLinksAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/crawlrecordlist?type=all&sort=start_time", nil)

	links := buildCrawlRecordTypeLinks(c, crawlRecordTypeAll)
	if len(links) == 0 {
		t.Fatal("buildCrawlRecordTypeLinks returned no links")
	}

	all := links[0]
	if all.Label != "All" {
		t.Fatalf("first link label = %q, want All", all.Label)
	}
	if !all.Active {
		t.Fatal("All link is not active")
	}
	if all.Href != "/crawlrecordlist?p=1&sort=start_time&type=all" {
		t.Fatalf("All link href = %q", all.Href)
	}
}
