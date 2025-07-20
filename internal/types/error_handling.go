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

// FailureCategory represents different categories of logical failures
type FailureCategory string

const (
	FailureCategoryAssertion  FailureCategory = "assertion"
	FailureCategoryValidation FailureCategory = "validation"
	FailureCategoryBusiness   FailureCategory = "business_rule"
	FailureCategoryData       FailureCategory = "data_mismatch"
	FailureCategoryResponse   FailureCategory = "response_validation"
)

// ErrorInfo contains structured information about an error
type ErrorInfo struct {
	Category  ErrorCategory `json:"category"`
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Timestamp time.Time     `json:"timestamp"`
}

// FailureInfo contains structured information about a logical failure
type FailureInfo struct {
	Category  FailureCategory `json:"category"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Timestamp time.Time       `json:"timestamp"`
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

// NewFailure creates a simple failure result
func NewFailure(category FailureCategory, code, message string) ActionResult {
	return ActionResult{
		Status: ActionStatusFailed,
		FailureInfo: &FailureInfo{
			Category:  category,
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
		},
	}
}

// Backward compatibility builders - simple wrappers

// ErrorBuilder provides backward compatibility
type ErrorBuilder struct {
	category ErrorCategory
	code     string
	message  string
}

// FailureBuilder provides backward compatibility  
type FailureBuilder struct {
	category FailureCategory
	code     string
	message  string
}

// NewErrorBuilder creates a new ErrorBuilder (simplified)
func NewErrorBuilder(category ErrorCategory, code string) *ErrorBuilder {
	return &ErrorBuilder{
		category: category,
		code:     code,
	}
}

// NewFailureBuilder creates a new FailureBuilder (simplified)
func NewFailureBuilder(category FailureCategory, code string) *FailureBuilder {
	return &FailureBuilder{
		category: category,
		code:     code,
	}
}

// WithTemplate sets the message (ignores template complexity)
func (eb *ErrorBuilder) WithTemplate(template string) *ErrorBuilder {
	eb.message = template
	return eb
}

// WithContext ignored for simplicity
func (eb *ErrorBuilder) WithContext(key string, value any) *ErrorBuilder {
	return eb
}

// WithSuggestion ignored for simplicity
func (eb *ErrorBuilder) WithSuggestion(suggestion string) *ErrorBuilder {
	return eb
}

// WithExpected ignored for simplicity
func (eb *ErrorBuilder) WithExpected(expected any) *ErrorBuilder {
	return eb
}

// WithActual ignored for simplicity
func (eb *ErrorBuilder) WithActual(actual any) *ErrorBuilder {
	return eb
}

// WithComparison ignored for simplicity
func (eb *ErrorBuilder) WithComparison(comparison string) *ErrorBuilder {
	return eb
}

// Build creates the final error result with formatted message
func (eb *ErrorBuilder) Build(args ...any) ActionResult {
	message := eb.message
	if len(args) > 0 {
		message = fmt.Sprintf(eb.message, args...)
	}
	return NewError(eb.category, eb.code, message)
}

// FailureBuilder methods (similar pattern)

func (fb *FailureBuilder) WithTemplate(template string) *FailureBuilder {
	fb.message = template
	return fb
}

func (fb *FailureBuilder) WithExpected(expected any) *FailureBuilder {
	return fb
}

func (fb *FailureBuilder) WithActual(actual any) *FailureBuilder {
	return fb
}

func (fb *FailureBuilder) WithComparison(comparison string) *FailureBuilder {
	return fb
}

func (fb *FailureBuilder) WithContext(key string, value any) *FailureBuilder {
	return fb
}

func (fb *FailureBuilder) Build(args ...any) ActionResult {
	message := fb.message
	if len(args) > 0 {
		message = fmt.Sprintf(fb.message, args...)
	}
	return NewFailure(fb.category, fb.code, message)
}