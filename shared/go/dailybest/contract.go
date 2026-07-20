package dailybest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const (
	DailyBestCategoryMonth int64 = 1
)

type DailyBestSnapshotManifest struct {
	Revision     int64  `json:"revision"`
	BusinessDate string `json:"business_date"`
	Category     int64  `json:"category"`
	PageCount    int64  `json:"page_count"`
	ItemCount    int64  `json:"item_count"`
	Checksum     string `json:"checksum"`
	PublishedAt  int64  `json:"published_at"`
}

type DailyBestSnapshotListResponse struct {
	Snapshots      []DailyBestSnapshotManifest `json:"snapshots"`
	LatestRevision int64                       `json:"latest_revision"`
}

type DailyBestSnapshotPage struct {
	PageNumber int64  `json:"page_number"`
	SourceURL  string `json:"source_url"`
	Content    string `json:"content"`
	Checksum   string `json:"checksum"`
}

type DailyBestSnapshotItem struct {
	RankPos     int64  `json:"rank_pos"`
	SourcePage  int64  `json:"source_page"`
	SourceIndex int64  `json:"source_index"`
	JavID       string `json:"jav_id"`
	MovieName   string `json:"movie_name"`
	MovieCode   string `json:"movie_code"`
	Title       string `json:"title"`
	CoverURL    string `json:"cover_url"`
}

type DailyBestSnapshotResponse struct {
	DailyBestSnapshotManifest
	Pages []DailyBestSnapshotPage `json:"pages,omitempty"`
	Items []DailyBestSnapshotItem `json:"items"`
}

type DailyBestFetchRequest struct {
	BusinessDate string `json:"business_date"`
	Category     int64  `json:"category"`
}

type DailyBestFetchAcceptedResponse struct {
	RunID   int64  `json:"run_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type DailyBestFetchRun struct {
	ID           int64  `json:"id"`
	BusinessDate string `json:"business_date"`
	Category     int64  `json:"category"`
	TriggerType  string `json:"trigger_type"`
	Status       string `json:"status"`
	StartedAt    int64  `json:"started_at"`
	FinishedAt   int64  `json:"finished_at"`
	PageCount    int64  `json:"page_count"`
	ItemCount    int64  `json:"item_count"`
	SnapshotID   int64  `json:"snapshot_id"`
	ErrorMessage string `json:"error_message,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type DailyBestFetchRunListResponse struct {
	Runs []DailyBestFetchRun `json:"runs"`
}

type DailyBestAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ComputeDailyBestPageChecksum(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func ComputeDailyBestSnapshotChecksum(snapshot DailyBestSnapshotResponse) string {
	pages := append([]DailyBestSnapshotPage(nil), snapshot.Pages...)
	sort.Slice(pages, func(i, j int) bool { return pages[i].PageNumber < pages[j].PageNumber })
	items := append([]DailyBestSnapshotItem(nil), snapshot.Items...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].RankPos == items[j].RankPos {
			return items[i].JavID < items[j].JavID
		}
		return items[i].RankPos < items[j].RankPos
	})

	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s\n%d\n", strings.TrimSpace(snapshot.BusinessDate), snapshot.Category)
	for _, page := range pages {
		checksum := strings.TrimSpace(page.Checksum)
		if checksum == "" {
			checksum = ComputeDailyBestPageChecksum(page.Content)
		}
		_, _ = fmt.Fprintf(h, "p:%d:%s:%s\n", page.PageNumber, strings.TrimSpace(page.SourceURL), checksum)
	}
	for _, item := range items {
		_, _ = fmt.Fprintf(
			h,
			"i:%d:%d:%d:%s:%s:%s:%s:%s\n",
			item.RankPos,
			item.SourcePage,
			item.SourceIndex,
			strings.TrimSpace(item.JavID),
			strings.TrimSpace(item.MovieName),
			strings.TrimSpace(item.MovieCode),
			strings.TrimSpace(item.Title),
			strings.TrimSpace(item.CoverURL),
		)
	}
	return hex.EncodeToString(h.Sum(nil))
}
