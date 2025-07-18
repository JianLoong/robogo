package types

import (
	"fmt"
	"time"
)

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
	Status    ActionStatus `json:"status"`               // "pending", "running", "success", "error", "skipped"
	ErrorInfo *ErrorInfo   `json:"error_info,omitempty"` // Structured error information
	Data      any          `json:"data,omitempty"`       // Result data if status == "success"
	Meta      any          `json:"meta,omitempty"`       // Optional metadata (timing, logs, etc.)
}

// NewErrorResult creates an ActionResult with error status.
// Deprecated: Use ErrorBuilder for structured error handling
func NewErrorResult(msg string, args ...any) ActionResult {
	// Use SafeFormatter to prevent format string injection
	formatter := GetDefaultSafeFormatter()
	formatted, err := formatter.Format(msg, args...)
	if err != nil {
		// If formatting fails, create a safe fallback message
		formatted = fmt.Sprintf("Error formatting failed: %s (original message: %s)", err.Error(), msg)
	}

	errorInfo := &ErrorInfo{
		Category:  ErrorCategorySystem,
		Code:      "LEGACY_ERROR",
		Message:   formatted,
		Context:   make(map[string]any),
		Timestamp: time.Now(),
	}
	return ActionResult{
		Status:    ActionStatusError,
		ErrorInfo: errorInfo,
	}
}

// NewSuccessResult creates an ActionResult with passed status
func NewSuccessResult() ActionResult {
	return ActionResult{
		Status: ActionStatusPassed,
	}
}

// NewSuccessResultWithData creates an ActionResult with passed status and data
func NewSuccessResultWithData(data any) ActionResult {
	return ActionResult{
		Status: ActionStatusPassed,
		Data:   data,
	}
}

// NewSkippedResult creates an ActionResult with skipped status
func NewSkippedResult(reason string) ActionResult {
	errorInfo := &ErrorInfo{
		Category:  ErrorCategoryValidation,
		Code:      "SKIPPED",
		Message:   reason,
		Context:   make(map[string]any),
		Timestamp: time.Now(),
	}
	return ActionResult{
		Status:    ActionStatusSkipped,
		ErrorInfo: errorInfo,
	}
}

// GetErrorMessage returns the error message from ErrorInfo
func (ar *ActionResult) GetErrorMessage() string {
	if ar.ErrorInfo != nil {
		return ar.ErrorInfo.Message
	}
	return ""
}

// GetSkipReason returns the skip reason from ErrorInfo
func (ar *ActionResult) GetSkipReason() string {
	if ar.ErrorInfo != nil && ar.ErrorInfo.Category == ErrorCategoryValidation {
		return ar.ErrorInfo.Message
	}
	return ""
}

// IsError returns true if the result represents an error
func (ar *ActionResult) IsError() bool {
	return ar.Status == ActionStatusError || ar.Status == ActionStatusFailed
}

// IsSuccess returns true if the result represents success
func (ar *ActionResult) IsSuccess() bool {
	return ar.Status == ActionStatusPassed
}

// IsSkipped returns true if the result was skipped
func (ar *ActionResult) IsSkipped() bool {
	return ar.Status == ActionStatusSkipped
}
