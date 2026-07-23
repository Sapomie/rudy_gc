package browseragent

const (
	StateReady    = "ready"
	StateFetching = "fetching"

	PurposeDailyBest = "dailybest"
	PurposeProbe     = "probe"
)

type FetchResult struct {
	RequestID    string `json:"request_id"`
	StatusCode   int    `json:"status_code"`
	Body         string `json:"body"`
	FinalURL     string `json:"final_url"`
	DurationMS   int64  `json:"duration_ms"`
	StartedAt    int64  `json:"started_at"`
	FinishedAt   int64  `json:"finished_at"`
	BrowserState string `json:"browser_state"`
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
}

type Status struct {
	State           string `json:"state"`
	LastHeartbeatAt int64  `json:"last_heartbeat_at"`
	LastError       string `json:"last_error"`
	Version         string `json:"version"`
	UserAgent       string `json:"user_agent"`
}
