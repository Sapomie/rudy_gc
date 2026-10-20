package dailybest

import contracts "rudy_gc/shared/dailybest"

const (
	RankDayStatusMissing         = "missing"
	RankDayStatusIncomplete      = "incomplete"
	RankDayStatusVersionMismatch = "version_mismatch"
	RankDayStatusPending         = "pending"
	RankDayStatusFailed          = "failed"
	RankDayStatusCompleted       = "completed"
)

type RankDayQuery struct {
	BusinessDateStart string
	BusinessDateEnd   string
	DayNumber         int64
	Page              int64
	PageSize          int64
}

type SyncTarget struct {
	DayNumber int64 `json:"day_number"`
	Category  int64 `json:"category"`
}

type SyncRequest struct {
	Targets []SyncTarget `json:"targets"`
}

type SyncManifest struct {
	FetchRunID   int64
	BusinessDate string
	DayNumber    int64
	Category     int64
	PageCount    int64
	ItemCount    int64
	Checksum     string
	FinishedAt   int64
}

// RankBatch 是 DailyBest 排名进入本地 c_rank 前的统一数据结构。
// 本地直接抓取和远端同步只负责生成该结构，后续写库与聚合共用同一处理链路。
type RankBatch struct {
	BusinessDate string
	DayNumber    int64
	Category     int64
	FetchedOn    int64
	SyncedOn     int64
	Ranks        []contracts.DailyBestRank
}

type RankDayTiming struct {
	DayNumber int64
	FetchedOn int64
	SyncedOn  int64
}

type RankCategoryKey struct {
	DayNumber int64
	Category  int64
}

type LocalRankCategoryQuery struct {
	BusinessDate string
	DayNumber    int64
	Category     int64
}

type LocalRankCategoryState struct {
	RankCount        int64  `json:"rank_count"`
	Checksum         string `json:"checksum"`
	RemoteFetchRunID int64  `json:"remote_fetch_run_id"`
	SyncStatus       string `json:"sync_status"`
	ProcessStatus    string `json:"process_status"`
	SyncedAt         int64  `json:"synced_at"`
	ProcessedAt      int64  `json:"processed_at"`
	LastError        string `json:"last_error"`
}

type RankDayCategoryRow struct {
	Remote    contracts.DailyBestRankDayCategory `json:"remote"`
	Local     LocalRankCategoryState             `json:"local"`
	Status    string                             `json:"status"`
	NeedsSync bool                               `json:"needs_sync"`
}

type RankDayRow struct {
	BusinessDate    string               `json:"business_date"`
	DayNumber       int64                `json:"day_number"`
	PageCount       int64                `json:"page_count"`
	ItemCount       int64                `json:"item_count"`
	RemoteRankCount int64                `json:"remote_rank_count"`
	LocalRankCount  int64                `json:"local_rank_count"`
	FinishedAt      int64                `json:"finished_at"`
	Status          string               `json:"status"`
	NeedsSync       bool                 `json:"needs_sync"`
	Categories      []RankDayCategoryRow `json:"categories"`
}

type RankDayListResponse struct {
	Days     []RankDayRow
	Total    int64
	Page     int64
	PageSize int64
}
