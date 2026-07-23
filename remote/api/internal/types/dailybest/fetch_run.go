package dailybest

const (
	RunStatusQueued    = "queued"
	RunStatusRunning   = "running"
	RunStatusSucceeded = "succeeded"
	RunStatusFailed    = "failed"
	RunStatusNoChange  = "no_change"

	TriggerScheduled = "scheduled"
	TriggerRecovery  = "recovery"
	TriggerManual    = "manual"
)
