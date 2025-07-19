package types

import (
	"fmt"
	"testing"
)

// Example demonstrates basic usage of the ErrorBuilder
func ExampleErrorBuilder() {
	// Create a validation error with context and suggestions
	result := NewErrorBuilder(ErrorCategoryValidation, "FIELD_REQUIRED").
		WithTemplate("Field '%s' is required but was empty").
		WithContext("field_name", "username").
		WithContext("field_type", "string").
		WithSuggestion("Provide a non-empty value for the username field").
		WithSuggestion("Check the input validation rules").
		Build("username")

	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Category: %s\n", result.ErrorInfo.Category)
	fmt.Printf("Code: %s\n", result.ErrorInfo.Code)
	fmt.Printf("Message: %s\n", result.ErrorInfo.Message)
	fmt.Printf("Suggestions: %d\n", len(result.ErrorInfo.Suggestions))

	// Output:
	// Status: error
	// Category: validation
	// Code: FIELD_REQUIRED
	// Message: Field 'username' is required but was empty
	// Suggestions: 2
}

// Example demonstrates usage of the ErrorFactory
func ExampleErrorFactory() {
	factory := GetDefaultErrorFactory()

	// Create different types of errors
	validationErr := factory.CreateValidationError("INVALID_EMAIL", "Email '%s' is not valid", "invalid-email")
	httpErr := factory.CreateHTTPError("GET", "https://api.example.com/users", 404)
	dbErr := factory.CreateDatabaseError("connection timeout after 30 seconds")

	fmt.Printf("Validation: %s\n", validationErr.ErrorInfo.Category)
	fmt.Printf("HTTP: %s\n", httpErr.ErrorInfo.Category)
	fmt.Printf("Database: %s\n", dbErr.ErrorInfo.Category)

	// Output:
	// Validation: validation
	// HTTP: network
	// Database: system
}

// Example demonstrates safe formatting
func ExampleSafeFormatter() {
	formatter := GetDefaultSafeFormatter()

	// Safe formatting with validation
	message, err := formatter.Format("User %s has %d unread messages", "john", 5)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("Message: %s\n", message)

	// Using registered templates
	template, exists := formatter.GetTemplate("assertion.failed")
	if exists {
		formatted, _ := formatter.Format(template, "expected", "==", "actual", "different")
		fmt.Printf("Assertion: %s\n", formatted)
	}

	// Output:
	// Message: User john has 5 unread messages
	// Assertion: Assertion failed: expected expected == actual, but got different
}

// TestErrorHandlingIntegration demonstrates how the new system integrates
func TestErrorHandlingIntegration(t *testing.T) {
	// Simulate an action that might fail
	simulateAction := func(input string) ActionResult {
		if input == "" {
			return NewErrorBuilder(ErrorCategoryValidation, "EMPTY_INPUT").
				WithTemplate("Input cannot be empty").
				WithSuggestion("Provide a non-empty input value").
				Build()
		}

		if input == "fail" {
			return GetDefaultErrorFactory().CreateExecutionError("test_action", "simulated failure")
		}

		return NewSuccessResultWithData(fmt.Sprintf("Processed: %s", input))
	}

	// Test successful case
	result := simulateAction("valid input")
	if !result.IsSuccess() {
		t.Errorf("Expected success, got %s", result.Status)
	}

	// Test validation error
	result = simulateAction("")
	if !result.IsError() {
		t.Error("Expected error for empty input")
	}
	if result.ErrorInfo.Category != ErrorCategoryValidation {
		t.Errorf("Expected validation error, got %s", result.ErrorInfo.Category)
	}

	// Test execution error
	result = simulateAction("fail")
	if !result.IsError() {
		t.Error("Expected error for 'fail' input")
	}
	if result.ErrorInfo.Category != ErrorCategoryExecution {
		t.Errorf("Expected execution error, got %s", result.ErrorInfo.Category)
	}

	// Test error message retrieval
	errorMsg := result.GetErrorMessage()
	if errorMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestErrorBuilderBasics ensures the ErrorBuilder works correctly
func TestErrorBuilderBasics(t *testing.T) {
	// Test ErrorBuilder creates proper errors
	errorResult := NewErrorBuilder(ErrorCategorySystem, "TEST_ERROR").
		WithTemplate("Test error: %s").
		Build("test")

	if errorResult.Status != ActionStatusError {
		t.Errorf("Expected error status, got %s", errorResult.Status)
	}
	if errorResult.ErrorInfo == nil {
		t.Error("Expected ErrorInfo to be set for error")
	}
	if errorResult.ErrorInfo.Message != "Test error: test" {
		t.Errorf("Expected 'Test error: test', got '%s'", errorResult.ErrorInfo.Message)
	}

	// Test FailureBuilder creates proper failures
	failureResult := NewFailureBuilder(FailureCategoryAssertion, "TEST_FAILURE").
		WithTemplate("Test failure: expected %v, got %v").
		WithExpected("expected").
		WithActual("actual").
		Build("expected", "actual")

	if failureResult.Status != ActionStatusFailed {
		t.Errorf("Expected failed status, got %s", failureResult.Status)
	}
	if failureResult.FailureInfo == nil {
		t.Error("Expected FailureInfo to be set for failure")
	}
	if failureResult.FailureInfo.Expected != "expected" {
		t.Errorf("Expected 'expected', got '%v'", failureResult.FailureInfo.Expected)
	}
	if failureResult.FailureInfo.Actual != "actual" {
		t.Errorf("Expected 'actual', got '%v'", failureResult.FailureInfo.Actual)
	}
}
