package util

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeExecution     ErrorType = "execution"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypeDatabase      ErrorType = "database"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeAssertion     ErrorType = "assertion"
)

// RobogoError represents a standardized error for the Robogo framework
type RobogoError struct {
	Type      ErrorType              `json:"type"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Cause     error                  `json:"cause,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Step      string                 `json:"step,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

func (e *RobogoError) Error() string {
	var parts []string

	if e.Action != "" {
		parts = append(parts, fmt.Sprintf("action=%s", e.Action))
	}
	if e.Step != "" {
		parts = append(parts, fmt.Sprintf("step=%s", e.Step))
	}

	parts = append(parts, fmt.Sprintf("type=%s", e.Type))
	parts = append(parts, e.Message)

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("cause=%v", e.Cause))
	}

	return strings.Join(parts, " | ")
}

// Unwrap returns the underlying cause error
func (e *RobogoError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]interface{}) *RobogoError {
	return &RobogoError{
		Type:    ErrorTypeValidation,
		Message: message,
		Details: details,
	}
}

// NewExecutionError creates a new execution error
func NewExecutionError(message string, cause error, action string) *RobogoError {
	return &RobogoError{
		Type:    ErrorTypeExecution,
		Message: message,
		Cause:   cause,
		Action:  action,
	}
}

// NewAssertionError creates a new assertion error
func NewAssertionError(message string, actual, expected interface{}, operator string) *RobogoError {
	return &RobogoError{
		Type:    ErrorTypeAssertion,
		Message: message,
		Details: map[string]interface{}{
			"actual":   actual,
			"expected": expected,
			"operator": operator,
		},
	}
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(message string, field string, value interface{}) *RobogoError {
	return &RobogoError{
		Type:    ErrorTypeConfiguration,
		Message: message,
		Details: map[string]interface{}{
			"field": field,
			"value": value,
		},
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, cause error, url string) *RobogoError {
	return &RobogoError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Cause:   cause,
		Details: map[string]interface{}{
			"url": url,
		},
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string, cause error, operation string) *RobogoError {
	return &RobogoError{
		Type:    ErrorTypeDatabase,
		Message: message,
		Cause:   cause,
		Details: map[string]interface{}{
			"operation": operation,
		},
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, duration string) *RobogoError {
	return &RobogoError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Details: map[string]interface{}{
			"timeout": duration,
		},
	}
}

// IsRobogoError checks if an error is a RobogoError
func IsRobogoError(err error) bool {
	_, ok := err.(*RobogoError)
	return ok
}

// GetRobogoError extracts RobogoError from an error chain
func GetRobogoError(err error) *RobogoError {
	if err == nil {
		return nil
	}

	if roboErr, ok := err.(*RobogoError); ok {
		return roboErr
	}

	// Check if it's wrapped
	if roboErr, ok := err.(interface{ Unwrap() error }); ok {
		return GetRobogoError(roboErr.Unwrap())
	}

	return nil
}
