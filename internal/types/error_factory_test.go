package types

import (
	"strings"
	"testing"
)

func TestErrorFactory(t *testing.T) {
	factory := NewErrorFactory()

	// Test validation error
	validationErr := factory.CreateValidationErrorWithTemplate("VAL001", "Field '%s' is required", "username")
	if validationErr.ErrorInfo.Category != ErrorCategoryValidation {
		t.Errorf("Expected validation category, got %s", validationErr.ErrorInfo.Category)
	}
	if validationErr.ErrorInfo.Code != "VAL001" {
		t.Errorf("Expected code VAL001, got %s", validationErr.ErrorInfo.Code)
	}
	if !strings.Contains(validationErr.ErrorInfo.Message, "username") {
		t.Errorf("Expected message to contain 'username', got: %s", validationErr.ErrorInfo.Message)
	}

	// Test execution error
	execErr := factory.CreateExecutionError("http_get", "connection timeout")
	if execErr.ErrorInfo.Category != ErrorCategoryExecution {
		t.Errorf("Expected execution category, got %s", execErr.ErrorInfo.Category)
	}
	if execErr.ErrorInfo.Context["action"] != "http_get" {
		t.Errorf("Expected action context to be 'http_get', got %v", execErr.ErrorInfo.Context["action"])
	}

	// Test assertion error
	assertErr := factory.CreateAssertionError("actual", "expected", "==")
	if assertErr.ErrorInfo.Category != ErrorCategoryAssertion {
		t.Errorf("Expected assertion category, got %s", assertErr.ErrorInfo.Category)
	}
	if assertErr.ErrorInfo.Context["actual"] != "actual" {
		t.Errorf("Expected actual context to be 'actual', got %v", assertErr.ErrorInfo.Context["actual"])
	}

	// Test variable error
	varErr := factory.CreateVariableError("missing_var", "${missing_var}")
	if varErr.ErrorInfo.Category != ErrorCategoryVariable {
		t.Errorf("Expected variable category, got %s", varErr.ErrorInfo.Category)
	}
	if len(varErr.ErrorInfo.Suggestions) == 0 {
		t.Error("Expected variable error to have suggestions")
	}

	// Test network error
	netErr := factory.CreateNetworkError("NET001", "connection refused")
	if netErr.ErrorInfo.Category != ErrorCategoryNetwork {
		t.Errorf("Expected network category, got %s", netErr.ErrorInfo.Category)
	}

	// Test system error
	sysErr := factory.CreateSystemError("SYS001", "file not found")
	if sysErr.ErrorInfo.Category != ErrorCategorySystem {
		t.Errorf("Expected system category, got %s", sysErr.ErrorInfo.Category)
	}

	// Test HTTP error
	httpErr := factory.CreateHTTPError("GET", "http://example.com", 404)
	if httpErr.ErrorInfo.Category != ErrorCategoryNetwork {
		t.Errorf("Expected network category, got %s", httpErr.ErrorInfo.Category)
	}
	if httpErr.ErrorInfo.Context["method"] != "GET" {
		t.Errorf("Expected method context to be 'GET', got %v", httpErr.ErrorInfo.Context["method"])
	}
	if httpErr.ErrorInfo.Context["status_code"] != 404 {
		t.Errorf("Expected status_code context to be 404, got %v", httpErr.ErrorInfo.Context["status_code"])
	}

	// Test database error
	dbErr := factory.CreateDatabaseError("connection timeout")
	if dbErr.ErrorInfo.Category != ErrorCategorySystem {
		t.Errorf("Expected system category, got %s", dbErr.ErrorInfo.Category)
	}
	if len(dbErr.ErrorInfo.Suggestions) == 0 {
		t.Error("Expected database error to have suggestions")
	}
}

func TestDefaultErrorFactory(t *testing.T) {
	factory1 := GetDefaultErrorFactory()
	factory2 := GetDefaultErrorFactory()

	// Test that it returns the same instance (singleton pattern)
	if factory1 != factory2 {
		t.Error("Expected GetDefaultErrorFactory to return the same instance")
	}

	// Test that it works
	err := factory1.CreateValidationErrorWithTemplate("TEST", "Test message")
	if err.ErrorInfo == nil {
		t.Error("Expected error info to be set")
	}
}

func TestErrorFactoryTemplateRegistration(t *testing.T) {
	factory := NewErrorFactory()

	// Test registering a valid template
	err := factory.RegisterTemplate("test.template", "Test message: %s")
	if err != nil {
		t.Errorf("Expected no error registering valid template, got: %v", err)
	}

	// Test retrieving the registered template
	template, exists := factory.GetTemplate("test.template")
	if !exists {
		t.Error("Expected template to exist after registration")
	}
	if template != "Test message: %s" {
		t.Errorf("Expected template 'Test message: %%s', got: %s", template)
	}

	// Test registering an invalid template (with dangerous format specifier)
	err = factory.RegisterTemplate("dangerous.template", "Dangerous: %n")
	if err == nil {
		t.Error("Expected error when registering template with dangerous format specifier")
	}
}

func TestErrorFactoryTemplateValidation(t *testing.T) {
	factory := NewErrorFactory()

	// Test validation of all templates
	err := factory.ValidateAllTemplates()
	if err != nil {
		t.Errorf("Expected no error validating default templates, got: %v", err)
	}
}

func TestErrorFactoryWithPredefinedTemplates(t *testing.T) {
	factory := NewErrorFactory()

	// Test using predefined assertion template
	assertErr := factory.CreateAssertionErrorWithTemplate("ASSERT_FAILED", "assertion.failed", "expected", "==", "expected", "actual")
	if assertErr.ErrorInfo.Category != ErrorCategoryAssertion {
		t.Errorf("Expected assertion category, got %s", assertErr.ErrorInfo.Category)
	}
	if !strings.Contains(assertErr.ErrorInfo.Message, "expected") {
		t.Errorf("Expected message to contain 'expected', got: %s", assertErr.ErrorInfo.Message)
	}

	// Test using predefined variable template
	varErr := factory.CreateVariableErrorWithTemplate("VAR_UNRESOLVED", "variable.unresolved", "missing_var", "${missing_var}")
	if varErr.ErrorInfo.Category != ErrorCategoryVariable {
		t.Errorf("Expected variable category, got %s", varErr.ErrorInfo.Category)
	}
	if !strings.Contains(varErr.ErrorInfo.Message, "missing_var") {
		t.Errorf("Expected message to contain 'missing_var', got: %s", varErr.ErrorInfo.Message)
	}

	// Test using predefined HTTP template
	httpErr := factory.CreateNetworkErrorWithTemplate("HTTP_FAILED", "http.request_failed", "GET", "http://example.com", 404)
	if httpErr.ErrorInfo.Category != ErrorCategoryNetwork {
		t.Errorf("Expected network category, got %s", httpErr.ErrorInfo.Category)
	}
	if !strings.Contains(httpErr.ErrorInfo.Message, "GET") {
		t.Errorf("Expected message to contain 'GET', got: %s", httpErr.ErrorInfo.Message)
	}
}

func TestErrorFactoryWithContext(t *testing.T) {
	factory := NewErrorFactory()

	context := map[string]any{
		"step":   1,
		"action": "test_action",
	}

	err := factory.CreateErrorWithContext(ErrorCategoryExecution, "EXEC_FAILED", "action.execution_failed", context, "test_action", "test error")
	if err.ErrorInfo.Category != ErrorCategoryExecution {
		t.Errorf("Expected execution category, got %s", err.ErrorInfo.Category)
	}
	if err.ErrorInfo.Context["step"] != 1 {
		t.Errorf("Expected step context to be 1, got %v", err.ErrorInfo.Context["step"])
	}
	if err.ErrorInfo.Context["action"] != "test_action" {
		t.Errorf("Expected action context to be 'test_action', got %v", err.ErrorInfo.Context["action"])
	}
}

func TestErrorFactoryWithSuggestions(t *testing.T) {
	factory := NewErrorFactory()

	suggestions := []string{
		"Check your configuration",
		"Verify network connectivity",
	}

	err := factory.CreateErrorWithSuggestions(ErrorCategoryNetwork, "NET_FAILED", "http.connection_failed", suggestions, "connection refused")
	if err.ErrorInfo.Category != ErrorCategoryNetwork {
		t.Errorf("Expected network category, got %s", err.ErrorInfo.Category)
	}
	if len(err.ErrorInfo.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(err.ErrorInfo.Suggestions))
	}
	if err.ErrorInfo.Suggestions[0] != "Check your configuration" {
		t.Errorf("Expected first suggestion to be 'Check your configuration', got: %s", err.ErrorInfo.Suggestions[0])
	}
}
