package runner

import (
	"fmt"
	"strconv"
	"strings"
)

// EvaluateCondition evaluates a simple condition string
// This is a basic implementation - a full implementation would use a proper expression parser
func EvaluateCondition(condition string) (bool, error) {
	// Trim whitespace
	condition = strings.TrimSpace(condition)
	
	// Handle simple boolean values
	if condition == "true" {
		return true, nil
	}
	if condition == "false" {
		return false, nil
	}
	
	// Handle numeric comparisons (basic implementation)
	// Examples: "5 > 3", "10 == 10", "7 <= 8"
	if strings.Contains(condition, " ") {
		return evaluateComparison(condition)
	}
	
	// Handle variable references (would need context)
	// For now, assume unknown conditions are false
	return false, fmt.Errorf("unsupported condition format: %s", condition)
}

// evaluateComparison evaluates basic comparison expressions
func evaluateComparison(condition string) (bool, error) {
	// Simple parsing - in production, use a proper expression parser
	parts := strings.Fields(condition)
	if len(parts) != 3 {
		return false, fmt.Errorf("invalid comparison format: %s", condition)
	}
	
	left := parts[0]
	operator := parts[1]
	right := parts[2]
	
	// Try to parse as numbers first
	leftNum, leftErr := strconv.ParseFloat(left, 64)
	rightNum, rightErr := strconv.ParseFloat(right, 64)
	
	if leftErr == nil && rightErr == nil {
		return evaluateNumericComparison(leftNum, operator, rightNum)
	}
	
	// Fall back to string comparison
	return evaluateStringComparison(left, operator, right)
}

// evaluateNumericComparison evaluates numeric comparisons
func evaluateNumericComparison(left float64, operator string, right float64) (bool, error) {
	switch operator {
	case "==", "=":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "<":
		return left < right, nil
	case "<=":
		return left <= right, nil
	case ">":
		return left > right, nil
	case ">=":
		return left >= right, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", operator)
	}
}

// evaluateStringComparison evaluates string comparisons
func evaluateStringComparison(left, operator, right string) (bool, error) {
	switch operator {
	case "==", "=":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "contains":
		return strings.Contains(left, right), nil
	case "starts_with":
		return strings.HasPrefix(left, right), nil
	case "ends_with":
		return strings.HasSuffix(left, right), nil
	default:
		return false, fmt.Errorf("unsupported string operator: %s", operator)
	}
}