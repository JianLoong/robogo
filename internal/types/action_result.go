package types

import "fmt"

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

// ActionResult is the public, consistent result type for all actions.
type ActionResult struct {
	Status ActionStatus `json:"status"`           // "pending", "running", "success", "error", "skipped"
	Error  string       `json:"error,omitempty"`  // Error message if status == "error"
	Reason string       `json:"reason,omitempty"` // Reason for skip if status == "skipped"
	Data   interface{}  `json:"data,omitempty"`   // Result data if status == "success"
	Output string       `json:"output,omitempty"` // Human-readable summary for logs/UI
	Meta   interface{}  `json:"meta,omitempty"`   // Optional metadata (timing, logs, etc.)
}

// NewErrorResult creates an ActionResult with error status and a Go error.
func NewErrorResult(msg string, args ...interface{}) (ActionResult, error) {
	formatted := fmt.Sprintf(msg, args...)
	return ActionResult{
		Status: ActionStatusError,
		Error:  formatted,
		Output: formatted,
	}, fmt.Errorf(formatted)
}
