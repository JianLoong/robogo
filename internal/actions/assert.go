package actions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func AssertAction(args []interface{}) (string, error) {

	if len(args) < 3 {
		return "", fmt.Errorf("assert action requires at least 3 arguments: value, operator, expected")
	}

	value := args[0]
	operator := fmt.Sprintf("%v", args[1])
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
		return compareNumeric(valueFloat, expectedFloat, operator, msg)
	}

	// Otherwise, use string comparison
	valueStr := fmt.Sprintf("%v", value)
	expectedStr := fmt.Sprintf("%v", expected)
	return compareString(valueStr, expectedStr, operator, msg)
}

func handleModuloOperation(value interface{}, divisor interface{}, operator string, expected interface{}, msg string) (string, error) {
	// Convert value and divisor to float
	valueFloat, valueOk := toFloat(value)
	divisorFloat, divisorOk := toFloat(divisor)
	expectedFloat, expectedOk := toFloat(expected)

	if !valueOk || !divisorOk || !expectedOk {
		return "", fmt.Errorf("modulo operation requires numeric values: value=%v, divisor=%v, expected=%v", value, divisor, expected)
	}

	if divisorFloat == 0 {
		return "", fmt.Errorf("modulo operation: division by zero")
	}

	// Calculate modulo
	result := int(valueFloat) % int(divisorFloat)
	resultFloat := float64(result)

	// Compare result with expected value
	return compareNumeric(resultFloat, expectedFloat, operator, msg)
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func compareNumeric(value, expected float64, operator, msg string) (string, error) {
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
		return "", fmt.Errorf("unsupported numeric operator: %s", operator)
	}

	if !result {
		return "", fmt.Errorf("assertion failed: %v %s %v - %s", value, operator, expected, msg)
	}

	fmt.Printf("✅ %s\n", msg)
	return msg, nil
}

func compareString(value, expected, operator, msg string) (string, error) {
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
			return "", fmt.Errorf("invalid regex pattern '%s': %v", expected, err)
		}
		result = matched
	case "not_matches":
		// Regex pattern not matching
		matched, err := regexp.MatchString(expected, value)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern '%s': %v", expected, err)
		}
		result = !matched
	case "empty":
		result = strings.TrimSpace(value) == ""
	case "not_empty":
		result = strings.TrimSpace(value) != ""
	default:
		return "", fmt.Errorf("unsupported string operator: %s", operator)
	}

	if !result {
		return "", fmt.Errorf("assertion failed: '%v' %s '%v' - %s", value, operator, expected, msg)
	}

	fmt.Printf("✅ %s\n", msg)
	return msg, nil
}
