package types

import (
	"testing"
)

func TestNewValueInfo(t *testing.T) {
	tests := []struct {
		name                string
		value               any
		expectedType        string
		expectedString      string
		expectedNumeric     bool
		expectedNativeNumeric bool
		expectedCanParse    bool
	}{
		{
			name:                "nil value",
			value:               nil,
			expectedType:        "nil",
			expectedString:      "<nil>",
			expectedNumeric:     false,
			expectedNativeNumeric: false,
			expectedCanParse:    false,
		},
		{
			name:                "string number",
			value:               "42",
			expectedType:        "string",
			expectedString:      "42",
			expectedNumeric:     true,
			expectedNativeNumeric: false,
			expectedCanParse:    true,
		},
		{
			name:                "string text",
			value:               "hello",
			expectedType:        "string",
			expectedString:      "hello",
			expectedNumeric:     false,
			expectedNativeNumeric: false,
			expectedCanParse:    false,
		},
		{
			name:                "integer",
			value:               42,
			expectedType:        "int",
			expectedString:      "42",
			expectedNumeric:     true,
			expectedNativeNumeric: true,
			expectedCanParse:    true,
		},
		{
			name:                "float",
			value:               42.5,
			expectedType:        "float64",
			expectedString:      "42.5",
			expectedNumeric:     true,
			expectedNativeNumeric: true,
			expectedCanParse:    true,
		},
		{
			name:                "boolean",
			value:               true,
			expectedType:        "bool",
			expectedString:      "true",
			expectedNumeric:     false,
			expectedNativeNumeric: false,
			expectedCanParse:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewValueInfo(tt.value)
			
			if info.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, info.Type)
			}
			
			if info.StringValue != tt.expectedString {
				t.Errorf("Expected string value %s, got %s", tt.expectedString, info.StringValue)
			}
			
			if info.IsNumeric != tt.expectedNumeric {
				t.Errorf("Expected IsNumeric %v, got %v", tt.expectedNumeric, info.IsNumeric)
			}

			if info.IsNativeNumeric != tt.expectedNativeNumeric {
				t.Errorf("Expected IsNativeNumeric %v, got %v", tt.expectedNativeNumeric, info.IsNativeNumeric)
			}

			if info.CanParseNumeric != tt.expectedCanParse {
				t.Errorf("Expected CanParseNumeric %v, got %v", tt.expectedCanParse, info.CanParseNumeric)
			}
			
			if tt.expectedNumeric && info.NumericValue == nil {
				t.Error("Expected NumericValue to be set for numeric values")
			}
			
			if !tt.expectedNumeric && info.NumericValue != nil {
				t.Error("Expected NumericValue to be nil for non-numeric values")
			}
		})
	}
}

func TestNewBooleanAssertionContext(t *testing.T) {
	ctx := NewBooleanAssertionContext(false, "test message")
	
	if ctx.Type != AssertionTypeBoolean {
		t.Errorf("Expected type %s, got %s", AssertionTypeBoolean, ctx.Type)
	}
	
	if ctx.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", ctx.Message)
	}
	
	if ctx.Actual == nil {
		t.Error("Expected Actual to be set")
	}
	
	if ctx.Actual.Type != "bool" {
		t.Errorf("Expected actual type 'bool', got '%s'", ctx.Actual.Type)
	}
}

func TestNewComparisonAssertionContext(t *testing.T) {
	tests := []struct {
		name               string
		actual             any
		expected           any
		operator           string
		expectedType       AssertionType
		expectedMethod     string
	}{
		{
			name:           "string equality",
			actual:         "hello",
			expected:       "world",
			operator:       "==",
			expectedType:   AssertionTypeComparison,
			expectedMethod: "string_exact",
		},
		{
			name:           "numeric comparison",
			actual:         42,
			expected:       24,
			operator:       ">",
			expectedType:   AssertionTypeComparison,
			expectedMethod: "numeric",
		},
		{
			name:           "string contains",
			actual:         "hello world",
			expected:       "world",
			operator:       "contains",
			expectedType:   AssertionTypeContains,
			expectedMethod: "string_contains",
		},
		{
			name:           "mixed types numeric comparison",
			actual:         "42",
			expected:       42,
			operator:       "==",
			expectedType:   AssertionTypeComparison,
			expectedMethod: "parsed_numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewComparisonAssertionContext(tt.actual, tt.expected, tt.operator, "")
			
			if ctx.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, ctx.Type)
			}
			
			if ctx.ComparisonMethod != tt.expectedMethod {
				t.Errorf("Expected comparison method %s, got %s", tt.expectedMethod, ctx.ComparisonMethod)
			}
			
			if ctx.Operator != tt.operator {
				t.Errorf("Expected operator %s, got %s", tt.operator, ctx.Operator)
			}
		})
	}
}

func TestNewComparisonAssertionContextWithActualComparison(t *testing.T) {
	tests := []struct {
		name               string
		actual             any
		expected           any
		operator           string
		wasNumeric         bool
		expectedType       AssertionType
		expectedCompType   ComparisonType
		expectedActual     string
		expectedBehavior   string
	}{
		{
			name:             "numeric comparison performed",
			actual:           "42",
			expected:         "24",
			operator:         ">",
			wasNumeric:       true,
			expectedType:     AssertionTypeNumeric,
			expectedCompType: ComparisonTypeNumeric,
			expectedActual:   "Numeric comparison performed",
			expectedBehavior: "Numeric comparison",
		},
		{
			name:             "string fallback performed",
			actual:           "abc",
			expected:         "def",
			operator:         ">",
			wasNumeric:       false,
			expectedType:     AssertionTypeString,
			expectedCompType: ComparisonTypeStringFallback,
			expectedActual:   "String comparison (numeric parsing failed)",
			expectedBehavior: "String comparison",
		},
		{
			name:             "mixed types with numeric fallback",
			actual:           42,
			expected:         "def",
			operator:         ">",
			wasNumeric:       false,
			expectedType:     AssertionTypeMixed,
			expectedCompType: ComparisonTypeStringFallback,
			expectedActual:   "String comparison (numeric parsing failed)",
			expectedBehavior: "Likely intended numeric comparison",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewComparisonAssertionContextWithActualComparison(tt.actual, tt.expected, tt.operator, "", tt.wasNumeric)
			
			if ctx.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, ctx.Type)
			}
			
			if ctx.ComparisonType != tt.expectedCompType {
				t.Errorf("Expected comparison type %s, got %s", tt.expectedCompType, ctx.ComparisonType)
			}
			
			if ctx.ActualComparison != tt.expectedActual {
				t.Errorf("Expected actual comparison '%s', got '%s'", tt.expectedActual, ctx.ActualComparison)
			}
			
			if ctx.ExpectedBehavior != tt.expectedBehavior {
				t.Errorf("Expected behavior '%s', got '%s'", tt.expectedBehavior, ctx.ExpectedBehavior)
			}
		})
	}
}

func TestGetTypeMismatchInfo(t *testing.T) {
	tests := []struct {
		name                string
		actual              any
		expected            any
		expectTypeMismatch  bool
	}{
		{
			name:               "same types",
			actual:             "hello",
			expected:           "world",
			expectTypeMismatch: false,
		},
		{
			name:               "different types",
			actual:             "42",
			expected:           42,
			expectTypeMismatch: true,
		},
		{
			name:               "string vs bool",
			actual:             "true",
			expected:           true,
			expectTypeMismatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewComparisonAssertionContext(tt.actual, tt.expected, "==", "")
			mismatch := ctx.GetTypeMismatchInfo()
			
			if tt.expectTypeMismatch && mismatch == nil {
				t.Error("Expected type mismatch info, got nil")
			}
			
			if !tt.expectTypeMismatch && mismatch != nil {
				t.Error("Expected no type mismatch info, got non-nil")
			}
			
			if mismatch != nil {
				if mismatch.ActualType == "" || mismatch.ExpectedType == "" {
					t.Error("Type mismatch info should have both actual and expected types")
				}
				
				if mismatch.Suggestion == "" {
					t.Error("Type mismatch info should have a suggestion")
				}
			}
		})
	}
}

func TestGetDetailedMessage(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *AssertionContext
		contains []string
	}{
		{
			name: "boolean assertion",
			ctx:  NewBooleanAssertionContext(false, ""),
			contains: []string{"Boolean assertion failed", "false", "bool"},
		},
		{
			name: "comparison assertion",
			ctx:  NewComparisonAssertionContext("hello", "world", "==", ""),
			contains: []string{"Comparison assertion failed", "hello", "==", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := tt.ctx.GetDetailedMessage()
			
			for _, expected := range tt.contains {
				if !contains(message, expected) {
					t.Errorf("Expected message to contain '%s', got: %s", expected, message)
				}
			}
		})
	}
}

func TestGetSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		ctx            *AssertionContext
		expectSuggestions bool
	}{
		{
			name:              "boolean with non-bool type",
			ctx:               NewBooleanAssertionContext("not a bool", ""),
			expectSuggestions: true,
		},
		{
			name:              "type mismatch comparison",
			ctx:               NewComparisonAssertionContext("42", 42, "==", ""),
			expectSuggestions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := tt.ctx.GetSuggestions()
			
			if tt.expectSuggestions && len(suggestions) == 0 {
				t.Error("Expected suggestions, got none")
			}
			
			for _, suggestion := range suggestions {
				if suggestion == "" {
					t.Error("Suggestions should not be empty strings")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (len(substr) == 0 || s == substr || 
		    (len(s) > len(substr) && (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}