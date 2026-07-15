package fetchsehuatang

import (
	"rudy_gc/internal/dep"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/service/fetchsite"
)

const (
	DefaultStartPage = int64(1)
	DefaultEndPage   = int64(1)

	SHFetchModeSearch    = "search"
	SHFetchModeForum     = "forum"
	SHFetchModeForumList = "forum_list"
	SHFetchModeDetail    = "detail"

	SHExistingPolicyUpsert = "upsert"
	SHExistingPolicySkip   = "skip"

	SHRunSourceFetch  = "fetch"
	SHRunSourceList   = "list"
	SHRunSourceDetail = "detail"

	SHForumAsiaUncensoredOriginal = "36"
	SHForumHDChineseSubtitle      = "103"

	SHRunStatusRunning = "running"
	SHRunStatusSuccess = "success"
	SHRunStatusFailed  = "failed"
	SHRunStatusStopped = "stopped"

	SHFetchItemStatusDiscovered      = "discovered"
	SHFetchItemStatusFetched         = "fetched"
	SHFetchItemStatusSkippedExisting = "skipped_existing"
	SHFetchItemStatusFailed          = "failed"

	SHDetailStatusPending = "pending"
	SHDetailStatusFetched = "fetched"
	SHDetailStatusFailed  = "failed"
)

type Service struct {
	deps *dep.Dep
}

type SHFetchRequest struct {
	FetchMode      string   `json:"fetch_mode"`
	ExistingPolicy string   `json:"existing_policy"`
	Keyword        string   `json:"keyword"`
	TargetForumIDs []string `json:"target_forum_ids"`
	StartPage      int64    `json:"start_page"`
	EndPage        int64    `json:"end_page"`
}

type SHFetchResult struct {
	RunID                int64    `json:"run_id"`
	FetchMode            string   `json:"fetch_mode"`
	ExistingPolicy       string   `json:"existing_policy"`
	Keyword              string   `json:"keyword"`
	TargetForumIDs       string   `json:"target_forum_ids"`
	StartPage            int64    `json:"start_page"`
	EndPage              int64    `json:"end_page"`
	TotalPageCount       int64    `json:"total_page_count"`
	CompletedPageCount   int64    `json:"completed_page_count"`
	CurrentPage          int64    `json:"current_page"`
	CurrentPageText      string   `json:"current_page_text"`
	CurrentForumID       string   `json:"current_forum_id"`
	HandledPageCount     int64    `json:"handled_page_count"`
	SuccessPageCount     int64    `json:"success_page_count"`
	FailedPageCount      int64    `json:"failed_page_count"`
	SeenCount            int64    `json:"seen_count"`
	MatchedCount         int64    `json:"matched_count"`
	InsertedCount        int64    `json:"inserted_count"`
	ExistingCount        int64    `json:"existing_count"`
	SkippedTopCount      int64    `json:"skipped_top_count"`
	SkippedExistingCount int64    `json:"skipped_existing_count"`
	SkippedForumCount    int64    `json:"skipped_forum_count"`
	FetchedCount         int64    `json:"fetched_count"`
	PersistFailedCount   int64    `json:"persist_failed_count"`
	FailedCount          int64    `json:"failed_count"`
	LastError            string   `json:"last_error"`
	DetailThreadTIDs     []string `json:"-"`
	aggTargets           map[string]moviex.ShThreadAggTarget
}

type remoteState struct {
	hasPreviousRemoteRequest bool
}

func NewService(d *dep.Dep) *Service {
	return &Service{deps: d}
}

func siteCode() string {
	return fetchsite.FetchSiteCodeSehuatang
}
