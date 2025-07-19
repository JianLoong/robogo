package constants

// ActionStatus represents the lifecycle state of an action.
type ActionStatus string

const (
	ActionStatusPending ActionStatus = "pending"
	ActionStatusRunning ActionStatus = "running"
	ActionStatusPassed  ActionStatus = "passed"
	ActionStatusFailed  ActionStatus = "failed"
	ActionStatusError   ActionStatus = "error"
	ActionStatusSkipped ActionStatus = "skipped"
)