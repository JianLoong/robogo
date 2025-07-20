package types

import (
	"time"

	"github.com/JianLoong/robogo/internal/constants"
)

// ActionStatus is now defined in constants package
type ActionStatus = constants.ActionStatus

const (
	ActionStatusPassed  = constants.ActionStatusPassed
	ActionStatusFailed  = constants.ActionStatusFailed
	ActionStatusError   = constants.ActionStatusError
	ActionStatusSkipped = constants.ActionStatusSkipped
)

// ActionResult is the public, consistent result type for all actions.
type ActionResult struct {
	Status      ActionStatus `json:"status"`                 // "pending", "running", "success", "error", "skipped"
	ErrorInfo   *ErrorInfo   `json:"error_info,omitempty"`   // Structured error information (technical errors)
	FailureInfo *FailureInfo `json:"failure_info,omitempty"` // Structured failure information (logical failures)
	Data        any          `json:"data,omitempty"`         // Result data if status == "success"
	Meta        any          `json:"meta,omitempty"`         // Optional metadata (timing, logs, etc.)
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
		Timestamp: time.Now(),
	}
	return ActionResult{
		Status:    ActionStatusSkipped,
		ErrorInfo: errorInfo,
	}
}

// GetMessage returns the error or failure message
func (ar *ActionResult) GetMessage() string {
	if ar.ErrorInfo != nil {
		return ar.ErrorInfo.Message
	}
	if ar.FailureInfo != nil {
		return ar.FailureInfo.Message
	}
	return ""
}

// GetErrorMessage returns the error or failure message
func (ar *ActionResult) GetErrorMessage() string {
	return ar.GetMessage()
}

// GetSkipReason returns the skip reason from ErrorInfo
func (ar *ActionResult) GetSkipReason() string {
	if ar.ErrorInfo != nil && ar.ErrorInfo.Category == ErrorCategoryValidation {
		return ar.ErrorInfo.Message
	}
	return ""
}

// IsError returns true if the result represents a technical error
func (ar *ActionResult) IsError() bool {
	return ar.Status == ActionStatusError
}

// IsFailed returns true if the result represents a logical failure
func (ar *ActionResult) IsFailed() bool {
	return ar.Status == ActionStatusFailed
}

// HasIssue returns true if the result has either an error or failure
func (ar *ActionResult) HasIssue() bool {
	return ar.IsError() || ar.IsFailed()
}

// IsSuccess returns true if the result represents success
func (ar *ActionResult) IsSuccess() bool {
	return ar.Status == ActionStatusPassed
}

// IsSkipped returns true if the result was skipped
func (ar *ActionResult) IsSkipped() bool {
	return ar.Status == ActionStatusSkipped
}
