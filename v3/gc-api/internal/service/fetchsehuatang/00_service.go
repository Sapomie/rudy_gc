package fetchsehuatang

import (
	"strings"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"
)

const (
	DefaultListURL     = "https://vzzr.qnc8.net/forum-103-1.html"
	DefaultKeyword     = ""
	DefaultStartPage   = int64(1)
	DefaultEndPage     = int64(1)
	PersistModeUpsert  = "upsert_all"
	PersistModeSkipOld = "skip_existing"
	DefaultPersistMode = PersistModeSkipOld

	SHFetchModeSearch    = "search"
	SHFetchModeForum     = "forum"
	SHFetchModeForumList = "forum_list"
	SHFetchModeDetail    = "detail"

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

type FetchRequest struct {
	ListURL        string `json:"list_url"`
	Keyword        string `json:"keyword"`
	StartPage      int64  `json:"start_page"`
	EndPage        int64  `json:"end_page"`
	PersistMode    string `json:"persist_mode"`
	StopOnExisting bool   `json:"stop_on_existing"`
}

type FetchTopicItem struct {
	Title      string   `json:"title"`
	DetailURL  string   `json:"detail_url"`
	MovieJavID string   `json:"movie_jav_id"`
	MovieName  string   `json:"movie_name"`
	PostTime   int64    `json:"post_time"`
	PostDate   int64    `json:"post_date"`
	Magnets    []string `json:"magnets"`
	InfoHashes []string `json:"info_hashes"`
	Error      string   `json:"error,omitempty"`
}

type FetchResult struct {
	ListURL           string            `json:"list_url"`
	Keyword           string            `json:"keyword"`
	StartPage         int64             `json:"start_page"`
	EndPage           int64             `json:"end_page"`
	HandledPageCount  int64             `json:"handled_page_count"`
	SkippedTopCount   int               `json:"skipped_top_count"`
	MatchedCount      int               `json:"matched_count"`
	SuccessCount      int               `json:"success_count"`
	FailedCount       int               `json:"failed_count"`
	InsertedCount     int               `json:"inserted_count"`
	UpdatedCount      int               `json:"updated_count"`
	PersistFailCount  int               `json:"persist_fail_count"`
	SkippedExisting   int               `json:"skipped_existing"`
	StopOnExisting    bool              `json:"stop_on_existing"`
	StoppedOnExisting bool              `json:"stopped_on_existing"`
	StopExistingPage  int64             `json:"stop_existing_page"`
	StopExistingTitle string            `json:"stop_existing_title"`
	StopExistingURL   string            `json:"stop_existing_url"`
	Items             []*FetchTopicItem `json:"items"`
}

type SHFetchRequest struct {
	FetchMode      string   `json:"fetch_mode"`
	Keyword        string   `json:"keyword"`
	TargetForumIDs []string `json:"target_forum_ids"`
	StartPage      int64    `json:"start_page"`
	EndPage        int64    `json:"end_page"`
}

type SHFetchResult struct {
	RunID                int64    `json:"run_id"`
	FetchMode            string   `json:"fetch_mode"`
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

func defaultRequest(req FetchRequest) FetchRequest {
	out := req
	if out.ListURL == "" {
		out.ListURL = DefaultListURL
	}
	if strings.TrimSpace(out.Keyword) == "" {
		out.Keyword = DefaultKeyword
	}
	if out.StartPage == 0 {
		out.StartPage = DefaultStartPage
	}
	if out.EndPage == 0 {
		out.EndPage = DefaultEndPage
	}
	out.PersistMode = NormalizePersistMode(out.PersistMode)
	return out
}

func NormalizePersistMode(raw string) string {
	switch strings.TrimSpace(raw) {
	case PersistModeSkipOld:
		return PersistModeSkipOld
	default:
		return PersistModeUpsert
	}
}

func siteCode() string {
	return fetchsite.FetchSiteCodeSehuatang
}
