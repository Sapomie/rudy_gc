package dailybest

import (
	"testing"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func TestParsePagesBuildsStructuredRanking(t *testing.T) {
	content := `<div class="videothumblist"><div class="videos">
<div class="video"><div class="title">ABC-001 标题</div><div class="id">ABC-001</div><div class="toolbar"><a id="jav-1"></a></div><img src="cover-1.jpg"></div>
<div class="video"><div class="title">ABC-002（ブルーレイディスク）</div><div class="id">ABC-002</div><div class="toolbar"><a id="jav-2"></a></div></div>
<div class="video"><div class="title">ABC-003 标题</div><div class="id">ABC-003</div><div class="toolbar"><a id="jav-3"></a></div><img src="cover-3.jpg"></div>
</div></div>`
	items, err := ParsePages([]dailybesttypes.SourcePage{{
		PageNumber: 1,
		SourceURL:  "https://example.com/cn/rank?page=1",
		Content:    content,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items=%d, want 2", len(items))
	}
	if items[0].Rank.RankPos != 1 || items[0].Rank.MovieJavID != "jav-1" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[0].Movie.MovieName != "ABC-001" || items[0].Movie.Title != "ABC-001 标题" {
		t.Fatalf("unexpected first movie: %+v", items[0].Movie)
	}
	if items[0].Movie.CoverURL != "https://example.com/cn/cover-1.jpg" {
		t.Fatalf("unexpected first cover URL: %s", items[0].Movie.CoverURL)
	}
	if items[1].Rank.RankPos != 2 || items[1].Rank.MovieJavID != "jav-3" {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
}

func TestParsePagesRejectsDuplicateJavID(t *testing.T) {
	content := `<div class="videothumblist"><div class="videos">
<div class="video"><div class="title">A</div><div class="id">A</div><div class="toolbar"><a id="same"></a></div></div>
<div class="video"><div class="title">B</div><div class="id">B</div><div class="toolbar"><a id="same"></a></div></div>
</div></div>`
	if _, err := ParsePages([]dailybesttypes.SourcePage{{PageNumber: 1, Content: content}}); err == nil {
		t.Fatal("expected duplicate JavID error")
	}
}
