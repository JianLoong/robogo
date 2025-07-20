package types

import (
	"fmt"
	"time"
)

// ErrorCategory represents different categories of errors that can occur
type ErrorCategory string

const (
	ErrorCategoryValidation ErrorCategory = "validation"
	ErrorCategoryExecution  ErrorCategory = "execution"
	ErrorCategoryAssertion  ErrorCategory = "assertion"
	ErrorCategoryVariable   ErrorCategory = "variable"
	ErrorCategoryNetwork    ErrorCategory = "network"
	ErrorCategoryDatabase   ErrorCategory = "database"
	ErrorCategorySystem     ErrorCategory = "system"
)

// ErrorInfo contains structured information about an error
type ErrorInfo struct {
	Category  ErrorCategory `json:"category"`
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewError creates a simple error result
func NewError(category ErrorCategory, code, message string) ActionResult {
	return ActionResult{
		Status: ActionStatusError,
		ErrorInfo: &ErrorInfo{
			Category:  category,
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
		},
	}
}

// Backward compatibility builders - simple wrappers

// ErrorBuilder provides rich error construction
type ErrorBuilder struct {
	category    ErrorCategory
	code        string
	template    string
	context     map[string]any
	suggestions []string
	expected    any
	actual      any
	comparison  string
}

// NewErrorBuilder creates a new ErrorBuilder
func NewErrorBuilder(category ErrorCategory, code string) *ErrorBuilder {
	return &ErrorBuilder{
		category: category,
		code:     code,
		context:  make(map[string]any),
	}
}

// WithTemplate sets the error message template
func (eb *ErrorBuilder) WithTemplate(template string) *ErrorBuilder {
	eb.template = template
	return eb
}

// WithContext adds contextual information to the error
func (eb *ErrorBuilder) WithContext(key string, value any) *ErrorBuilder {
	if eb.context == nil {
		eb.context = make(map[string]any)
	}
	eb.context[key] = value
	return eb
}

// WithSuggestion adds a suggestion for fixing the error
func (eb *ErrorBuilder) WithSuggestion(suggestion string) *ErrorBuilder {
	eb.suggestions = append(eb.suggestions, suggestion)
	return eb
}

// WithExpected sets the expected value for comparison errors
func (eb *ErrorBuilder) WithExpected(expected any) *ErrorBuilder {
	eb.expected = expected
	return eb
}

// WithActual sets the actual value for comparison errors
func (eb *ErrorBuilder) WithActual(actual any) *ErrorBuilder {
	eb.actual = actual
	return eb
}

// WithComparison sets the comparison operator for assertion errors
func (eb *ErrorBuilder) WithComparison(comparison string) *ErrorBuilder {
	eb.comparison = comparison
	return eb
}

// Build creates the final error result with rich context
func (eb *ErrorBuilder) Build(args ...any) ActionResult {
	// Start with the template
	message := eb.template

	// Apply template formatting if args provided
	if len(args) > 0 && eb.template != "" {
		message = fmt.Sprintf(eb.template, args...)
	}

	// Enhance message with context if available
	if len(eb.context) > 0 {
		message += "\nContext:"
		for key, value := range eb.context {
			message += fmt.Sprintf("\n  %s: %v", key, value)
		}
	}

	// Add comparison details for assertion errors
	if eb.expected != nil || eb.actual != nil {
		message += "\nComparison Details:"
		if eb.expected != nil {
			message += fmt.Sprintf("\n  Expected: %v", eb.expected)
		}
		if eb.actual != nil {
			message += fmt.Sprintf("\n  Actual: %v", eb.actual)
		}
		if eb.comparison != "" {
			message += fmt.Sprintf("\n  Operator: %s", eb.comparison)
		}
	}

	// Add suggestions if available
	if len(eb.suggestions) > 0 {
		message += "\nSuggestions:"
		for _, suggestion := range eb.suggestions {
			message += fmt.Sprintf("\n  • %s", suggestion)
		}
	}

	return NewError(eb.category, eb.code, message)
}
