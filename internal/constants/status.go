package constants

// ActionStatus represents the lifecycle state of an action.
type ActionStatus string

const (
	ActionStatusPassed  ActionStatus = "passed"
	ActionStatusFailed  ActionStatus = "failed"
	ActionStatusError   ActionStatus = "error"
	ActionStatusSkipped ActionStatus = "skipped"
)
