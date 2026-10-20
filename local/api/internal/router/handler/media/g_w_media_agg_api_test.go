package media

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBirthBucketChangeSortLinks(t *testing.T) {
	columns := []struct {
		sort  string
		label string
		link  func(wMediaBirthBucketHeaderLinks) wMediaBirthBucketSortLink
	}{
		{"added", "新增数目", func(h wMediaBirthBucketHeaderLinks) wMediaBirthBucketSortLink { return h.Added }},
		{"added_size", "新增大小", func(h wMediaBirthBucketHeaderLinks) wMediaBirthBucketSortLink { return h.AddedSize }},
		{"removed", "删除数目", func(h wMediaBirthBucketHeaderLinks) wMediaBirthBucketSortLink { return h.Removed }},
		{"removed_size", "删除大小", func(h wMediaBirthBucketHeaderLinks) wMediaBirthBucketSortLink { return h.RemovedSize }},
		{"net_size", "净增加大小", func(h wMediaBirthBucketHeaderLinks) wMediaBirthBucketSortLink { return h.NetSize }},
	}
	for _, column := range columns {
		t.Run(column.sort, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/bucketsbirth?level=month&year=2026&month=8&scope_key=2026&p=3&ps=24", nil)
			if got := normalizeBucketListSort(column.sort); got != column.sort {
				t.Fatalf("normalized sort = %q", got)
			}
			for _, currentDir := range []string{"asc", "desc"} {
				link := column.link(buildBucketListHeaderLinks(c, column.sort, currentDir))
				if link.Label != column.label || !link.Active || link.Desc != (currentDir == "desc") {
					t.Fatalf("header = %+v", link)
				}
				parsed, err := url.Parse(link.Href)
				if err != nil {
					t.Fatal(err)
				}
				wantDir := "desc"
				if currentDir == "desc" {
					wantDir = "asc"
				}
				query := parsed.Query()
				for key, want := range map[string]string{"sort": column.sort, "dir": wantDir, "p": "1", "ps": "24", "year": "2026", "month": "8", "scope_key": "2026", "level": "month"} {
					if query.Get(key) != want {
						t.Fatalf("%s = %q, want %q", key, query.Get(key), want)
					}
				}
				if parsed.Path != "/bucketsbirth" {
					t.Fatalf("sort path = %q", parsed.Path)
				}
			}
		})
	}
}
