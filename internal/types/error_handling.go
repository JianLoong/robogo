package types

import (
	"fmt"
	"strings"
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

// Build creates an ActionResult with the structured error information
func (eb *ErrorBuilder) Build(args ...any) ActionResult {
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
		Status:    ActionStatusError,
		ErrorInfo: errorInfo,
	}
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
	templates := map[string]string{
		"assertion.failed":           "Assertion failed: expected %v %s %v, but got %v",
		"assertion.type_mismatch":    "Type mismatch in assertion: expected %s but got %s",
		"variable.unresolved":        "Variable '%s' could not be resolved in template '%s'",
		"variable.invalid_access":    "Invalid access path '%s' in variable '%s'",
		"action.invalid_args":        "Action '%s' received invalid arguments: %s",
		"action.execution_failed":    "Action '%s' execution failed: %s",
		"http.request_failed":        "HTTP request failed: %s %s returned %d",
		"database.connection_failed": "Database connection failed: %s",
		"validation.required_field":  "Required field '%s' is missing or empty",
		"validation.invalid_format":  "Field '%s' has invalid format: %s",
		"system.file_not_found":      "File not found: %s",
		"system.permission_denied":   "Permission denied: %s",
	}

	for name, template := range templates {
		formatter.RegisterTemplate(name, template)
	}
}
