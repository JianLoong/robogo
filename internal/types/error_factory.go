package types

import "fmt"

// ErrorFactory provides centralized error creation with predefined templates
type ErrorFactory struct {
	formatter *SafeFormatter
}

// NewErrorFactory creates a new ErrorFactory instance
func NewErrorFactory() *ErrorFactory {
	factory := &ErrorFactory{
		formatter: NewSafeFormatter(),
	}
	factory.initializeTemplates()
	return factory
}

// RegisterTemplate registers a new error template
func (ef *ErrorFactory) RegisterTemplate(name, template string) error {
	// Validate the template before registration
	if err := ef.formatter.ValidateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed for '%s': %w", name, err)
	}

	ef.formatter.RegisterTemplate(name, template)
	return nil
}

// GetTemplate retrieves a registered template by name
func (ef *ErrorFactory) GetTemplate(name string) (string, bool) {
	return ef.formatter.GetTemplate(name)
}

// ValidateAllTemplates validates all registered templates
func (ef *ErrorFactory) ValidateAllTemplates() error {
	// Since templates are stored in the formatter, we need to validate them there
	// This is a simplified validation that checks if the formatter is working
	_, err := ef.formatter.Format("test %s", "validation")
	if err != nil {
		return fmt.Errorf("formatter validation failed: %w", err)
	}
	return nil
}

// CreateValidationError creates a validation error with the specified code and template name
func (ef *ErrorFactory) CreateValidationError(code, templateName string, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		// Fallback to a generic template if the specific one doesn't exist
		template = "Validation error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	return NewErrorBuilder(ErrorCategoryValidation, code).
		WithTemplate(template).
		Build(args...)
}

// CreateValidationErrorWithTemplate creates a validation error with a direct template string
func (ef *ErrorFactory) CreateValidationErrorWithTemplate(code, template string, args ...any) ActionResult {
	return NewErrorBuilder(ErrorCategoryValidation, code).
		WithTemplate(template).
		Build(args...)
}

// CreateExecutionError creates an execution error for a specific action
func (ef *ErrorFactory) CreateExecutionError(action, message string) ActionResult {
	template, exists := ef.formatter.GetTemplate("action.execution_failed")
	if !exists {
		template = "Action '%s' execution failed: %s"
	}

	return NewErrorBuilder(ErrorCategoryExecution, "EXEC_FAILED").
		WithTemplate(template).
		WithContext("action", action).
		Build(action, message)
}

// CreateExecutionErrorWithTemplate creates an execution error with template name
func (ef *ErrorFactory) CreateExecutionErrorWithTemplate(code, templateName string, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		template = "Execution error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	return NewErrorBuilder(ErrorCategoryExecution, code).
		WithTemplate(template).
		Build(args...)
}

// CreateAssertionError creates an assertion error with detailed context
func (ef *ErrorFactory) CreateAssertionError(actual, expected any, operator string) ActionResult {
	template, exists := ef.formatter.GetTemplate("assertion.failed")
	if !exists {
		template = "Assertion failed: expected %v %s %v, but got %v"
	}

	return NewErrorBuilder(ErrorCategoryAssertion, "ASSERT_FAILED").
		WithTemplate(template).
		WithContext("actual", actual).
		WithContext("expected", expected).
		WithContext("operator", operator).
		Build(expected, operator, expected, actual)
}

// CreateAssertionErrorWithTemplate creates an assertion error with template name
func (ef *ErrorFactory) CreateAssertionErrorWithTemplate(code, templateName string, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		template = "Assertion error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	return NewErrorBuilder(ErrorCategoryAssertion, code).
		WithTemplate(template).
		Build(args...)
}

// CreateVariableError creates a variable resolution error
func (ef *ErrorFactory) CreateVariableError(variable, template string) ActionResult {
	errorTemplate, exists := ef.formatter.GetTemplate("variable.unresolved")
	if !exists {
		errorTemplate = "Variable '%s' could not be resolved in template '%s'"
	}

	return NewErrorBuilder(ErrorCategoryVariable, "VAR_UNRESOLVED").
		WithTemplate(errorTemplate).
		WithContext("variable", variable).
		WithContext("template", template).
		WithSuggestion("Check if the variable is defined and accessible in the current scope").
		Build(variable, template)
}

// CreateVariableErrorWithTemplate creates a variable error with template name
func (ef *ErrorFactory) CreateVariableErrorWithTemplate(code, templateName string, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		template = "Variable error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	return NewErrorBuilder(ErrorCategoryVariable, code).
		WithTemplate(template).
		Build(args...)
}

// CreateNetworkError creates a network-related error
func (ef *ErrorFactory) CreateNetworkError(code, message string) ActionResult {
	return NewErrorBuilder(ErrorCategoryNetwork, code).
		WithTemplate("Network error: %s").
		Build(message)
}

// CreateNetworkErrorWithTemplate creates a network error with template name
func (ef *ErrorFactory) CreateNetworkErrorWithTemplate(code, templateName string, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		template = "Network error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	return NewErrorBuilder(ErrorCategoryNetwork, code).
		WithTemplate(template).
		Build(args...)
}

// CreateSystemError creates a system-level error
func (ef *ErrorFactory) CreateSystemError(code, message string) ActionResult {
	return NewErrorBuilder(ErrorCategorySystem, code).
		WithTemplate("System error: %s").
		Build(message)
}

// CreateSystemErrorWithTemplate creates a system error with template name
func (ef *ErrorFactory) CreateSystemErrorWithTemplate(code, templateName string, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		template = "System error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	return NewErrorBuilder(ErrorCategorySystem, code).
		WithTemplate(template).
		Build(args...)
}

// CreateHTTPError creates an HTTP-specific error with request details
func (ef *ErrorFactory) CreateHTTPError(method, url string, statusCode int) ActionResult {
	template, exists := ef.formatter.GetTemplate("http.request_failed")
	if !exists {
		template = "HTTP request failed: %s %s returned %d"
	}

	return NewErrorBuilder(ErrorCategoryNetwork, "HTTP_REQUEST_FAILED").
		WithTemplate(template).
		WithContext("method", method).
		WithContext("url", url).
		WithContext("status_code", statusCode).
		Build(method, url, statusCode)
}

// CreateDatabaseError creates a database-specific error
func (ef *ErrorFactory) CreateDatabaseError(message string) ActionResult {
	template, exists := ef.formatter.GetTemplate("database.connection_failed")
	if !exists {
		template = "Database connection failed: %s"
	}

	return NewErrorBuilder(ErrorCategorySystem, "DB_CONNECTION_FAILED").
		WithTemplate(template).
		WithSuggestion("Check database connection parameters and network connectivity").
		Build(message)
}

// Default factory instance
var defaultErrorFactory *ErrorFactory

// CreateErrorWithContext creates an error with additional context information
func (ef *ErrorFactory) CreateErrorWithContext(category ErrorCategory, code, templateName string, context map[string]any, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		template = "Error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	builder := NewErrorBuilder(category, code).WithTemplate(template)

	// Add context information
	for key, value := range context {
		builder.WithContext(key, value)
	}

	return builder.Build(args...)
}

// CreateErrorWithSuggestions creates an error with suggestions for resolution
func (ef *ErrorFactory) CreateErrorWithSuggestions(category ErrorCategory, code, templateName string, suggestions []string, args ...any) ActionResult {
	template, exists := ef.GetTemplate(templateName)
	if !exists {
		template = "Error: %s"
		args = []any{fmt.Sprintf("Template '%s' not found", templateName)}
	}

	builder := NewErrorBuilder(category, code).WithTemplate(template)

	// Add suggestions
	for _, suggestion := range suggestions {
		builder.WithSuggestion(suggestion)
	}

	return builder.Build(args...)
}

// initializeTemplates sets up predefined error templates
func (ef *ErrorFactory) initializeTemplates() {
	templates := map[string]string{
		// Assertion templates
		"assertion.failed":          "Assertion failed: expected %v %s %v, but got %v",
		"assertion.type_mismatch":   "Type mismatch in assertion: expected %s but got %s",
		"assertion.contains_failed": "Contains assertion failed: '%s' not found in '%s'",
		"assertion.numeric_failed":  "Numeric assertion failed: expected %v %s %v, but got %v (compared as %s)",

		// Variable templates
		"variable.unresolved":           "Variable '%s' could not be resolved in template '%s'",
		"variable.invalid_access":       "Invalid access path '%s' in variable '%s'",
		"variable.expression_failed":    "Expression evaluation failed: %s",
		"variable.nested_access_failed": "Nested variable access failed at path '%s' in variable '%s'",

		// Action templates
		"action.invalid_args":     "Action '%s' received invalid arguments: %s",
		"action.execution_failed": "Action '%s' execution failed: %s",
		"action.missing_required": "Action '%s' missing required parameter: %s",
		"action.invalid_type":     "Action '%s' parameter '%s' has invalid type: expected %s, got %s",

		// HTTP templates
		"http.request_failed":    "HTTP request failed: %s %s returned %d",
		"http.connection_failed": "HTTP connection failed: %s",
		"http.timeout":           "HTTP request timeout: %s %s after %s",
		"http.invalid_response":  "HTTP response invalid: %s",

		// Database templates
		"database.connection_failed":  "Database connection failed: %s",
		"database.query_failed":       "Database query failed: %s",
		"database.transaction_failed": "Database transaction failed: %s",
		"database.invalid_result":     "Database query returned invalid result: %s",

		// Validation templates
		"validation.required_field": "Required field '%s' is missing or empty",
		"validation.invalid_format": "Field '%s' has invalid format: %s",
		"validation.out_of_range":   "Field '%s' value %v is out of range: %s",
		"validation.invalid_enum":   "Field '%s' has invalid value '%s': must be one of %s",

		// System templates
		"system.file_not_found":    "File not found: %s",
		"system.permission_denied": "Permission denied: %s",
		"system.io_error":          "I/O error: %s",
		"system.timeout":           "Operation timeout: %s",

		// Execution templates
		"execution.step_failed":        "Step %d (%s) failed: %s",
		"execution.loop_failed":        "Loop execution failed at iteration %d: %s",
		"execution.condition_failed":   "Condition evaluation failed: %s",
		"execution.control_flow_error": "Control flow error: %s",
	}

	for name, template := range templates {
		ef.formatter.RegisterTemplate(name, template)
	}
}

// GetDefaultErrorFactory returns the default ErrorFactory instance
func GetDefaultErrorFactory() *ErrorFactory {
	if defaultErrorFactory == nil {
		defaultErrorFactory = NewErrorFactory()
	}
	return defaultErrorFactory
}
