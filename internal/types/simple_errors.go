package types

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

// Variable errors
func UnresolvedVariableError(count int, args []int) ActionResult {
	return NewErrorBuilder(ErrorCategoryVariable, "UNRESOLVED_VARIABLES").
		WithTemplate("action failed: %d unresolved variable(s) in arguments %v").
		WithContext("unresolved_args", args).
		WithSuggestion("Check that all variables used in arguments are defined").
		WithSuggestion("Use variable action to debug missing variables").
		Build(count, args)
}

// Assertion failures (these return FAILED status for logical failures)
func AssertionFailure(expected, actual any, operator string) ActionResult {
	return NewFailureBuilder(FailureCategoryAssertion, "ASSERTION_FAILED").
		WithTemplate("assertion failed: expected %v %s %v, got %v").
		WithExpected(expected).
		WithActual(actual).
		WithComparison(operator).
		Build(actual, operator, expected, actual)
}

func BooleanAssertionFailure(actual any) ActionResult {
	return NewFailureBuilder(FailureCategoryAssertion, "BOOLEAN_ASSERTION_FAILED").
		WithTemplate("assertion failed: expected true, got %v (%T)").
		WithExpected(true).
		WithActual(actual).
		WithComparison("boolean equality").
		Build(actual, actual)
}
