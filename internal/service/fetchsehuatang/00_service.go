package fetchsehuatang

import (
	"strings"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/fetchsite"
)

const (
	DefaultListURL     = "https://vzzr.qnc8.net/forum-103-1.html"
	DefaultKeyword     = ""
	DefaultStartPage   = int64(1)
	DefaultEndPage     = int64(1)
	PersistModeUpsert  = "upsert_all"
	PersistModeSkipOld = "skip_existing"
	DefaultPersistMode = PersistModeUpsert
)

type Service struct {
	deps *dep.Dep
}

type FetchRequest struct {
	ListURL     string `json:"list_url"`
	Keyword     string `json:"keyword"`
	StartPage   int64  `json:"start_page"`
	EndPage     int64  `json:"end_page"`
	PersistMode string `json:"persist_mode"`
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
	ListURL          string            `json:"list_url"`
	Keyword          string            `json:"keyword"`
	StartPage        int64             `json:"start_page"`
	EndPage          int64             `json:"end_page"`
	HandledPageCount int64             `json:"handled_page_count"`
	SkippedTopCount  int               `json:"skipped_top_count"`
	MatchedCount     int               `json:"matched_count"`
	SuccessCount     int               `json:"success_count"`
	FailedCount      int               `json:"failed_count"`
	InsertedCount    int               `json:"inserted_count"`
	UpdatedCount     int               `json:"updated_count"`
	PersistFailCount int               `json:"persist_fail_count"`
	SkippedExisting  int               `json:"skipped_existing"`
	Items            []*FetchTopicItem `json:"items"`
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
