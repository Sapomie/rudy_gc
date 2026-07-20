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
	RiseMovieCount      int64  `json:"rise_movie_count"`
	FallMovieCount      int64  `json:"fall_movie_count"`
	UnchangedMovieCount int64  `json:"unchanged_movie_count"`
	SourceChecksum      string `json:"source_checksum"`
	CalculatedOn        int64  `json:"calculated_on"`
	CreatedOn           int64  `json:"created_on"`
	UpdatedOn           int64  `json:"updated_on"`
}

type DailyBestRankDayStatListResponse struct {
	Stats    []DailyBestRankDayStat `json:"stats"`
	PageInfo *DailyBestPageInfo     `json:"page_info,omitempty"`
}

type DailyBestRankPeriod struct {
	ID             int64  `json:"id"`
	PeriodType     int64  `json:"period_type"`
	PeriodKey      string `json:"period_key"`
	Category       int64  `json:"category"`
	StartDayNumber int64  `json:"start_day_number"`
	EndDayNumber   int64  `json:"end_day_number"`
	PickDays       int64  `json:"pick_days"`
	TopN           int64  `json:"top_n"`
	Status         int64  `json:"status"`
	CreatedOn      int64  `json:"created_on"`
	UpdatedOn      int64  `json:"updated_on"`
}

type DailyBestRankPeriodListResponse struct {
	Periods  []DailyBestRankPeriod `json:"periods"`
	PageInfo *DailyBestPageInfo    `json:"page_info,omitempty"`
}

type DailyBestRankPeriodItem struct {
	ID                 int64               `json:"id"`
	PeriodID           int64               `json:"period_id"`
	MovieJavID         string              `json:"movie_jav_id"`
	Movie              DailyBestMovieBasic `json:"movie"`
	RankPos            int64               `json:"rank_pos"`
	Score              float64             `json:"score"`
	DaysInRank         int64               `json:"days_in_rank"`
	UsedPickDays       int64               `json:"used_pick_days"`
	FirstRankDayNumber int64               `json:"first_rank_day_number"`
	LastRankDayNumber  int64               `json:"last_rank_day_number"`
	BestRank           int64               `json:"best_rank"`
	WorstPickedRank    int64               `json:"worst_picked_rank"`
	PrevRank           int64               `json:"prev_rank"`
	RankChange         int64               `json:"rank_change"`
	PeakRank           int64               `json:"peak_rank"`
	CreatedOn          int64               `json:"created_on"`
	UpdatedOn          int64               `json:"updated_on"`
}

type DailyBestRankPeriodItemListResponse struct {
	Items    []DailyBestRankPeriodItem `json:"items"`
	PageInfo *DailyBestPageInfo        `json:"page_info,omitempty"`
}

type DailyBestRankPeriodStat struct {
	ID                   int64               `json:"id"`
	PeriodID             int64               `json:"period_id"`
	PrevPeriodID         int64               `json:"prev_period_id"`
	SourceRankRowCount   int64               `json:"source_rank_row_count"`
	SourceRankDayCount   int64               `json:"source_rank_day_count"`
	SourceMovieCount     int64               `json:"source_movie_count"`
	MissingRankDayCount  int64               `json:"missing_rank_day_count"`
	PeriodMovieCount     int64               `json:"period_movie_count"`
	ContinuedMovieCount  int64               `json:"continued_movie_count"`
	FirstSeenMovieCount  int64               `json:"first_seen_movie_count"`
	ReturningMovieCount  int64               `json:"returning_movie_count"`
	DroppedMovieCount    int64               `json:"dropped_movie_count"`
	RiseMovieCount       int64               `json:"rise_movie_count"`
	FallMovieCount       int64               `json:"fall_movie_count"`
	UnchangedMovieCount  int64               `json:"unchanged_movie_count"`
	PeakMovieCount       int64               `json:"peak_movie_count"`
	RankOneMovieJavID    string              `json:"rank_one_movie_jav_id"`
	RankOneMovie         DailyBestMovieBasic `json:"rank_one_movie"`
	AverageScore         float64             `json:"average_score"`
	AverageDaysInRank    float64             `json:"average_days_in_rank"`
	SourceFetchedStartOn int64               `json:"source_fetched_start_on"`
	SourceFetchedEndOn   int64               `json:"source_fetched_end_on"`
	SourceChecksum       string              `json:"source_checksum"`
	CalculatedOn         int64               `json:"calculated_on"`
	CreatedOn            int64               `json:"created_on"`
	UpdatedOn            int64               `json:"updated_on"`
}

type DailyBestRankPeriodStatListResponse struct {
	Stats    []DailyBestRankPeriodStat `json:"stats"`
	PageInfo *DailyBestPageInfo        `json:"page_info,omitempty"`
}

type DailyBestRankSyncManifest struct {
	FetchRunID   int64  `json:"fetch_run_id"`
	BusinessDate string `json:"business_date"`
	DayNumber    int64  `json:"day_number"`
	Category     int64  `json:"category"`
	PageCount    int64  `json:"page_count"`
	ItemCount    int64  `json:"item_count"`
	Checksum     string `json:"checksum"`
	FinishedAt   int64  `json:"finished_at"`
}

type DailyBestRankSyncManifestListResponse struct {
	Manifests []DailyBestRankSyncManifest `json:"manifests"`
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
