package dailybest

import browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"

type RuntimeStatus struct {
	APIStatus        string                   `json:"api_status"`
	DatabaseStatus   string                   `json:"database_status"`
	BrowserStatus    browseragenttypes.Status `json:"browser_status"`
	BrowserError     string                   `json:"browser_error"`
	TargetStatus     string                   `json:"target_status"`
	LatestProbe      *TargetProbe             `json:"latest_probe,omitempty"`
	NextProbeAt      int64                    `json:"next_probe_at"`
	LastProbeSuccess int64                    `json:"last_probe_success"`
	LastProbeFailure int64                    `json:"last_probe_failure"`
}
