package dailybest

const (
	RuntimeLevelInfo  = "info"
	RuntimeLevelError = "error"

	RuntimeComponentAPI          = "api"
	RuntimeComponentBrowserAgent = "browser_agent"
	RuntimeComponentTarget       = "target"
)

type RuntimeEvent struct {
	ID            int64  `json:"id"`
	Component     string `json:"component"`
	EventType     string `json:"event_type"`
	Level         string `json:"level"`
	Message       string `json:"message"`
	FetchRunID    int64  `json:"fetch_run_id"`
	TargetProbeID int64  `json:"target_probe_id"`
	OccurredAt    int64  `json:"occurred_at"`
	CreatedOn     int64  `json:"created_on"`
}

type RuntimeEventDraft struct {
	Component     string
	EventType     string
	Level         string
	Message       string
	FetchRunID    int64
	TargetProbeID int64
	OccurredAt    int64
}

type RuntimeEventFilter struct {
	PageQuery
	Component string
	Level     string
}
