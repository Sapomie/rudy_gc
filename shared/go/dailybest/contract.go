package dailybest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	DailyBestCategoryMonth int64 = 1
	DailyBestGenesisDate         = "2023-01-01"
)

type DailyBestPageInfo struct {
	Total     int64 `json:"total"`
	Page      int64 `json:"page"`
	PageSize  int64 `json:"page_size"`
	PageTotal int64 `json:"page_total"`
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
	Checksum     string `json:"checksum"`
	ErrorMessage string `json:"error_message,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type DailyBestFetchRunListResponse struct {
	Runs     []DailyBestFetchRun `json:"runs"`
	PageInfo *DailyBestPageInfo  `json:"page_info,omitempty"`
}

type DailyBestMovieBasic struct {
	MovieName string `json:"movie_name"`
	MovieCode string `json:"movie_code"`
	Title     string `json:"title"`
	CoverURL  string `json:"cover_url"`
}

type DailyBestRank struct {
	ID         int64               `json:"id"`
	RankKey    string              `json:"rank_key"`
	MovieJavID string              `json:"movie_jav_id"`
	DayNumber  int64               `json:"day_number"`
	RankPos    int64               `json:"rank_pos"`
	Category   int64               `json:"category"`
	Movie      DailyBestMovieBasic `json:"movie"`
	CreatedOn  int64               `json:"created_on"`
	UpdatedOn  int64               `json:"updated_on"`
}

type DailyBestRankListResponse struct {
	Ranks    []DailyBestRank    `json:"ranks"`
	PageInfo *DailyBestPageInfo `json:"page_info,omitempty"`
}

type DailyBestRankDayStat struct {
	ID                  int64  `json:"id"`
	DayNumber           int64  `json:"day_number"`
	RankDate            string `json:"rank_date"`
	PrevDayNumber       int64  `json:"prev_day_number"`
	RankRowCount        int64  `json:"rank_row_count"`
	MovieCount          int64  `json:"movie_count"`
	ContinuedMovieCount int64  `json:"continued_movie_count"`
	FirstSeenMovieCount int64  `json:"first_seen_movie_count"`
	ReturningMovieCount int64  `json:"returning_movie_count"`
	DroppedMovieCount   int64  `json:"dropped_movie_count"`
	SourceChecksum      string `json:"source_checksum"`
	FetchedAt           int64  `json:"fetched_at"`
	SyncedAt            int64  `json:"synced_at"`
	CalculatedOn        int64  `json:"calculated_on"`
	CreatedOn           int64  `json:"created_on"`
	UpdatedOn           int64  `json:"updated_on"`
}

type DailyBestRankDayStatListResponse struct {
	Stats    []DailyBestRankDayStat `json:"stats"`
	PageInfo *DailyBestPageInfo     `json:"page_info,omitempty"`
}

type DailyBestRankDayCategory struct {
	FetchRunID int64  `json:"fetch_run_id"`
	Category   int64  `json:"category"`
	PageCount  int64  `json:"page_count"`
	ItemCount  int64  `json:"item_count"`
	RankCount  int64  `json:"rank_count"`
	Checksum   string `json:"checksum"`
	FinishedAt int64  `json:"finished_at"`
}

type DailyBestRankDay struct {
	BusinessDate string                     `json:"business_date"`
	DayNumber    int64                      `json:"day_number"`
	PageCount    int64                      `json:"page_count"`
	ItemCount    int64                      `json:"item_count"`
	RankCount    int64                      `json:"rank_count"`
	FinishedAt   int64                      `json:"finished_at"`
	Categories   []DailyBestRankDayCategory `json:"categories"`
}

type DailyBestRankDayListResponse struct {
	Days     []DailyBestRankDay `json:"days"`
	PageInfo *DailyBestPageInfo `json:"page_info,omitempty"`
}

type DailyBestAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RankDayNumber(businessDate string) (int64, error) {
	date, err := time.ParseInLocation(time.DateOnly, strings.TrimSpace(businessDate), time.UTC)
	if err != nil {
		return 0, err
	}
	genesis, _ := time.ParseInLocation(time.DateOnly, DailyBestGenesisDate, time.UTC)
	return int64(date.Sub(genesis)/(24*time.Hour)) + 1, nil
}

func RankDate(dayNumber int64) string {
	genesis, _ := time.ParseInLocation(time.DateOnly, DailyBestGenesisDate, time.UTC)
	return genesis.AddDate(0, 0, int(dayNumber-1)).Format(time.DateOnly)
}

func ComputeDailyBestRankChecksum(businessDate string, dayNumber, category int64, ranks []DailyBestRank) string {
	ordered := append([]DailyBestRank(nil), ranks...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].RankPos != ordered[j].RankPos {
			return ordered[i].RankPos < ordered[j].RankPos
		}
		return ordered[i].MovieJavID < ordered[j].MovieJavID
	})
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "v1:%s:%d:%d\n", strings.TrimSpace(businessDate), dayNumber, category)
	for _, rank := range ordered {
		_, _ = fmt.Fprintf(h, "%d:%s\n", rank.RankPos, strings.TrimSpace(rank.MovieJavID))
	}
	return hex.EncodeToString(h.Sum(nil))
}
