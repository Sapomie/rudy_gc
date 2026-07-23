package dailybest

const (
	TargetProbeStatusRunning     = "running"
	TargetProbeStatusSucceeded   = "succeeded"
	TargetProbeStatusFailed      = "failed"
	TargetProbeStatusInterrupted = "interrupted"

	TargetStatusUnknown     = "unknown"
	TargetStatusAvailable   = "available"
	TargetStatusUnavailable = "unavailable"
)

type TargetProbe struct {
	ID            int64  `json:"id"`
	RequestID     string `json:"request_id"`
	Status        string `json:"status"`
	TargetURL     string `json:"target_url"`
	FinalURL      string `json:"final_url"`
	HTTPStatus    int64  `json:"http_status"`
	DurationMS    int64  `json:"duration_ms"`
	BrowserStatus string `json:"browser_status"`
	ErrorType     string `json:"error_type"`
	ErrorMessage  string `json:"error_message"`
	StartedAt     int64  `json:"started_at"`
	FinishedAt    int64  `json:"finished_at"`
	CreatedOn     int64  `json:"created_on"`
	UpdatedOn     int64  `json:"updated_on"`
}

type TargetProbeDraft struct {
	RequestID string
	TargetURL string
}

type TargetProbeResult struct {
	Status        string
	FinalURL      string
	HTTPStatus    int64
	DurationMS    int64
	BrowserStatus string
	ErrorType     string
	ErrorMessage  string
	StartedAt     int64
	FinishedAt    int64
}

type TargetProbeFilter struct {
	PageQuery
	Status      string
	StartedFrom int64
	StartedTo   int64
}
