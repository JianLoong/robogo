package types

import (
	"testing"
)

func TestNewVariableContext(t *testing.T) {
	template := "Hello ${user.name}!"
	availableVars := map[string]any{
		"user": map[string]any{
			"name": "John",
			"age":  30,
		},
	}
	
	context := NewVariableContext(template, availableVars)
	
	if context.OriginalTemplate != template {
		t.Errorf("Expected original template '%s', got '%s'", template, context.OriginalTemplate)
	}
	
	if context.Status != VariableStatusUnresolved {
		t.Errorf("Expected initial status %s, got %s", VariableStatusUnresolved, context.Status)
	}
	
	if len(context.AvailableVariables) != len(availableVars) {
		t.Errorf("Expected %d available variables, got %d", len(availableVars), len(context.AvailableVariables))
	}
}

func TestParseVariableExpression(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		expected    []string
	}{
		{
			name:     "single variable",
			template: "Hello ${name}!",
			expected: []string{"name"},
		},
		{
			name:     "multiple variables",
			template: "${greeting} ${user.name}, you are ${user.age} years old",
			expected: []string{"greeting", "user.name", "user.age"},
		},
		{
			name:     "no variables",
			template: "Hello world!",
			expected: []string{},
		},
		{
			name:     "expression with operators",
			template: "Result: ${x + y * 2}",
			expected: []string{"x + y * 2"},
		},
		{
			name:     "nested expressions",
			template: "${data.items[0].name}",
			expected: []string{"data.items[0].name"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseVariableExpression(tt.template)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d expressions, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected expression '%s', got '%s'", expected, result[i])
				}
			}
		})
	}
}

func TestAnalyzeExpression(t *testing.T) {
	availableVars := map[string]any{
		"user": map[string]any{
			"name":    "John",
			"age":     30,
			"profile": map[string]any{
				"email": "john@example.com",
			},
		},
		"items": []any{"apple", "banana", "cherry"},
		"count": 42,
	}
	
	tests := []struct {
		name           string
		expression     string
		expectedStatus VariableResolutionStatus
		expectedReason VariableFailureReason
		shouldHaveValue bool
	}{
		{
			name:            "simple variable found",
			expression:      "count",
			expectedStatus:  VariableStatusResolved,
			shouldHaveValue: true,
		},
		{
			name:            "nested property found",
			expression:      "user.name",
			expectedStatus:  VariableStatusResolved,
			shouldHaveValue: true,
		},
		{
			name:            "deep nested property found",
			expression:      "user.profile.email",
			expectedStatus:  VariableStatusResolved,
			shouldHaveValue: true,
		},
		{
			name:            "array length access",
			expression:      "items.length",
			expectedStatus:  VariableStatusResolved,
			shouldHaveValue: true,
		},
		{
			name:            "array index access",
			expression:      "items.0",
			expectedStatus:  VariableStatusResolved,
			shouldHaveValue: true,
		},
		{
			name:           "variable not found",
			expression:     "nonexistent",
			expectedStatus: VariableStatusUnresolved,
			expectedReason: FailureReasonNotFound,
		},
		{
			name:           "property not found",
			expression:     "user.nonexistent",
			expectedStatus: VariableStatusUnresolved,
			expectedReason: FailureReasonAccessError,
		},
		{
			name:           "array index out of bounds",
			expression:     "items.10",
			expectedStatus: VariableStatusUnresolved,
			expectedReason: FailureReasonAccessError,
		},
		{
			name:           "invalid array index",
			expression:     "items.invalid",
			expectedStatus: VariableStatusUnresolved,
			expectedReason: FailureReasonAccessError,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempt := AnalyzeExpression(tt.expression, availableVars)
			
			if attempt.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, attempt.Status)
			}
			
			if tt.expectedStatus != VariableStatusResolved && attempt.FailureReason != tt.expectedReason {
				t.Errorf("Expected failure reason %s, got %s", tt.expectedReason, attempt.FailureReason)
			}
			
			if tt.shouldHaveValue && attempt.ResolvedValue == nil {
				t.Error("Expected resolved value, got nil")
			}
			
			if !tt.shouldHaveValue && attempt.ResolvedValue != nil {
				t.Errorf("Expected no resolved value, got %v", attempt.ResolvedValue)
			}
			
			if attempt.Expression != tt.expression {
				t.Errorf("Expected expression '%s', got '%s'", tt.expression, attempt.Expression)
			}
		})
	}
}

func TestVariableContext_AddAttempt(t *testing.T) {
	context := NewVariableContext("test template", map[string]any{})
	
	// Add successful attempt
	successAttempt := VariableAccessAttempt{
		Expression: "test",
		Status:     VariableStatusResolved,
	}
	context.AddAttempt(successAttempt)
	
	if context.ResolvedCount != 1 {
		t.Errorf("Expected resolved count 1, got %d", context.ResolvedCount)
	}
	
	if context.UnresolvedCount != 0 {
		t.Errorf("Expected unresolved count 0, got %d", context.UnresolvedCount)
	}
	
	if context.Status != VariableStatusResolved {
		t.Errorf("Expected status %s, got %s", VariableStatusResolved, context.Status)
	}
	
	// Add failed attempt
	failedAttempt := VariableAccessAttempt{
		Expression:    "missing",
		Status:        VariableStatusUnresolved,
		FailureReason: FailureReasonNotFound,
	}
	context.AddAttempt(failedAttempt)
	
	if context.ResolvedCount != 1 {
		t.Errorf("Expected resolved count 1, got %d", context.ResolvedCount)
	}
	
	if context.UnresolvedCount != 1 {
		t.Errorf("Expected unresolved count 1, got %d", context.UnresolvedCount)
	}
	
	if context.Status != VariableStatusPartial {
		t.Errorf("Expected status %s, got %s", VariableStatusPartial, context.Status)
	}
}

func TestVariableContext_GetUnresolvedVariables(t *testing.T) {
	context := NewVariableContext("test template", map[string]any{})
	
	// Add mix of resolved and unresolved attempts
	context.AddAttempt(VariableAccessAttempt{
		Expression: "resolved",
		Status:     VariableStatusResolved,
	})
	
	context.AddAttempt(VariableAccessAttempt{
		Expression:    "unresolved1",
		Status:        VariableStatusUnresolved,
		FailureReason: FailureReasonNotFound,
	})
	
	context.AddAttempt(VariableAccessAttempt{
		Expression:    "unresolved2",
		Status:        VariableStatusUnresolved,
		FailureReason: FailureReasonAccessError,
	})
	
	unresolved := context.GetUnresolvedVariables()
	
	if len(unresolved) != 2 {
		t.Errorf("Expected 2 unresolved variables, got %d", len(unresolved))
	}
	
	// Check that only unresolved variables are returned
	for _, attempt := range unresolved {
		if attempt.Status == VariableStatusResolved {
			t.Error("Found resolved variable in unresolved list")
		}
	}
}

func TestVariableFailureAnalysis_generateRecommendations(t *testing.T) {
	context := NewVariableContext("${missing} ${user.invalid}", map[string]any{
		"user": map[string]any{"name": "John"},
		"data": map[string]any{"value": 42},
	})
	
	// Add various failure types
	context.AddAttempt(VariableAccessAttempt{
		Expression:    "missing",
		Status:        VariableStatusUnresolved,
		FailureReason: FailureReasonNotFound,
	})
	
	context.AddAttempt(VariableAccessAttempt{
		Expression:    "user.invalid",
		Status:        VariableStatusUnresolved,
		FailureReason: FailureReasonAccessError,
	})
	
	analysis := context.GetFailureAnalysis()
	
	if len(analysis.Recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}
	
	// Check for specific recommendation types
	foundTypoCheck := false
	foundAccessCheck := false
	foundVariableDebug := false
	
	for _, rec := range analysis.Recommendations {
		if contains(rec, "typo") || contains(rec, "case sensitivity") {
			foundTypoCheck = true
		}
		if contains(rec, "nested property") || contains(rec, "access paths") {
			foundAccessCheck = true
		}
		if contains(rec, "variable action") {
			foundVariableDebug = true
		}
	}
	
	if !foundTypoCheck {
		t.Error("Expected typo/case sensitivity recommendation")
	}
	if !foundAccessCheck {
		t.Error("Expected nested property access recommendation")
	}
	if !foundVariableDebug {
		t.Error("Expected variable debugging recommendation")
	}
}

func TestGetDetailedErrorMessage(t *testing.T) {
	context := NewVariableContext("Hello ${missing}!", map[string]any{
		"user": "John",
	})
	
	context.AddAttempt(VariableAccessAttempt{
		Expression:    "missing",
		Status:        VariableStatusUnresolved,
		FailureReason: FailureReasonNotFound,
		ErrorMessage:  "Variable 'missing' not found",
		Suggestions:   []string{"Check variable name", "Use 'user' instead"},
	})
	
	message := context.GetDetailedErrorMessage()
	
	if message == "" {
		t.Error("Expected non-empty error message")
	}
	
	// Check that message contains key information
	if !contains(message, "Hello ${missing}!") {
		t.Error("Expected error message to contain original template")
	}
	
	if !contains(message, "missing") {
		t.Error("Expected error message to contain failed variable name")
	}
	
	if !contains(message, "user") {
		t.Error("Expected error message to contain available variables")
	}
}

func TestFindSimilarVariables(t *testing.T) {
	availableVars := map[string]any{
		"user":     "John",
		"userName": "john123",
		"userData": map[string]any{},
		"count":    42,
		"counter":  43,
	}
	
	tests := []struct {
		name     string
		target   string
		expected []string
	}{
		{
			name:     "exact match different case",
			target:   "USER",
			expected: []string{"user", "userName", "userData"}, // All contain "user"
		},
		{
			name:     "partial match",
			target:   "name",
			expected: []string{"userName"},
		},
		{
			name:     "similar names",
			target:   "count",
			expected: []string{"count", "counter"}, // Exact match and similar
		},
		{
			name:     "no similar",
			target:   "xyz",
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findSimilarVariables(tt.target, availableVars)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d similar variables, got %d", len(tt.expected), len(result))
				return
			}
			
			for _, expected := range tt.expected {
				found := false
				for _, actual := range result {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find '%s' in similar variables", expected)
				}
			}
		})
	}
}

// Using contains function from assertion_context_test.go