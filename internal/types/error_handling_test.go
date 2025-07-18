package types

import (
	"strings"
	"testing"
	"time"
)

func TestErrorBuilder(t *testing.T) {
	// Test basic error building
	builder := NewErrorBuilder(ErrorCategoryValidation, "TEST001")
	result := builder.
		WithTemplate("Test error: %s").
		WithContext("field", "username").
		WithSuggestion("Check the field value").
		Build("invalid value")

	if result.Status != ActionStatusError {
		t.Errorf("Expected status to be error, got %s", result.Status)
	}

	if result.ErrorInfo == nil {
		t.Fatal("Expected ErrorInfo to be set")
	}

	if result.ErrorInfo.Category != ErrorCategoryValidation {
		t.Errorf("Expected category to be validation, got %s", result.ErrorInfo.Category)
	}

	if result.ErrorInfo.Code != "TEST001" {
		t.Errorf("Expected code to be TEST001, got %s", result.ErrorInfo.Code)
	}

	expectedMessage := "Test error: invalid value"
	if result.ErrorInfo.Message != expectedMessage {
		t.Errorf("Expected message to be '%s', got '%s'", expectedMessage, result.ErrorInfo.Message)
	}

	if len(result.ErrorInfo.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(result.ErrorInfo.Suggestions))
	}

	if result.ErrorInfo.Context["field"] != "username" {
		t.Errorf("Expected context field to be 'username', got %v", result.ErrorInfo.Context["field"])
	}

	// Verify timestamp is recent
	if time.Since(result.ErrorInfo.Timestamp) > time.Second {
		t.Error("Expected timestamp to be recent")
	}
}

func TestSafeFormatter(t *testing.T) {
	formatter := NewSafeFormatter()

	// Test basic formatting
	result, err := formatter.Format("Hello %s", "world")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "Hello world" {
		t.Errorf("Expected 'Hello world', got '%s'", result)
	}

	// Test dangerous format specifier detection
	dangerousFormat := "Dangerous %" + "n format"
	_, err = formatter.Format(dangerousFormat, 42)
	if err == nil {
		t.Error("Expected error for dangerous format specifier %n")
	}
	if !strings.Contains(err.Error(), "dangerous format specifier") {
		t.Errorf("Expected error message about dangerous format specifier, got: %s", err.Error())
	}

	// Test template registration and retrieval
	formatter.RegisterTemplate("test", "Test template: %s")
	template, exists := formatter.GetTemplate("test")
	if !exists {
		t.Error("Expected template to exist")
	}
	if template != "Test template: %s" {
		t.Errorf("Expected 'Test template: %%s', got '%s'", template)
	}
}

func TestDefaultTemplates(t *testing.T) {
	formatter := GetDefaultSafeFormatter()

	// Test that default templates are loaded
	template, exists := formatter.GetTemplate("assertion.failed")
	if !exists {
		t.Error("Expected assertion.failed template to exist")
	}
	if !strings.Contains(template, "Assertion failed") {
		t.Errorf("Expected assertion template to contain 'Assertion failed', got: %s", template)
	}

	// Test formatting with default template
	result, err := formatter.Format(template, "expected", "==", "actual", "different")
	if err != nil {
		t.Errorf("Expected no error formatting default template, got %v", err)
	}
	if !strings.Contains(result, "expected") || !strings.Contains(result, "actual") {
		t.Errorf("Expected formatted result to contain expected values, got: %s", result)
	}
}

func TestActionResultHelpers(t *testing.T) {
	// Test success result
	success := NewSuccessResult()
	if !success.IsSuccess() {
		t.Error("Expected IsSuccess to return true")
	}
	if success.IsError() {
		t.Error("Expected IsError to return false")
	}

	// Test success with data
	successWithData := NewSuccessResultWithData("test data")
	if successWithData.Data != "test data" {
		t.Errorf("Expected data to be 'test data', got %v", successWithData.Data)
	}

	// Test skipped result
	skipped := NewSkippedResult("test reason")
	if !skipped.IsSkipped() {
		t.Error("Expected IsSkipped to return true")
	}
	if skipped.GetSkipReason() != "test reason" {
		t.Errorf("Expected skip reason to be 'test reason', got '%s'", skipped.GetSkipReason())
	}

	// Test error result with ErrorInfo
	errorResult := NewErrorBuilder(ErrorCategoryExecution, "ERR001").
		WithTemplate("Test error").
		Build()

	if !errorResult.IsError() {
		t.Error("Expected IsError to return true")
	}
	if errorResult.GetErrorMessage() != "Test error" {
		t.Errorf("Expected error message to be 'Test error', got '%s'", errorResult.GetErrorMessage())
	}
}

func TestErrorCategories(t *testing.T) {
	categories := []ErrorCategory{
		ErrorCategoryValidation,
		ErrorCategoryExecution,
		ErrorCategoryAssertion,
		ErrorCategoryVariable,
		ErrorCategoryNetwork,
		ErrorCategorySystem,
	}

	expectedValues := []string{
		"validation",
		"execution",
		"assertion",
		"variable",
		"network",
		"system",
	}

	for i, category := range categories {
		if string(category) != expectedValues[i] {
			t.Errorf("Expected category %d to be '%s', got '%s'", i, expectedValues[i], string(category))
		}
	}
}
