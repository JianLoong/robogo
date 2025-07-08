package actions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/JianLoong/robogo/internal/util"
)

// AssertAction validates conditions using various comparison operators and returns detailed results.
//
// Parameters:
//   - actual: Actual value to compare
//   - operator: Comparison operator (==, !=, >, <, >=, <=, contains, starts_with, ends_with)
//   - expected: Expected value to compare against
//   - message: Optional custom error message
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Success message or detailed error with actual vs expected values
//
// Supported Operators:
//   - Equality: ==, != (works with strings, numbers, booleans)
//   - Numeric: >, <, >=, <= (numeric comparison)
//   - String: contains, starts_with, ends_with (string operations)
//   - Boolean: true, false (boolean validation)
//
// Examples:
//   - String equality: ["response", "==", "success", "Response should be success"]
//   - Numeric comparison: ["count", ">", "0", "Count should be positive"]
//   - String contains: ["body", "contains", "error", "Body should contain error"]
//   - Boolean check: ["is_valid", "==", "true", "Should be valid"]
//   - Numeric range: ["age", ">=", "18", "Age should be 18 or older"]
//
// Error Messages:
//   - Custom messages are displayed on failure
//   - Default messages show actual vs expected values
//   - Includes operator used in comparison
//
// Notes:
//   - Supports automatic type conversion for numeric comparisons
//   - String operations are case-sensitive
//   - Boolean values can be strings ("true"/"false") or actual booleans
//   - Use continue_on_failure to prevent test termination on assertion failure
func AssertAction(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	parser := util.NewArgParser(args, options)

	if err := parser.RequireMinArgs(3); err != nil {
		return nil, util.NewExecutionError("assert action requires at least 3 arguments: value, operator, expected", err, "assert")
	}

	value := args[0]
	operator, err := parser.GetString(1)
	if err != nil {
		return nil, util.NewExecutionError("operator must be a string", err, "assert")
	}
	expected := args[2]

	// Get custom message if provided
	msg := "Assertion passed"
	if len(args) > 3 {
		msg = fmt.Sprintf("%v", args[3])
	}

	// Handle modulo operation (value % divisor operator expected)
	if operator == "%" && len(args) >= 5 {
		return handleModuloOperation(value, args[2], fmt.Sprintf("%v", args[3]), args[4], msg)
	}

	// Convert values to comparable types
	valueFloat, valueOk := toFloat(value)
	expectedFloat, expectedOk := toFloat(expected)

	// If both values can be converted to float, use numeric comparison
	if valueOk && expectedOk {
		return compareNumeric(valueFloat, expectedFloat, operator, msg, silent)
	}

	// Otherwise, use string comparison
	valueStr := fmt.Sprintf("%v", value)
	expectedStr := fmt.Sprintf("%v", expected)
	return compareString(valueStr, expectedStr, operator, msg, silent)
}

func handleModuloOperation(value interface{}, divisor interface{}, operator string, expected interface{}, msg string) (interface{}, error) {
	// Convert value and divisor to float
	valueFloat, valueOk := toFloat(value)
	divisorFloat, divisorOk := toFloat(divisor)
	expectedFloat, expectedOk := toFloat(expected)

	if !valueOk || !divisorOk || !expectedOk {
		return nil, fmt.Errorf("modulo operation requires numeric values: value=%v, divisor=%v, expected=%v", value, divisor, expected)
	}

	if divisorFloat == 0 {
		return nil, fmt.Errorf("modulo operation: division by zero")
	}

	// Calculate modulo
	result := int(valueFloat) % int(divisorFloat)
	resultFloat := float64(result)

	// Compare result with expected value
	return compareNumeric(resultFloat, expectedFloat, operator, msg, false)
}

func toFloat(v interface{}) (float64, bool) {
	f, err := util.ToFloat(v)
	return f, err == nil
}

func compareNumeric(value, expected float64, operator, msg string, silent bool) (interface{}, error) {
	var result bool

	switch operator {
	case "==", "=":
		result = value == expected
	case "!=", "<>":
		result = value != expected
	case ">":
		result = value > expected
	case "<":
		result = value < expected
	case ">=":
		result = value >= expected
	case "<=":
		result = value <= expected
	default:
		return nil, fmt.Errorf("unsupported numeric operator: %s", operator)
	}

	if !result {
		// Create a clearer error message that doesn't include the success message
		fullMsg := fmt.Sprintf("Assertion failed: %v %s %v", value, operator, expected)
		if msg != "Assertion passed" {
			// Only include custom message if it's not the default
			fullMsg += fmt.Sprintf(" (%s)", msg)
		}
		if !silent {
			fmt.Printf("Failed: %s\n", fullMsg)
		}
		return nil, util.NewAssertionError(fullMsg, value, expected, operator)
	}

	msg = fmt.Sprintf("Success: %s", msg)
	if !silent {
		fmt.Printf("%s\n", msg)
	}
	return msg, nil
}

func compareString(value, expected, operator, msg string, silent bool) (interface{}, error) {
	var result bool

	switch operator {
	case "==", "=":
		result = value == expected
	case "!=", "<>":
		result = value != expected
	case ">":
		result = value > expected
	case "<":
		result = value < expected
	case ">=":
		result = value >= expected
	case "<=":
		result = value <= expected
	case "contains":
		result = strings.Contains(value, expected)
	case "not_contains":
		result = !strings.Contains(value, expected)
	case "starts_with":
		result = strings.HasPrefix(value, expected)
	case "ends_with":
		result = strings.HasSuffix(value, expected)
	case "matches":
		// Regex pattern matching
		matched, err := regexp.MatchString(expected, value)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern '%s': %v", expected, err)
		}
		result = matched
	case "not_matches":
		// Regex pattern not matching
		matched, err := regexp.MatchString(expected, value)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern '%s': %v", expected, err)
		}
		result = !matched
	case "empty":
		result = strings.TrimSpace(value) == ""
	case "not_empty":
		result = strings.TrimSpace(value) != ""
	default:
		return nil, fmt.Errorf("unsupported string operator: %s", operator)
	}

	if !result {
		// Create a clearer error message that doesn't include the success message
		fullMsg := fmt.Sprintf("Assertion failed: '%v' %s '%v'", value, operator, expected)
		if msg != "Assertion passed" {
			// Only include custom message if it's not the default
			fullMsg += fmt.Sprintf(" (%s)", msg)
		}
		if !silent {
			fmt.Printf("Failed: %s\n", fullMsg)
		}
		return nil, util.NewAssertionError(fullMsg, value, expected, operator)
	}

	msg = fmt.Sprintf("Success: %s", msg)
	if !silent {
		fmt.Printf("%s\n", msg)
	}
	return msg, nil
}
