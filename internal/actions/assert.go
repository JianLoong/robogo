package actions

import (
	"fmt"
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
		return "", fmt.Errorf("unsupported operator: %s", operator)
	}

	if !result {
		return "", fmt.Errorf("assertion failed: %v %s %v", value, operator, expected)
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
	default:
		return "", fmt.Errorf("unsupported operator: %s", operator)
	}

	if !result {
		return "", fmt.Errorf("assertion failed: %v %s %v", value, operator, expected)
	}

	fmt.Printf("✅ %s\n", msg)
	return msg, nil
}
