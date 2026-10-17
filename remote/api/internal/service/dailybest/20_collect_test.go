package dailybest

import (
	"context"
	"fmt"
	"testing"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type repeatedLastPageSource struct {
	requestedPages []int64
}

func (s *repeatedLastPageSource) FetchPage(_ context.Context, _ int64, pageNumber int64) (dailybesttypes.SourcePage, error) {
	s.requestedPages = append(s.requestedPages, pageNumber)
	switch pageNumber {
	case 1:
		return collectTestPage(1, "page-1", "JAV-001"), nil
	case 2:
		return collectTestPage(2, "page-2", "JAV-002"), nil
	case 3:
		return collectTestPage(3, "page-2", "JAV-002"), nil
	default:
		return dailybesttypes.SourcePage{}, fmt.Errorf("不应继续请求第 %d 页", pageNumber)
	}
}

type collectPublishRepository struct {
	Repository
	publishedDraft dailybesttypes.RankDraft
	publishedStat  contracts.DailyBestRankDayStat
}

func (r *collectPublishRepository) ListAllRanks(context.Context) ([]contracts.DailyBestRank, error) {
	return nil, nil
}

func (r *collectPublishRepository) PublishRanks(
	_ context.Context,
	_ int64,
	draft dailybesttypes.RankDraft,
	stat contracts.DailyBestRankDayStat,
	_ int64,
) error {
	r.publishedDraft = draft
	r.publishedStat = stat
	return nil
}

func TestCollectAndPublishTreatsRepeatedPageAsPaginationEnd(t *testing.T) {
	repository := &collectPublishRepository{}
	source := &repeatedLastPageSource{}
	service := New(context.Background(), repository, source, nil)

	if err := service.collectAndPublish(context.Background(), 48, "2026-07-28", contracts.DailyBestCategoryMonth); err != nil {
		t.Fatal(err)
	}

	if len(source.requestedPages) != 3 || source.requestedPages[0] != 1 || source.requestedPages[1] != 2 || source.requestedPages[2] != 3 {
		t.Fatalf("requested pages=%v, want [1 2 3]", source.requestedPages)
	}
	if repository.publishedDraft.PageCount != 2 {
		t.Fatalf("published page count=%d, want 2", repository.publishedDraft.PageCount)
	}
	if len(repository.publishedDraft.Ranks) != 2 {
		t.Fatalf("published ranks=%d, want 2", len(repository.publishedDraft.Ranks))
	}
	if repository.publishedDraft.Ranks[0].MovieJavID != "JAV-001" || repository.publishedDraft.Ranks[1].MovieJavID != "JAV-002" {
		t.Fatalf("published ranks=%+v", repository.publishedDraft.Ranks)
	}
	if repository.publishedStat.RankRowCount != 2 {
		t.Fatalf("published stat rank rows=%d, want 2", repository.publishedStat.RankRowCount)
	}
}

func collectTestPage(pageNumber int64, checksum, javID string) dailybesttypes.SourcePage {
	return dailybesttypes.SourcePage{
		PageNumber: pageNumber,
		SourceURL:  fmt.Sprintf("https://example.com/cn/rank?page=%d", pageNumber),
		Checksum:   checksum,
		Content: `<div class="videothumblist"><div class="videos">
<div class="video"><div class="title">` + javID + ` 标题</div><div class="id">` + javID + `</div><div class="toolbar"><a id="` + javID + `"></a></div></div>
</div></div>`,
	}
}
