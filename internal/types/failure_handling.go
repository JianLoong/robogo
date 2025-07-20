package types

import (
	"fmt"
	"time"
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

// FailureInfo contains structured information about a logical failure
type FailureInfo struct {
	Category  FailureCategory `json:"category"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Timestamp time.Time       `json:"timestamp"`
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

// FailureBuilder provides rich failure construction
type FailureBuilder struct {
	category    FailureCategory
	code        string
	template    string
	context     map[string]any
	suggestions []string
	expected    any
	actual      any
	comparison  string
}

// NewFailureBuilder creates a new FailureBuilder
func NewFailureBuilder(category FailureCategory, code string) *FailureBuilder {
	return &FailureBuilder{
		category: category,
		code:     code,
		context:  make(map[string]any),
	}
}

// WithTemplate sets the failure message template
func (fb *FailureBuilder) WithTemplate(template string) *FailureBuilder {
	fb.template = template
	return fb
}

// WithContext adds contextual information to the failure
func (fb *FailureBuilder) WithContext(key string, value any) *FailureBuilder {
	if fb.context == nil {
		fb.context = make(map[string]any)
	}
	fb.context[key] = value
	return fb
}

// WithSuggestion adds a suggestion for fixing the failure
func (fb *FailureBuilder) WithSuggestion(suggestion string) *FailureBuilder {
	fb.suggestions = append(fb.suggestions, suggestion)
	return fb
}

// WithExpected sets the expected value for comparison failures
func (fb *FailureBuilder) WithExpected(expected any) *FailureBuilder {
	fb.expected = expected
	return fb
}

// WithActual sets the actual value for comparison failures
func (fb *FailureBuilder) WithActual(actual any) *FailureBuilder {
	fb.actual = actual
	return fb
}

// WithComparison sets the comparison operator for assertion failures
func (fb *FailureBuilder) WithComparison(comparison string) *FailureBuilder {
	fb.comparison = comparison
	return fb
}

// Build creates the final failure result with rich context
func (fb *FailureBuilder) Build(args ...any) ActionResult {
	// Start with the template
	message := fb.template

	// Apply template formatting if args provided
	if len(args) > 0 && fb.template != "" {
		message = fmt.Sprintf(fb.template, args...)
	}

	// Enhance message with context if available
	if len(fb.context) > 0 {
		message += "\nContext:"
		for key, value := range fb.context {
			message += fmt.Sprintf("\n  %s: %v", key, value)
		}
	}

	// Add comparison details for assertion failures
	if fb.expected != nil || fb.actual != nil {
		message += "\nComparison Details:"
		if fb.expected != nil {
			message += fmt.Sprintf("\n  Expected: %v", fb.expected)
		}
		if fb.actual != nil {
			message += fmt.Sprintf("\n  Actual: %v", fb.actual)
		}
		if fb.comparison != "" {
			message += fmt.Sprintf("\n  Operator: %s", fb.comparison)
		}
	}

	// Add suggestions if available
	if len(fb.suggestions) > 0 {
		message += "\nSuggestions:"
		for _, suggestion := range fb.suggestions {
			message += fmt.Sprintf("\n  • %s", suggestion)
		}
	}

	return NewFailure(fb.category, fb.code, message)
}
