package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/templates"
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
	Category    ErrorCategory  `json:"category"`
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	Context     map[string]any `json:"context,omitempty"`
	Suggestions []string       `json:"suggestions,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
}

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
	Category    FailureCategory `json:"category"`
	Code        string          `json:"code"`
	Message     string          `json:"message"`
	Context     map[string]any  `json:"context,omitempty"`
	Suggestions []string        `json:"suggestions,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`

	// Failure-specific fields
	Expected   any    `json:"expected,omitempty"`
	Actual     any    `json:"actual,omitempty"`
	Comparison string `json:"comparison,omitempty"`
}

// ErrorBuilder provides a fluent interface for building structured errors
type ErrorBuilder struct {
	category    ErrorCategory
	code        string
	template    string
	context     map[string]any
	suggestions []string
	formatter   *SafeFormatter
}

// NewErrorBuilder creates a new ErrorBuilder with the specified category and code
func NewErrorBuilder(category ErrorCategory, code string) *ErrorBuilder {
	return &ErrorBuilder{
		category:  category,
		code:      code,
		context:   make(map[string]any),
		formatter: GetDefaultSafeFormatter(),
	}
}

// WithTemplate sets the error message template
func (eb *ErrorBuilder) WithTemplate(template string) *ErrorBuilder {
	eb.template = template
	return eb
}

// WithContext adds context information to the error
func (eb *ErrorBuilder) WithContext(key string, value any) *ErrorBuilder {
	eb.context[key] = value
	return eb
}

// WithSuggestion adds a suggestion for resolving the error
func (eb *ErrorBuilder) WithSuggestion(suggestion string) *ErrorBuilder {
	eb.suggestions = append(eb.suggestions, suggestion)
	return eb
}

// WithExpected adds expected value context (useful for failures)
func (eb *ErrorBuilder) WithExpected(expected any) *ErrorBuilder {
	eb.context["expected"] = expected
	return eb
}

// WithActual adds actual value context (useful for failures)
func (eb *ErrorBuilder) WithActual(actual any) *ErrorBuilder {
	eb.context["actual"] = actual
	return eb
}

// WithComparison adds comparison context (useful for failures)
func (eb *ErrorBuilder) WithComparison(comparison string) *ErrorBuilder {
	eb.context["comparison"] = comparison
	return eb
}

// Build creates an ActionResult with the structured error information
func (eb *ErrorBuilder) Build(args ...any) ActionResult {
	return eb.BuildWithStatus(ActionStatusError, args...)
}

// BuildWithStatus creates an ActionResult with the specified status
func (eb *ErrorBuilder) BuildWithStatus(status ActionStatus, args ...any) ActionResult {
	message, err := eb.formatter.Format(eb.template, args...)
	if err != nil {
		// Fallback to a safe error message if formatting fails
		message = fmt.Sprintf("Error formatting failed: %s (template: %s)", err.Error(), eb.template)
	}

	errorInfo := &ErrorInfo{
		Category:    eb.category,
		Code:        eb.code,
		Message:     message,
		Context:     eb.context,
		Suggestions: eb.suggestions,
		Timestamp:   time.Now(),
	}

	return ActionResult{
		Status:    status,
		ErrorInfo: errorInfo,
	}
}

// BuildError creates an ActionResult with ERROR status (technical errors)
func (eb *ErrorBuilder) BuildError(args ...any) ActionResult {
	return eb.BuildWithStatus(ActionStatusError, args...)
}

// BuildFailure creates an ActionResult with FAILED status (logical failures)
func (eb *ErrorBuilder) BuildFailure(args ...any) ActionResult {
	return eb.BuildWithStatus(ActionStatusFailed, args...)
}

// SafeFormatter provides secure string formatting with template validation
type SafeFormatter struct {
	templates map[string]string
}

// NewSafeFormatter creates a new SafeFormatter instance
func NewSafeFormatter() *SafeFormatter {
	return &SafeFormatter{
		templates: make(map[string]string),
	}
}

// RegisterTemplate registers a named template for safe formatting
func (sf *SafeFormatter) RegisterTemplate(name, template string) {
	sf.templates[name] = template
}

// Format safely formats a template string with the provided arguments
func (sf *SafeFormatter) Format(template string, args ...any) (string, error) {
	// Validate the template for basic safety
	if err := sf.ValidateTemplate(template); err != nil {
		return "", fmt.Errorf("template validation failed: %w", err)
	}

	// Use fmt.Sprintf for formatting, but with validation
	return fmt.Sprintf(template, args...), nil
}

// ValidateTemplate performs comprehensive validation on format templates
func (sf *SafeFormatter) ValidateTemplate(template string) error {
	// Check for potentially dangerous format specifiers
	dangerousSpecs := []string{"%n", "%*", "%#"}
	for _, spec := range dangerousSpecs {
		if strings.Contains(template, spec) {
			return fmt.Errorf("dangerous format specifier '%s' not allowed", spec)
		}
	}

	// Check for width and precision specifiers that could be exploited
	if strings.Contains(template, "%*") {
		return fmt.Errorf("width specifier '*' not allowed for security reasons")
	}

	// Validate that format verbs are properly formed
	i := 0
	for i < len(template) {
		if template[i] == '%' {
			if i+1 >= len(template) {
				return fmt.Errorf("incomplete format verb at end of template")
			}
			if template[i+1] == '%' {
				// Escaped percent, skip both characters
				i += 2
				continue
			}
			// Find the end of the format verb
			j := i + 1
			for j < len(template) && (template[j] == '-' || template[j] == '+' || template[j] == '#' || template[j] == ' ' || template[j] == '0') {
				j++
			}
			// Skip width
			for j < len(template) && template[j] >= '0' && template[j] <= '9' {
				j++
			}
			// Skip precision
			if j < len(template) && template[j] == '.' {
				j++
				for j < len(template) && template[j] >= '0' && template[j] <= '9' {
					j++
				}
			}
			// Check verb character
			if j >= len(template) {
				return fmt.Errorf("incomplete format verb starting at position %d", i)
			}
			verb := template[j]
			allowedVerbs := "vTtbcdoOxXeEfFgGsqp"
			if !strings.ContainsRune(allowedVerbs, rune(verb)) {
				return fmt.Errorf("unsupported format verb '%%%c' at position %d", verb, i)
			}
			i = j + 1
		} else {
			i++
		}
	}

	return nil
}

// GetTemplate retrieves a registered template by name
func (sf *SafeFormatter) GetTemplate(name string) (string, bool) {
	template, exists := sf.templates[name]
	return template, exists
}

// Default formatter instance
var defaultFormatter *SafeFormatter

// FailureBuilder provides a fluent interface for building structured failures
type FailureBuilder struct {
	category    FailureCategory
	code        string
	template    string
	context     map[string]any
	suggestions []string
	expected    any
	actual      any
	comparison  string
	formatter   *SafeFormatter
}

// NewFailureBuilder creates a new FailureBuilder with the specified category and code
func NewFailureBuilder(category FailureCategory, code string) *FailureBuilder {
	return &FailureBuilder{
		category:  category,
		code:      code,
		context:   make(map[string]any),
		formatter: GetDefaultSafeFormatter(),
	}
}

// WithTemplate sets the failure message template
func (fb *FailureBuilder) WithTemplate(template string) *FailureBuilder {
	fb.template = template
	return fb
}

// WithContext adds context information to the failure
func (fb *FailureBuilder) WithContext(key string, value any) *FailureBuilder {
	fb.context[key] = value
	return fb
}

// WithSuggestion adds a suggestion for resolving the failure
func (fb *FailureBuilder) WithSuggestion(suggestion string) *FailureBuilder {
	fb.suggestions = append(fb.suggestions, suggestion)
	return fb
}

// WithExpected adds expected value for the failure
func (fb *FailureBuilder) WithExpected(expected any) *FailureBuilder {
	fb.expected = expected
	return fb
}

// WithActual adds actual value for the failure
func (fb *FailureBuilder) WithActual(actual any) *FailureBuilder {
	fb.actual = actual
	return fb
}

// WithComparison adds comparison context for the failure
func (fb *FailureBuilder) WithComparison(comparison string) *FailureBuilder {
	fb.comparison = comparison
	return fb
}

// Build creates an ActionResult with FAILED status and FailureInfo
func (fb *FailureBuilder) Build(args ...any) ActionResult {
	message, err := fb.formatter.Format(fb.template, args...)
	if err != nil {
		// Fallback to a safe error message if formatting fails
		message = fmt.Sprintf("Failure formatting failed: %s (template: %s)", err.Error(), fb.template)
	}

	failureInfo := &FailureInfo{
		Category:    fb.category,
		Code:        fb.code,
		Message:     message,
		Context:     fb.context,
		Suggestions: fb.suggestions,
		Expected:    fb.expected,
		Actual:      fb.actual,
		Comparison:  fb.comparison,
		Timestamp:   time.Now(),
	}

	return ActionResult{
		Status:      ActionStatusFailed,
		FailureInfo: failureInfo,
	}
}

// BuildFailure is an alias for Build for consistency with ErrorBuilder
func (fb *FailureBuilder) BuildFailure(args ...any) ActionResult {
	return fb.Build(args...)
}

// GetDefaultSafeFormatter returns the default SafeFormatter instance
func GetDefaultSafeFormatter() *SafeFormatter {
	if defaultFormatter == nil {
		defaultFormatter = NewSafeFormatter()
		initializeDefaultTemplates(defaultFormatter)
	}
	return defaultFormatter
}

// initializeDefaultTemplates sets up commonly used error templates
func initializeDefaultTemplates(formatter *SafeFormatter) {
	// Import all error templates from constants package
	errorTemplates := templates.InitializeErrorTemplates()

	for name, template := range errorTemplates {
		formatter.RegisterTemplate(name, template)
	}
}
