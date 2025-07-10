package util

import (
	"fmt"
	"runtime"
	"strings"
	"time"
	"github.com/google/uuid"
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

// StackFrame represents a single frame in a stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

// ErrorContext provides contextual information about where and when an error occurred
type ErrorContext struct {
	CorrelationID string                 `json:"correlation_id"`
	Source        *StackFrame            `json:"source,omitempty"`
	StackTrace    []StackFrame           `json:"stack_trace,omitempty"`
	Breadcrumbs   []string               `json:"breadcrumbs,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Environment   map[string]string      `json:"environment,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	TestRun       string                 `json:"test_run,omitempty"`
}

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
	Context      *ErrorContext          `json:"context,omitempty"`
}

// ErrorBuilder provides a fluent interface for building RobogoError instances
type ErrorBuilder struct {
	error      *RobogoError
	captureStack bool
	skipFrames   int
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
			Context: &ErrorContext{
				CorrelationID: generateCorrelationID(),
				Breadcrumbs:   make([]string, 0),
				Variables:     make(map[string]interface{}),
				Environment:   make(map[string]string),
			},
		},
		captureStack: true,
		skipFrames:   1, // Skip NewErrorBuilder frame
	}
}

// NewErrorBuilderWithoutStack creates an error builder without stack trace capture
func NewErrorBuilderWithoutStack(errorType ErrorType, message string) *ErrorBuilder {
	builder := NewErrorBuilder(errorType, message)
	builder.captureStack = false
	return builder
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

// WithContext adds error context information
func (b *ErrorBuilder) WithContext(context *ErrorContext) *ErrorBuilder {
	if context != nil {
		b.error.Context = context
	}
	return b
}

// WithCorrelationID sets a correlation ID for error tracking
func (b *ErrorBuilder) WithCorrelationID(correlationID string) *ErrorBuilder {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	b.error.Context.CorrelationID = correlationID
	return b
}

// WithBreadcrumbs adds breadcrumb trail to the error
func (b *ErrorBuilder) WithBreadcrumbs(breadcrumbs []string) *ErrorBuilder {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	b.error.Context.Breadcrumbs = breadcrumbs
	return b
}

// AddBreadcrumb adds a single breadcrumb to the trail
func (b *ErrorBuilder) AddBreadcrumb(breadcrumb string) *ErrorBuilder {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	if b.error.Context.Breadcrumbs == nil {
		b.error.Context.Breadcrumbs = make([]string, 0)
	}
	b.error.Context.Breadcrumbs = append(b.error.Context.Breadcrumbs, breadcrumb)
	return b
}

// WithVariables adds variable context to the error
func (b *ErrorBuilder) WithVariables(variables map[string]interface{}) *ErrorBuilder {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	if b.error.Context.Variables == nil {
		b.error.Context.Variables = make(map[string]interface{})
	}
	for k, v := range variables {
		b.error.Context.Variables[k] = v
	}
	return b
}

// WithEnvironment adds environment context to the error
func (b *ErrorBuilder) WithEnvironment(env map[string]string) *ErrorBuilder {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	if b.error.Context.Environment == nil {
		b.error.Context.Environment = make(map[string]string)
	}
	for k, v := range env {
		b.error.Context.Environment[k] = v
	}
	return b
}

// WithRequestID sets a request ID for error tracking
func (b *ErrorBuilder) WithRequestID(requestID string) *ErrorBuilder {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	b.error.Context.RequestID = requestID
	return b
}

// WithTestRun sets a test run identifier
func (b *ErrorBuilder) WithTestRun(testRun string) *ErrorBuilder {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	b.error.Context.TestRun = testRun
	return b
}

// CaptureStack enables or disables stack trace capture
func (b *ErrorBuilder) CaptureStack(capture bool) *ErrorBuilder {
	b.captureStack = capture
	return b
}

// SkipFrames sets the number of stack frames to skip
func (b *ErrorBuilder) SkipFrames(skip int) *ErrorBuilder {
	b.skipFrames = skip
	return b
}

// Build creates the final RobogoError instance
func (b *ErrorBuilder) Build() *RobogoError {
	// Capture stack trace if enabled
	if b.captureStack {
		b.captureStackTrace()
	}
	return b.error
}

// captureStackTrace captures the current stack trace
func (b *ErrorBuilder) captureStackTrace() {
	if b.error.Context == nil {
		b.error.Context = &ErrorContext{}
	}
	
	// Capture source location (current frame)
	if pc, file, line, ok := runtime.Caller(b.skipFrames + 1); ok {
		funcName := runtime.FuncForPC(pc).Name()
		pkgName := extractPackageName(funcName)
		
		b.error.Context.Source = &StackFrame{
			Function: funcName,
			File:     file,
			Line:     line,
			Package:  pkgName,
		}
	}
	
	// Capture full stack trace
	b.error.Context.StackTrace = captureFullStackTrace(b.skipFrames + 1)
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
	
	// Add correlation ID for error tracking
	if e.Context != nil && e.Context.CorrelationID != "" {
		parts = append(parts, fmt.Sprintf("correlation_id=%s", e.Context.CorrelationID))
	}

	return strings.Join(parts, " | ")
}

// ErrorWithStackTrace returns the error message with stack trace
func (e *RobogoError) ErrorWithStackTrace() string {
	msg := e.Error()
	if e.Context != nil && len(e.Context.StackTrace) > 0 {
		msg += "\n\nStack Trace:\n"
		for i, frame := range e.Context.StackTrace {
			msg += fmt.Sprintf("  %d. %s\n", i+1, frame.String())
			if i >= 10 { // Limit to first 10 frames
				msg += "     ... (truncated)\n"
				break
			}
		}
	}
	return msg
}

// ErrorWithContext returns the error message with full context
func (e *RobogoError) ErrorWithContext() string {
	msg := e.Error()
	if e.Context == nil {
		return msg
	}
	
	msg += "\n\nError Context:"
	if e.Context.CorrelationID != "" {
		msg += fmt.Sprintf("\n  Correlation ID: %s", e.Context.CorrelationID)
	}
	if e.Context.RequestID != "" {
		msg += fmt.Sprintf("\n  Request ID: %s", e.Context.RequestID)
	}
	if e.Context.TestRun != "" {
		msg += fmt.Sprintf("\n  Test Run: %s", e.Context.TestRun)
	}
	
	if e.Context.Source != nil {
		msg += fmt.Sprintf("\n  Source: %s", e.Context.Source.String())
	}
	
	if len(e.Context.Breadcrumbs) > 0 {
		msg += "\n  Breadcrumbs:"
		for i, crumb := range e.Context.Breadcrumbs {
			msg += fmt.Sprintf("\n    %d. %s", i+1, crumb)
		}
	}
	
	if len(e.Context.Variables) > 0 {
		msg += "\n  Variables:"
		for k, v := range e.Context.Variables {
			msg += fmt.Sprintf("\n    %s: %v", k, v)
		}
	}
	
	return msg
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
		// For debugging, prioritize the original error message
		if re.Cause != nil {
			// If we have a cause, show the original error first
			return re.Cause.Error()
		}
		// Otherwise show our enhanced message
		return re.Message
	}
	return err.Error()
}

// FormatRobogoErrorDetailed formats errors with full context for debugging
func FormatRobogoErrorDetailed(err error) string {
	if err == nil {
		return ""
	}
	if re, ok := err.(*RobogoError); ok {
		// Detailed format with type, action, step info
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

// FormatRobogoErrorForLogging formats errors for structured logging
func FormatRobogoErrorForLogging(err error) string {
	if err == nil {
		return ""
	}
	if re, ok := err.(*RobogoError); ok {
		var parts []string
		parts = append(parts, fmt.Sprintf("type=%s", re.Type))
		parts = append(parts, fmt.Sprintf("message=%s", re.Message))
		if re.Action != "" {
			parts = append(parts, fmt.Sprintf("action=%s", re.Action))
		}
		if re.Step != "" {
			parts = append(parts, fmt.Sprintf("step=%s", re.Step))
		}
		if re.Context != nil && re.Context.CorrelationID != "" {
			parts = append(parts, fmt.Sprintf("correlation_id=%s", re.Context.CorrelationID))
		}
		return strings.Join(parts, " | ")
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
		SkipFrames(1). // Skip WrapError frame
		Build()
}

// WrapErrorWithContext wraps an error with additional context
func WrapErrorWithContext(err error, errorType ErrorType, action string, context *ErrorContext) *RobogoError {
	if err == nil {
		return nil
	}
	
	builder := NewErrorBuilder(errorType, err.Error()).
		WithAction(action).
		WithCause(err).
		SkipFrames(1) // Skip WrapErrorWithContext frame
	
	if context != nil {
		builder = builder.WithContext(context)
	}
	
	return builder.Build()
}

// Helper functions for stack trace management

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	return uuid.New().String()
}

// extractPackageName extracts package name from function name
func extractPackageName(funcName string) string {
	parts := strings.Split(funcName, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return ""
}

// captureFullStackTrace captures the full stack trace
func captureFullStackTrace(skip int) []StackFrame {
	frames := make([]StackFrame, 0)
	
	for i := skip; i < skip+20; i++ { // Capture up to 20 frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		funcName := runtime.FuncForPC(pc).Name()
		pkgName := extractPackageName(funcName)
		
		frames = append(frames, StackFrame{
			Function: funcName,
			File:     file,
			Line:     line,
			Package:  pkgName,
		})
	}
	
	return frames
}

// String returns a string representation of a stack frame
func (sf StackFrame) String() string {
	return fmt.Sprintf("%s:%d in %s", sf.File, sf.Line, sf.Function)
}

// IsInternalFrame checks if the frame is from internal robogo code
func (sf StackFrame) IsInternalFrame() bool {
	return strings.Contains(sf.Package, "github.com/JianLoong/robogo")
}

// Context management functions

// NewErrorContext creates a new error context
func NewErrorContext() *ErrorContext {
	return &ErrorContext{
		CorrelationID: generateCorrelationID(),
		Breadcrumbs:   make([]string, 0),
		Variables:     make(map[string]interface{}),
		Environment:   make(map[string]string),
	}
}

// WithCorrelationID sets the correlation ID
func (ec *ErrorContext) WithCorrelationID(id string) *ErrorContext {
	ec.CorrelationID = id
	return ec
}

// AddBreadcrumb adds a breadcrumb to the context
func (ec *ErrorContext) AddBreadcrumb(breadcrumb string) *ErrorContext {
	ec.Breadcrumbs = append(ec.Breadcrumbs, breadcrumb)
	return ec
}

// SetVariable sets a variable in the context
func (ec *ErrorContext) SetVariable(key string, value interface{}) *ErrorContext {
	ec.Variables[key] = value
	return ec
}

// SetEnvironment sets an environment variable
func (ec *ErrorContext) SetEnvironment(key, value string) *ErrorContext {
	ec.Environment[key] = value
	return ec
}

// Enhanced error reporting functions

// PrintDetailedError prints a detailed error report
func PrintDetailedError(err error) {
	if roboErr, ok := err.(*RobogoError); ok {
		fmt.Printf("Error Report:\n%s\n", roboErr.ErrorWithContext())
		if len(roboErr.Context.StackTrace) > 0 {
			fmt.Printf("\nStack Trace:\n")
			for i, frame := range roboErr.Context.StackTrace {
				if i > 10 {
					fmt.Printf("  ... (truncated)\n")
					break
				}
				fmt.Printf("  %s\n", frame.String())
			}
		}
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}

// GetErrorSummary returns a summary of an error for logging
func GetErrorSummary(err error) map[string]interface{} {
	summary := make(map[string]interface{})
	
	if roboErr, ok := err.(*RobogoError); ok {
		summary["type"] = roboErr.Type
		summary["message"] = roboErr.Message
		summary["severity"] = roboErr.Severity
		summary["retryable"] = roboErr.Retryable
		summary["recoverable"] = roboErr.Recoverable
		summary["timestamp"] = roboErr.Timestamp
		
		if roboErr.Action != "" {
			summary["action"] = roboErr.Action
		}
		if roboErr.Step != "" {
			summary["step"] = roboErr.Step
		}
		if roboErr.TestCase != "" {
			summary["test_case"] = roboErr.TestCase
		}
		
		if roboErr.Context != nil {
			summary["correlation_id"] = roboErr.Context.CorrelationID
			if roboErr.Context.Source != nil {
				summary["source"] = roboErr.Context.Source.String()
			}
		}
	} else {
		summary["message"] = err.Error()
		summary["type"] = "unknown"
	}
	
	return summary
}
