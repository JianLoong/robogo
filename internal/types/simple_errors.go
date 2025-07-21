package types

import "fmt"

// Simple error functions to reduce boilerplate in actions

// Validation errors
func MissingArgsError(action string, required, provided int) ActionResult {
	return NewErrorBuilder(ErrorCategoryValidation, "MISSING_ARGS").
		WithTemplate("%s action requires %d arguments, got %d").
		Build(action, required, provided)
}

func InvalidArgError(action, arg string, expected any) ActionResult {
	return NewErrorBuilder(ErrorCategoryValidation, "INVALID_ARG").
		WithTemplate("%s action: invalid %s, expected %v").
		Build(action, arg, expected)
}

func UnknownOperationError(action, operation string) ActionResult {
	return NewErrorBuilder(ErrorCategoryValidation, "UNKNOWN_OPERATION").
		WithTemplate("%s action: unknown operation '%s'").
		Build(action, operation)
}

// Network errors
func ConnectionError(service, details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryNetwork, "CONNECTION_FAILED").
		WithTemplate("Failed to connect to %s: %s").
		Build(service, details)
}

func RequestError(operation, details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryNetwork, "REQUEST_FAILED").
		WithTemplate("%s failed: %s").
		Build(operation, details)
}

func TimeoutError(operation string) ActionResult {
	return NewErrorBuilder(ErrorCategoryNetwork, "TIMEOUT").
		WithTemplate("%s timed out").
		Build(operation)
}

// Database errors
func DatabaseConnectionError(dbType, details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryDatabase, "DB_CONNECTION_FAILED").
		WithTemplate("%s connection failed: %s").
		Build(dbType, details)
}

func DatabaseQueryError(dbType, details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryDatabase, "DB_QUERY_FAILED").
		WithTemplate("%s query failed: %s").
		Build(dbType, details)
}

func DatabaseExecuteError(dbType, details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryDatabase, "DB_EXECUTE_FAILED").
		WithTemplate("%s execute failed: %s").
		Build(dbType, details)
}

// Execution errors
func ActionFailedError(details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryExecution, "ACTION_FAILED").
		WithTemplate("Action failed: %s").
		Build(details)
}

func ExtractionFailedError(details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryExecution, "EXTRACTION_FAILED").
		WithTemplate("Failed to extract data: %s").
		Build(details)
}

func UnsupportedExtractionTypeError(extractionType string) ActionResult {
	return NewErrorBuilder(ErrorCategoryValidation, "UNSUPPORTED_EXTRACTION_TYPE").
		WithTemplate("Unsupported extraction type: %s").
		Build(extractionType)
}

func InvalidRegexPatternError(pattern, details string) ActionResult {
	return NewErrorBuilder(ErrorCategoryValidation, "INVALID_REGEX_PATTERN").
		WithTemplate("Invalid regex pattern '%s': %s").
		Build(pattern, details)
}

func NoRegexMatchError(pattern string) ActionResult {
	return NewErrorBuilder(ErrorCategoryExecution, "NO_REGEX_MATCH").
		WithTemplate("No matches found for pattern: %s").
		Build(pattern)
}

func InvalidCaptureGroupError(group, available int) ActionResult {
	return NewErrorBuilder(ErrorCategoryValidation, "INVALID_CAPTURE_GROUP").
		WithTemplate("Capture group %d not found (only %d groups available)").
		Build(group, available)
}

// Variable errors
func UnresolvedVariableError(count int, args []int) ActionResult {
	return NewErrorBuilder(ErrorCategoryVariable, "UNRESOLVED_VARIABLES").
		WithTemplate("action failed: %d unresolved variable(s) in arguments %v").
		WithContext("unresolved_args", args).
		WithSuggestion("Check that all variables used in arguments are defined").
		WithSuggestion("Use variable action to debug missing variables").
		Build(count, args)
}

// Go error functions for internal operations (not ActionResults)

// Extraction errors
func NewExtractionError(details string) error {
	return fmt.Errorf("extraction failed: %s", details)
}

func NewUnsupportedExtractionTypeError(extractionType string) error {
	return fmt.Errorf("unsupported extraction type: %s", extractionType)
}

func NewInvalidRegexPatternError(pattern, details string) error {
	return fmt.Errorf("invalid regex pattern '%s': %s", pattern, details)
}

func NewNoRegexMatchError(pattern string) error {
	return fmt.Errorf("no matches found for pattern: %s", pattern)
}

func NewInvalidCaptureGroupError(group, available int) error {
	return fmt.Errorf("capture group %d not found (only %d groups available)", group, available)
}

func NewNilDataError() error {
	return fmt.Errorf("cannot extract from nil data")
}

// Assertion failures (these return FAILED status for logical failures)
func AssertionFailure(expected, actual any, operator string) ActionResult {
	return NewFailureBuilder(FailureCategoryAssertion, "ASSERTION_FAILED").
		WithTemplate("Assertion failed: expected %v %s %v, but got %v").
		WithExpected(expected).
		WithActual(actual).
		WithComparison(operator).
		WithSuggestion("Check that your test data matches the expected values").
		WithSuggestion("Verify that variables are properly substituted").
		Build(actual, operator, expected, actual)
}

func BooleanAssertionFailure(actual any) ActionResult {
	return NewFailureBuilder(FailureCategoryAssertion, "BOOLEAN_ASSERTION_FAILED").
		WithTemplate("Boolean assertion failed: expected true, got %v (%T)").
		WithExpected(true).
		WithActual(actual).
		WithComparison("boolean equality").
		WithSuggestion("Ensure your condition evaluates to a boolean true value").
		WithSuggestion("Check if variables are properly resolved and contain expected values").
		Build(actual, actual)
}
