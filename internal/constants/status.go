package constants

// ActionStatus represents the lifecycle state of an action.
type ActionStatus string

const (
	ActionStatusPassed  ActionStatus = "PASS"
	ActionStatusFailed  ActionStatus = "FAIL"
	ActionStatusError   ActionStatus = "ERROR"
	ActionStatusSkipped ActionStatus = "SKIPPED"
)
