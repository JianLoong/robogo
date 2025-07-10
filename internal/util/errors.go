package util

import (
	"fmt"
	"strings"
	"time"
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
	ErrorTypeMessaging     ErrorType = "messaging"
	ErrorTypeTemplate      ErrorType = "template"
	ErrorTypeSecurity      ErrorType = "security"
	ErrorTypeFileSystem    ErrorType = "filesystem"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// ErrorCategory provides metadata about error types
type ErrorCategory struct {
	Type        ErrorType
	Severity    ErrorSeverity
	Retryable   bool
	Recoverable bool
	UserAction  string
	TechAction  string
}

// ErrorCatalog defines characteristics of each error type
var ErrorCatalog = map[ErrorType]ErrorCategory{
	ErrorTypeValidation: {
		Type:        ErrorTypeValidation,
		Severity:    SeverityLow,
		Retryable:   false,
		Recoverable: false,
		UserAction:  "Fix the invalid input and retry",
		TechAction:  "Validate input before processing",
	},
	ErrorTypeExecution: {
		Type:        ErrorTypeExecution,
		Severity:    SeverityMedium,
		Retryable:   true,
		Recoverable: true,
		UserAction:  "Check configuration and retry",
		TechAction:  "Implement retry with exponential backoff",
	},
	ErrorTypeNetwork: {
		Type:        ErrorTypeNetwork,
		Severity:    SeverityMedium,
		Retryable:   true,
		Recoverable: true,
		UserAction:  "Check network connectivity and retry",
		TechAction:  "Implement exponential backoff retry",
	},
	ErrorTypeDatabase: {
		Type:        ErrorTypeDatabase,
		Severity:    SeverityHigh,
		Retryable:   true,
		Recoverable: true,
		UserAction:  "Check database connection and retry",
		TechAction:  "Implement connection pooling and retry",
	},
	ErrorTypeTimeout: {
		Type:        ErrorTypeTimeout,
		Severity:    SeverityMedium,
		Retryable:   true,
		Recoverable: true,
		UserAction:  "Increase timeout value and retry",
		TechAction:  "Implement adaptive timeout strategies",
	},
	ErrorTypeAssertion: {
		Type:        ErrorTypeAssertion,
		Severity:    SeverityLow,
		Retryable:   false,
		Recoverable: false,
		UserAction:  "Fix the test assertion and retry",
		TechAction:  "Validate assertion logic",
	},
	ErrorTypeMessaging: {
		Type:        ErrorTypeMessaging,
		Severity:    SeverityHigh,
		Retryable:   true,
		Recoverable: true,
		UserAction:  "Check message broker connection and retry",
		TechAction:  "Implement message broker retry and circuit breaker",
	},
	ErrorTypeTemplate: {
		Type:        ErrorTypeTemplate,
		Severity:    SeverityMedium,
		Retryable:   false,
		Recoverable: true,
		UserAction:  "Fix template syntax and retry",
		TechAction:  "Validate template before processing",
	},
	ErrorTypeSecurity: {
		Type:        ErrorTypeSecurity,
		Severity:    SeverityCritical,
		Retryable:   false,
		Recoverable: false,
		UserAction:  "Check security credentials and permissions",
		TechAction:  "Implement secure credential management",
	},
	ErrorTypeFileSystem: {
		Type:        ErrorTypeFileSystem,
		Severity:    SeverityMedium,
		Retryable:   true,
		Recoverable: true,
		UserAction:  "Check file permissions and path",
		TechAction:  "Implement file operation retry",
	},
}

// Standard error message templates
const (
	MsgArgumentCount     = "requires %d arguments, got %d"
	MsgArgumentType      = "argument %d must be %s, got %s"
	MsgArgumentValue     = "argument %d has invalid value: %v"
	MsgConnectionFailed  = "failed to connect to %s"
	MsgOperationFailed   = "operation '%s' failed"
	MsgValidationFailed  = "validation failed for %s"
	MsgTimeoutExceeded   = "operation timed out after %v"
	MsgResourceNotFound  = "resource '%s' not found"
	MsgPermissionDenied  = "permission denied for operation '%s'"
	MsgConfigurationMissing = "missing required configuration: %s"
	MsgDependencyUnavailable = "dependency '%s' is unavailable"
)

// RobogoError represents a standardized error for the Robogo framework
type RobogoError struct {
	Type         ErrorType              `json:"type"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Cause        error                  `json:"cause,omitempty"`
	Action       string                 `json:"action,omitempty"`
	Step         string                 `json:"step,omitempty"`
	TestCase     string                 `json:"test_case,omitempty"`
	Timestamp    string                 `json:"timestamp,omitempty"`
	Severity     ErrorSeverity          `json:"severity,omitempty"`
	Retryable    bool                   `json:"retryable,omitempty"`
	Recoverable  bool                   `json:"recoverable,omitempty"`
	UserAction   string                 `json:"user_action,omitempty"`
	TechAction   string                 `json:"tech_action,omitempty"`
	Arguments    []interface{}          `json:"arguments,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// ErrorBuilder provides a fluent interface for building RobogoError instances
type ErrorBuilder struct {
	error *RobogoError
}

// NewErrorBuilder creates a new error builder with the specified type and message
func NewErrorBuilder(errorType ErrorType, message string) *ErrorBuilder {
	category := ErrorCatalog[errorType]
	return &ErrorBuilder{
		error: &RobogoError{
			Type:        errorType,
			Message:     message,
			Details:     make(map[string]interface{}),
			Timestamp:   time.Now().Format(time.RFC3339),
			Severity:    category.Severity,
			Retryable:   category.Retryable,
			Recoverable: category.Recoverable,
			UserAction:  category.UserAction,
			TechAction:  category.TechAction,
		},
	}
}

// WithCause adds the underlying cause error
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.error.Cause = cause
	return b
}

// WithAction sets the action name
func (b *ErrorBuilder) WithAction(action string) *ErrorBuilder {
	b.error.Action = action
	return b
}

// WithStep sets the step name
func (b *ErrorBuilder) WithStep(step string) *ErrorBuilder {
	b.error.Step = step
	return b
}

// WithTestCase sets the test case name
func (b *ErrorBuilder) WithTestCase(testCase string) *ErrorBuilder {
	b.error.TestCase = testCase
	return b
}

// WithDetails adds or updates error details
func (b *ErrorBuilder) WithDetails(details map[string]interface{}) *ErrorBuilder {
	for k, v := range details {
		b.error.Details[k] = v
	}
	return b
}

// WithArguments adds the arguments that caused the error
func (b *ErrorBuilder) WithArguments(args []interface{}) *ErrorBuilder {
	b.error.Arguments = args
	return b
}

// WithOptions adds the options that were used
func (b *ErrorBuilder) WithOptions(options map[string]interface{}) *ErrorBuilder {
	b.error.Options = options
	return b
}

// WithUserAction overrides the default user action message
func (b *ErrorBuilder) WithUserAction(userAction string) *ErrorBuilder {
	b.error.UserAction = userAction
	return b
}

// WithTechAction overrides the default technical action message
func (b *ErrorBuilder) WithTechAction(techAction string) *ErrorBuilder {
	b.error.TechAction = techAction
	return b
}

// WithRetryable overrides the default retryable flag
func (b *ErrorBuilder) WithRetryable(retryable bool) *ErrorBuilder {
	b.error.Retryable = retryable
	return b
}

// WithRecoverable overrides the default recoverable flag
func (b *ErrorBuilder) WithRecoverable(recoverable bool) *ErrorBuilder {
	b.error.Recoverable = recoverable
	return b
}

// Build creates the final RobogoError instance
func (b *ErrorBuilder) Build() *RobogoError {
	return b.error
}

func (e *RobogoError) Error() string {
	var parts []string

	if e.TestCase != "" {
		parts = append(parts, fmt.Sprintf("test_case=%s", e.TestCase))
	}
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

// GetCategory returns the error category metadata
func (e *RobogoError) GetCategory() ErrorCategory {
	return ErrorCatalog[e.Type]
}

// IsRetryable returns true if the error is retryable
func (e *RobogoError) IsRetryable() bool {
	return e.Retryable
}

// IsRecoverable returns true if the error is recoverable
func (e *RobogoError) IsRecoverable() bool {
	return e.Recoverable
}

// GetSeverity returns the error severity
func (e *RobogoError) GetSeverity() ErrorSeverity {
	return e.Severity
}

// GetUserAction returns the recommended user action
func (e *RobogoError) GetUserAction() string {
	return e.UserAction
}

// GetTechAction returns the recommended technical action
func (e *RobogoError) GetTechAction() string {
	return e.TechAction
}

// Unwrap returns the underlying cause error
func (e *RobogoError) Unwrap() error {
	return e.Cause
}

// WithType sets the error type (for method chaining)
func (e *RobogoError) WithType(errorType ErrorType) *RobogoError {
	e.Type = errorType
	return e
}

// WithDetails adds or updates details (for method chaining)
func (e *RobogoError) WithDetails(details map[string]interface{}) *RobogoError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithStep sets the step name (for method chaining)
func (e *RobogoError) WithStep(step string) *RobogoError {
	e.Step = step
	return e
}

// WithAction sets the action name (for method chaining)
func (e *RobogoError) WithAction(action string) *RobogoError {
	e.Action = action
	return e
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]interface{}) *RobogoError {
	return NewErrorBuilder(ErrorTypeValidation, message).
		WithDetails(details).
		Build()
}

// NewExecutionError creates a new execution error
func NewExecutionError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeExecution, message).
		WithCause(cause).
		WithAction(action).
		Build()
}

// NewAssertionError creates a new assertion error
func NewAssertionError(message string, actual, expected interface{}, operator string) *RobogoError {
	return NewErrorBuilder(ErrorTypeAssertion, message).
		WithDetails(map[string]interface{}{
			"actual":   actual,
			"expected": expected,
			"operator": operator,
		}).
		Build()
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(message string, field string, value interface{}) *RobogoError {
	return NewErrorBuilder(ErrorTypeConfiguration, message).
		WithDetails(map[string]interface{}{
			"field": field,
			"value": value,
		}).
		Build()
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeNetwork, message).
		WithCause(cause).
		WithAction(action).
		Build()
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeDatabase, message).
		WithCause(cause).
		WithAction(action).
		Build()
}

// NewMessagingError creates a new messaging error
func NewMessagingError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeMessaging, message).
		WithCause(cause).
		WithAction(action).
		Build()
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeTimeout, message).
		WithCause(cause).
		WithAction(action).
		Build()
}

// NewTemplateError creates a new template error
func NewTemplateError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeTemplate, message).
		WithCause(cause).
		WithAction(action).
		Build()
}

// NewSecurityError creates a new security error
func NewSecurityError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeSecurity, message).
		WithCause(cause).
		WithAction(action).
		Build()
}

// NewFileSystemError creates a new filesystem error
func NewFileSystemError(message string, cause error, action string) *RobogoError {
	return NewErrorBuilder(ErrorTypeFileSystem, message).
		WithCause(cause).
		WithAction(action).
		Build()
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

// FormatRobogoError formats errors for consistent reporting
func FormatRobogoError(err error) string {
	if err == nil {
		return ""
	}
	if re, ok := err.(*RobogoError); ok {
		// Customize as needed: include type, message, action, step
		msg := fmt.Sprintf("[%s] %s", re.Type, re.Message)
		if re.Action != "" {
			msg += fmt.Sprintf(" | action: %s", re.Action)
		}
		if re.Step != "" {
			msg += fmt.Sprintf(" | step: %s", re.Step)
		}
		if re.Cause != nil {
			msg += fmt.Sprintf(" | cause: %s", re.Cause.Error())
		}
		return msg
	}
	return err.Error()
}

// Common error builders for frequent scenarios

// NewArgumentCountError creates a standardized argument count error
func NewArgumentCountError(action string, expected, actual int) *RobogoError {
	return NewErrorBuilder(ErrorTypeValidation, fmt.Sprintf(MsgArgumentCount, expected, actual)).
		WithAction(action).
		WithDetails(map[string]interface{}{
			"expected_count": expected,
			"actual_count":   actual,
		}).
		Build()
}

// NewArgumentTypeError creates a standardized argument type error
func NewArgumentTypeError(action string, argIndex int, expectedType string, actualValue interface{}) *RobogoError {
	return NewErrorBuilder(ErrorTypeValidation, fmt.Sprintf(MsgArgumentType, argIndex, expectedType, fmt.Sprintf("%T", actualValue))).
		WithAction(action).
		WithDetails(map[string]interface{}{
			"argument_index": argIndex,
			"expected_type":  expectedType,
			"actual_type":    fmt.Sprintf("%T", actualValue),
			"actual_value":   actualValue,
		}).
		Build()
}

// NewArgumentValueError creates a standardized argument value error
func NewArgumentValueError(action string, argIndex int, value interface{}, reason string) *RobogoError {
	return NewErrorBuilder(ErrorTypeValidation, fmt.Sprintf(MsgArgumentValue, argIndex, value)).
		WithAction(action).
		WithDetails(map[string]interface{}{
			"argument_index": argIndex,
			"value":          value,
			"reason":         reason,
		}).
		Build()
}

// NewConnectionError creates a standardized connection error
func NewConnectionError(action string, endpoint string, cause error) *RobogoError {
	return NewErrorBuilder(ErrorTypeNetwork, fmt.Sprintf(MsgConnectionFailed, endpoint)).
		WithAction(action).
		WithCause(cause).
		WithDetails(map[string]interface{}{
			"endpoint": endpoint,
		}).
		Build()
}

// NewOperationError creates a standardized operation error
func NewOperationError(action string, operation string, cause error) *RobogoError {
	return NewErrorBuilder(ErrorTypeExecution, fmt.Sprintf(MsgOperationFailed, operation)).
		WithAction(action).
		WithCause(cause).
		WithDetails(map[string]interface{}{
			"operation": operation,
		}).
		Build()
}

// NewResourceNotFoundError creates a standardized resource not found error
func NewResourceNotFoundError(action string, resource string) *RobogoError {
	return NewErrorBuilder(ErrorTypeValidation, fmt.Sprintf(MsgResourceNotFound, resource)).
		WithAction(action).
		WithDetails(map[string]interface{}{
			"resource": resource,
		}).
		Build()
}

// NewConfigurationMissingError creates a standardized configuration missing error
func NewConfigurationMissingError(action string, configKey string) *RobogoError {
	return NewErrorBuilder(ErrorTypeConfiguration, fmt.Sprintf(MsgConfigurationMissing, configKey)).
		WithAction(action).
		WithDetails(map[string]interface{}{
			"config_key": configKey,
		}).
		Build()
}

// NewTimeoutExceededError creates a standardized timeout error
func NewTimeoutExceededError(action string, duration time.Duration, cause error) *RobogoError {
	return NewErrorBuilder(ErrorTypeTimeout, fmt.Sprintf(MsgTimeoutExceeded, duration)).
		WithAction(action).
		WithCause(cause).
		WithDetails(map[string]interface{}{
			"timeout_duration": duration,
		}).
		Build()
}

// WrapError wraps an existing error as a RobogoError
func WrapError(err error, errorType ErrorType, action string) *RobogoError {
	if err == nil {
		return nil
	}
	
	// If it's already a RobogoError, enhance it
	if roboErr := GetRobogoError(err); roboErr != nil {
		if roboErr.Action == "" {
			roboErr.Action = action
		}
		return roboErr
	}
	
	// Create a new RobogoError wrapping the original
	return NewErrorBuilder(errorType, err.Error()).
		WithAction(action).
		WithCause(err).
		Build()
}
