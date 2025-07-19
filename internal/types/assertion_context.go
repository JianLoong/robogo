package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// AssertionType represents the type of assertion being performed
type AssertionType string

const (
	AssertionTypeBoolean    AssertionType = "boolean"
	AssertionTypeComparison AssertionType = "comparison"
	AssertionTypeContains   AssertionType = "contains"
	AssertionTypeNumeric    AssertionType = "numeric"
	AssertionTypeString     AssertionType = "string"
	AssertionTypeMixed      AssertionType = "mixed"
)

// ComparisonType represents how values were actually compared
type ComparisonType string

const (
	ComparisonTypeNumeric          ComparisonType = "numeric"
	ComparisonTypeString           ComparisonType = "string"
	ComparisonTypeStringFallback   ComparisonType = "string_fallback"
	ComparisonTypeContains         ComparisonType = "contains"
	ComparisonTypeLexicographic    ComparisonType = "lexicographic"
	ComparisonTypeUnknown          ComparisonType = "unknown"
)

// ValueInfo contains detailed information about a value in an assertion
type ValueInfo struct {
	Value           any      `json:"value"`
	Type            string   `json:"type"`
	StringValue     string   `json:"string_value"`
	NumericValue    *float64 `json:"numeric_value,omitempty"`
	IsNumeric       bool     `json:"is_numeric"`
	IsNil           bool     `json:"is_nil"`
	OriginalType    string   `json:"original_type"`
	CanParseNumeric bool     `json:"can_parse_numeric"`
	IsNativeNumeric bool     `json:"is_native_numeric"`
}

// AssertionContext provides rich context information for assertion results
type AssertionContext struct {
	Type              AssertionType  `json:"type"`
	Operator          string         `json:"operator,omitempty"`
	Actual            *ValueInfo     `json:"actual,omitempty"`
	Expected          *ValueInfo     `json:"expected,omitempty"`
	ComparisonMethod  string         `json:"comparison_method,omitempty"`
	ComparisonType    ComparisonType `json:"comparison_type,omitempty"`
	Message           string         `json:"message,omitempty"`
	Timestamp         time.Time      `json:"timestamp"`
	ActualComparison  string         `json:"actual_comparison,omitempty"`  // How comparison was actually performed
	ExpectedBehavior  string         `json:"expected_behavior,omitempty"`  // How user likely intended comparison
}

// NewValueInfo creates a ValueInfo from any value with type analysis
func NewValueInfo(value any) *ValueInfo {
	if value == nil {
		return &ValueInfo{
			Value:           nil,
			Type:            "nil",
			OriginalType:    "nil",
			StringValue:     "<nil>",
			IsNil:           true,
			CanParseNumeric: false,
			IsNativeNumeric: false,
		}
	}

	valueType := reflect.TypeOf(value)
	originalType := valueType.String()
	stringValue := fmt.Sprintf("%v", value)
	
	info := &ValueInfo{
		Value:        value,
		Type:         originalType,
		OriginalType: originalType,
		StringValue:  stringValue,
		IsNil:        false,
	}

	// Determine if this is a native numeric type
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		info.IsNativeNumeric = true
		info.IsNumeric = true
		info.CanParseNumeric = true
		if numValue, err := strconv.ParseFloat(stringValue, 64); err == nil {
			info.NumericValue = &numValue
		}
	default:
		info.IsNativeNumeric = false
		// Try to parse as numeric even if not native numeric type
		if numValue, err := strconv.ParseFloat(stringValue, 64); err == nil {
			info.NumericValue = &numValue
			info.IsNumeric = true
			info.CanParseNumeric = true
		} else {
			info.IsNumeric = false
			info.CanParseNumeric = false
		}
	}

	return info
}

// NewBooleanAssertionContext creates context for boolean assertions
func NewBooleanAssertionContext(value any, message string) *AssertionContext {
	return &AssertionContext{
		Type:      AssertionTypeBoolean,
		Actual:    NewValueInfo(value),
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewComparisonAssertionContext creates context for comparison assertions
func NewComparisonAssertionContext(actual, expected any, operator, message string) *AssertionContext {
	actualInfo := NewValueInfo(actual)
	expectedInfo := NewValueInfo(expected)
	
	context := &AssertionContext{
		Type:      AssertionTypeComparison,
		Operator:  operator,
		Actual:    actualInfo,
		Expected:  expectedInfo,
		Message:   message,
		Timestamp: time.Now(),
	}

	// Determine comparison method and assertion type
	context.determineComparisonMethod()
	
	return context
}

// NewComparisonAssertionContextWithActualComparison creates context when we know how comparison was actually performed
func NewComparisonAssertionContextWithActualComparison(actual, expected any, operator, message string, wasNumeric bool) *AssertionContext {
	actualInfo := NewValueInfo(actual)
	expectedInfo := NewValueInfo(expected)
	
	context := &AssertionContext{
		Type:      AssertionTypeComparison,
		Operator:  operator,
		Actual:    actualInfo,
		Expected:  expectedInfo,
		Message:   message,
		Timestamp: time.Now(),
	}

	// Determine comparison type based on actual behavior
	context.determineActualComparisonType(wasNumeric)
	
	return context
}

// determineComparisonMethod analyzes how the comparison was performed
func (ac *AssertionContext) determineComparisonMethod() {
	if ac.Actual == nil || ac.Expected == nil {
		ac.ComparisonMethod = "unknown"
		ac.ComparisonType = ComparisonTypeUnknown
		return
	}

	switch ac.Operator {
	case "contains", "starts_with", "ends_with":
		ac.Type = AssertionTypeContains
		ac.ComparisonMethod = "string_contains"
		ac.ComparisonType = ComparisonTypeContains
		ac.ActualComparison = "String contains check"
		ac.ExpectedBehavior = "String contains check"
	case ">", "<", ">=", "<=":
		if ac.Actual.IsNumeric && ac.Expected.IsNumeric {
			if ac.Actual.IsNativeNumeric && ac.Expected.IsNativeNumeric {
				// Keep Type as AssertionTypeComparison, set comparison method details
				ac.ComparisonMethod = "numeric"
				ac.ComparisonType = ComparisonTypeNumeric
				ac.ActualComparison = "Native numeric comparison"
				ac.ExpectedBehavior = "Numeric comparison"
			} else {
				// Keep Type as AssertionTypeComparison, set comparison method details
				ac.ComparisonMethod = "parsed_numeric"
				ac.ComparisonType = ComparisonTypeNumeric
				ac.ActualComparison = "Parsed numeric comparison"
				ac.ExpectedBehavior = "Numeric comparison"
			}
		} else {
			// Keep Type as AssertionTypeComparison, set comparison method details
			ac.ComparisonMethod = "string_lexicographic"
			ac.ComparisonType = ComparisonTypeLexicographic
			ac.ActualComparison = "Lexicographic string comparison"
			if ac.Actual.CanParseNumeric && ac.Expected.CanParseNumeric {
				ac.ExpectedBehavior = "Likely intended numeric comparison"
			} else {
				ac.ExpectedBehavior = "String comparison"
			}
		}
	case "==", "!=":
		if ac.Actual.IsNumeric && ac.Expected.IsNumeric {
			if ac.Actual.IsNativeNumeric && ac.Expected.IsNativeNumeric {
				// Keep Type as AssertionTypeComparison, set comparison method details
				ac.ComparisonMethod = "numeric"
				ac.ComparisonType = ComparisonTypeNumeric
				ac.ActualComparison = "Native numeric equality"
				ac.ExpectedBehavior = "Numeric equality"
			} else {
				// Keep Type as AssertionTypeComparison, set comparison method details
				ac.ComparisonMethod = "parsed_numeric"
				ac.ComparisonType = ComparisonTypeNumeric
				ac.ActualComparison = "Parsed numeric equality"
				ac.ExpectedBehavior = "Numeric equality"
			}
		} else {
			// Keep Type as AssertionTypeComparison, set comparison method details
			ac.ComparisonMethod = "string_exact"
			ac.ComparisonType = ComparisonTypeString
			ac.ActualComparison = "String equality (after string conversion)"
			if ac.Actual.CanParseNumeric && ac.Expected.CanParseNumeric {
				ac.ExpectedBehavior = "Likely intended numeric equality"
			} else {
				ac.ExpectedBehavior = "String equality"
			}
		}
	default:
		ac.ComparisonMethod = "unknown"
		ac.ComparisonType = ComparisonTypeUnknown
	}
}

// determineActualComparisonType sets comparison type based on how comparison was actually performed
func (ac *AssertionContext) determineActualComparisonType(wasNumeric bool) {
	if ac.Actual == nil || ac.Expected == nil {
		ac.ComparisonMethod = "unknown"
		ac.ComparisonType = ComparisonTypeUnknown
		return
	}

	switch ac.Operator {
	case "contains", "starts_with", "ends_with":
		ac.Type = AssertionTypeContains
		ac.ComparisonMethod = "string_contains"
		ac.ComparisonType = ComparisonTypeContains
		ac.ActualComparison = "String contains check"
		ac.ExpectedBehavior = "String contains check"
	case ">", "<", ">=", "<=":
		if wasNumeric {
			ac.Type = AssertionTypeNumeric
			ac.ComparisonMethod = "numeric"
			ac.ComparisonType = ComparisonTypeNumeric
			ac.ActualComparison = "Numeric comparison performed"
			ac.ExpectedBehavior = "Numeric comparison"
		} else {
			ac.Type = AssertionTypeString
			ac.ComparisonMethod = "string_lexicographic"
			ac.ComparisonType = ComparisonTypeStringFallback
			ac.ActualComparison = "String comparison (numeric parsing failed)"
			if ac.Actual.CanParseNumeric || ac.Expected.CanParseNumeric {
				ac.ExpectedBehavior = "Likely intended numeric comparison"
			} else {
				ac.ExpectedBehavior = "String comparison"
			}
		}
	case "==", "!=":
		if wasNumeric {
			// Keep Type as AssertionTypeComparison, set comparison method details
			ac.ComparisonMethod = "numeric"
			ac.ComparisonType = ComparisonTypeNumeric
			ac.ActualComparison = "Numeric equality performed"
			ac.ExpectedBehavior = "Numeric equality"
		} else {
			// Keep Type as AssertionTypeComparison, set comparison method details
			ac.ComparisonMethod = "string_exact"
			ac.ComparisonType = ComparisonTypeString
			ac.ActualComparison = "String equality performed"
			if ac.Actual.CanParseNumeric && ac.Expected.CanParseNumeric {
				ac.ExpectedBehavior = "Likely intended numeric equality"
			} else {
				ac.ExpectedBehavior = "String equality"
			}
		}
	default:
		ac.ComparisonMethod = "unknown"
		ac.ComparisonType = ComparisonTypeUnknown
	}
	
	// Detect mixed types
	if ac.Actual.OriginalType != ac.Expected.OriginalType {
		ac.Type = AssertionTypeMixed
	}
}

// GetTypeMismatchInfo returns information about type mismatches
func (ac *AssertionContext) GetTypeMismatchInfo() *TypeMismatchInfo {
	if ac.Actual == nil || ac.Expected == nil {
		return nil
	}

	actualType := ac.Actual.Type
	expectedType := ac.Expected.Type
	
	if actualType == expectedType {
		return nil
	}

	return &TypeMismatchInfo{
		ActualType:   actualType,
		ExpectedType: expectedType,
		Suggestion:   ac.generateTypeMismatchSuggestion(),
	}
}

// TypeMismatchInfo provides details about type mismatches in assertions
type TypeMismatchInfo struct {
	ActualType   string `json:"actual_type"`
	ExpectedType string `json:"expected_type"`
	Suggestion   string `json:"suggestion"`
}

// generateTypeMismatchSuggestion provides helpful suggestions for type mismatches
func (ac *AssertionContext) generateTypeMismatchSuggestion() string {
	if ac.Actual == nil || ac.Expected == nil {
		return ""
	}

	actualType := ac.Actual.Type
	expectedType := ac.Expected.Type

	// Common type conversion suggestions
	suggestions := map[string]map[string]string{
		"string": {
			"int":     "Consider converting the string to an integer with strconv.Atoi() or use string comparison",
			"float64": "Consider converting the string to a float with strconv.ParseFloat() or use string comparison",
			"bool":    "Consider converting the string to a boolean with strconv.ParseBool() or use string comparison",
		},
		"int": {
			"string":  "Consider converting the integer to a string with fmt.Sprintf() or strconv.Itoa()",
			"float64": "Consider using float64 for both values to ensure consistent numeric comparison",
		},
		"float64": {
			"string": "Consider converting the float to a string with fmt.Sprintf() or use numeric comparison",
			"int":    "Consider using float64 for both values to ensure consistent numeric comparison",
		},
	}

	if typeMap, exists := suggestions[actualType]; exists {
		if suggestion, exists := typeMap[expectedType]; exists {
			return suggestion
		}
	}

	return fmt.Sprintf("Types %s and %s are incompatible. Consider converting one to match the other.", actualType, expectedType)
}

// ToMap converts AssertionContext to a map for error context
func (ac *AssertionContext) ToMap() map[string]any {
	result := map[string]any{
		"assertion_type":     string(ac.Type),
		"timestamp":          ac.Timestamp,
	}

	if ac.Operator != "" {
		result["operator"] = ac.Operator
	}

	if ac.ComparisonMethod != "" {
		result["comparison_method"] = ac.ComparisonMethod
	}

	if ac.ComparisonType != "" {
		result["comparison_type"] = string(ac.ComparisonType)
	}

	if ac.ActualComparison != "" {
		result["actual_comparison"] = ac.ActualComparison
	}

	if ac.ExpectedBehavior != "" {
		result["expected_behavior"] = ac.ExpectedBehavior
	}

	if ac.Message != "" {
		result["custom_message"] = ac.Message
	}

	if ac.Actual != nil {
		result["actual_value"] = ac.Actual.Value
		result["actual_type"] = ac.Actual.Type
		result["actual_original_type"] = ac.Actual.OriginalType
		result["actual_string"] = ac.Actual.StringValue
		result["actual_is_numeric"] = ac.Actual.IsNumeric
		result["actual_is_native_numeric"] = ac.Actual.IsNativeNumeric
		result["actual_can_parse_numeric"] = ac.Actual.CanParseNumeric
		if ac.Actual.NumericValue != nil {
			result["actual_numeric"] = *ac.Actual.NumericValue
		}
	}

	if ac.Expected != nil {
		result["expected_value"] = ac.Expected.Value
		result["expected_type"] = ac.Expected.Type
		result["expected_original_type"] = ac.Expected.OriginalType
		result["expected_string"] = ac.Expected.StringValue
		result["expected_is_numeric"] = ac.Expected.IsNumeric
		result["expected_is_native_numeric"] = ac.Expected.IsNativeNumeric
		result["expected_can_parse_numeric"] = ac.Expected.CanParseNumeric
		if ac.Expected.NumericValue != nil {
			result["expected_numeric"] = *ac.Expected.NumericValue
		}
	}

	if typeMismatch := ac.GetTypeMismatchInfo(); typeMismatch != nil {
		result["type_mismatch"] = map[string]any{
			"actual_type":   typeMismatch.ActualType,
			"expected_type": typeMismatch.ExpectedType,
			"suggestion":    typeMismatch.Suggestion,
		}
	}

	return result
}

// GetDetailedMessage returns a human-readable detailed message about the assertion
func (ac *AssertionContext) GetDetailedMessage() string {
	var parts []string

	switch ac.Type {
	case AssertionTypeBoolean:
		parts = append(parts, fmt.Sprintf("Boolean assertion failed: value is %s (%s)", 
			ac.Actual.StringValue, ac.Actual.Type))

	case AssertionTypeComparison:
		parts = append(parts, fmt.Sprintf("Comparison assertion failed: %s %s %s", 
			ac.Actual.StringValue, ac.Operator, ac.Expected.StringValue))
		
		parts = append(parts, fmt.Sprintf("Types: %s %s %s", 
			ac.Actual.Type, ac.Operator, ac.Expected.Type))
		
		if ac.ActualComparison != "" {
			parts = append(parts, fmt.Sprintf("How compared: %s", ac.ActualComparison))
		}
		
		if ac.ExpectedBehavior != "" && ac.ExpectedBehavior != ac.ActualComparison {
			parts = append(parts, fmt.Sprintf("Expected behavior: %s", ac.ExpectedBehavior))
		}

		if typeMismatch := ac.GetTypeMismatchInfo(); typeMismatch != nil {
			parts = append(parts, fmt.Sprintf("Type mismatch: %s", typeMismatch.Suggestion))
		}

	case AssertionTypeContains:
		parts = append(parts, fmt.Sprintf("Contains assertion failed: '%s' not found in '%s'", 
			ac.Expected.StringValue, ac.Actual.StringValue))

	case AssertionTypeNumeric:
		parts = append(parts, fmt.Sprintf("Numeric assertion failed: %v %s %v", 
			*ac.Actual.NumericValue, ac.Operator, *ac.Expected.NumericValue))
		
		if ac.ActualComparison != "" {
			parts = append(parts, fmt.Sprintf("Comparison type: %s", ac.ActualComparison))
		}

	case AssertionTypeString:
		parts = append(parts, fmt.Sprintf("String assertion failed: '%s' %s '%s'", 
			ac.Actual.StringValue, ac.Operator, ac.Expected.StringValue))
		
		if ac.ActualComparison != "" {
			parts = append(parts, fmt.Sprintf("Comparison type: %s", ac.ActualComparison))
		}

	case AssertionTypeMixed:
		parts = append(parts, fmt.Sprintf("Mixed-type assertion failed: %s (%s) %s %s (%s)", 
			ac.Actual.StringValue, ac.Actual.OriginalType, ac.Operator, 
			ac.Expected.StringValue, ac.Expected.OriginalType))
		
		if ac.ActualComparison != "" {
			parts = append(parts, fmt.Sprintf("How compared: %s", ac.ActualComparison))
		}
		
		if ac.ExpectedBehavior != "" && ac.ExpectedBehavior != ac.ActualComparison {
			parts = append(parts, fmt.Sprintf("Likely intended: %s", ac.ExpectedBehavior))
		}
	}

	if ac.Message != "" {
		parts = append(parts, fmt.Sprintf("Custom message: %s", ac.Message))
	}

	return strings.Join(parts, "; ")
}

// GetSuggestions returns helpful suggestions for fixing failed assertions
func (ac *AssertionContext) GetSuggestions() []string {
	var suggestions []string

	switch ac.Type {
	case AssertionTypeBoolean:
		if ac.Actual.Type != "bool" {
			suggestions = append(suggestions, 
				"Consider using a comparison assertion instead of boolean assertion",
				"Ensure the value evaluates to a boolean (true/false)")
		}

	case AssertionTypeComparison:
		if typeMismatch := ac.GetTypeMismatchInfo(); typeMismatch != nil {
			suggestions = append(suggestions, typeMismatch.Suggestion)
		}

		if ac.ComparisonMethod == "string_lexicographic" && (ac.Actual.IsNumeric || ac.Expected.IsNumeric) {
			suggestions = append(suggestions, 
				"Values appear to be numeric but were compared as strings. Consider ensuring both values are the same type.")
		}

	case AssertionTypeContains:
		suggestions = append(suggestions, 
			"Check if the search term is spelled correctly",
			"Verify that the target string contains the expected substring",
			"Consider case sensitivity in string comparisons")

	case AssertionTypeNumeric:
		suggestions = append(suggestions,
			"Verify the numeric values are correct",
			"Check for floating-point precision issues")

	case AssertionTypeString:
		suggestions = append(suggestions,
			"Check for extra whitespace or special characters",
			"Verify string case sensitivity",
			"Consider using 'contains' operator for partial matches")
	}

	return suggestions
}