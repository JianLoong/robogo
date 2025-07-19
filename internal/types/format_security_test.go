package types

import (
	"strings"
	"testing"
)

func TestFormatStringSecurity(t *testing.T) {
	formatter := NewSafeFormatter()

	// Test dangerous format specifiers are blocked
	dangerousTemplates := []string{
		"Test " + "%n" + " format",
		"Test " + "%*" + " format",
		"Test " + "%#" + " format",
	}

	for _, template := range dangerousTemplates {
		// Use ValidateTemplate directly to test security without triggering Go vet
		err := formatter.ValidateTemplate(template)
		if err == nil {
			t.Errorf("Expected error for dangerous template: %s", template)
		}
		if !strings.Contains(err.Error(), "not allowed") {
			t.Errorf("Expected 'not allowed' in error message for template: %s, got: %s", template, err.Error())
		}
	}

	// Test safe templates work correctly
	safeTemplates := []struct {
		template string
		args     []any
		expected string
	}{
		{"Hello %s", []any{"world"}, "Hello world"},
		{"Number: %d", []any{42}, "Number: 42"},
		{"Value: %v", []any{"test"}, "Value: test"},
		{"Multiple: %s %d %v", []any{"test", 42, true}, "Multiple: test 42 true"},
	}

	for _, test := range safeTemplates {
		result, err := formatter.Format(test.template, test.args...)
		if err != nil {
			t.Errorf("Unexpected error for safe template '%s': %s", test.template, err.Error())
		}
		if result != test.expected {
			t.Errorf("Expected '%s', got '%s' for template '%s'", test.expected, result, test.template)
		}
	}
}

func TestErrorBuilderSecurity(t *testing.T) {
	// Test that ErrorBuilder uses SafeFormatter
	result := NewErrorBuilder(ErrorCategorySystem, "TEST_ERROR").
		WithTemplate("Test error: %s").
		Build("safe_value")

	if result.Status != ActionStatusError {
		t.Errorf("Expected error status, got %s", result.Status)
	}

	if result.ErrorInfo == nil {
		t.Error("Expected ErrorInfo to be set")
	}

	if !strings.Contains(result.ErrorInfo.Message, "Test error: safe_value") {
		t.Errorf("Expected formatted message, got: %s", result.ErrorInfo.Message)
	}

	// Test that dangerous format strings are handled safely
	dangerousTemplate := "Dangerous " + "%n" + " format"
	dangerousResult := NewErrorBuilder(ErrorCategorySystem, "DANGEROUS_ERROR").
		WithTemplate(dangerousTemplate).
		Build("test")

	if dangerousResult.ErrorInfo == nil {
		t.Error("Expected ErrorInfo to be set for dangerous format")
	}

	// Should contain error about formatting failure
	if !strings.Contains(dangerousResult.ErrorInfo.Message, "Error formatting failed") {
		t.Errorf("Expected formatting failure message, got: %s", dangerousResult.ErrorInfo.Message)
	}
}
