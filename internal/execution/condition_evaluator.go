package execution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
)

// BasicConditionEvaluator implements ConditionEvaluator interface
type BasicConditionEvaluator struct {
	variables *common.Variables
}

// NewBasicConditionEvaluator creates a new basic condition evaluator
func NewBasicConditionEvaluator(variables *common.Variables) *BasicConditionEvaluator {
	return &BasicConditionEvaluator{
		variables: variables,
	}
}

// Evaluate evaluates a condition string and returns true/false.
func (evaluator *BasicConditionEvaluator) Evaluate(condition string) (bool, error) {
	// Substitute variables first
	condition = evaluator.variables.Substitute(condition)

	// Handle simple boolean values
	if condition == "true" {
		return true, nil
	}
	if condition == "false" {
		return false, nil
	}

	// Handle comparison operators
	operators := []string{">=", "<=", ">", "<", "==", "!=", "contains", "starts_with", "ends_with"}

	for _, op := range operators {
		if strings.Contains(condition, op) {
			return evaluator.evaluateComparison(condition, op)
		}
	}

	// If no operators found, treat non-empty strings as true
	return strings.TrimSpace(condition) != "" && strings.TrimSpace(condition) != "0", nil
}

// evaluateComparison evaluates a comparison expression
func (evaluator *BasicConditionEvaluator) evaluateComparison(condition, operator string) (bool, error) {
	parts := strings.SplitN(condition, operator, 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid comparison: %s", condition)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	switch operator {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "contains":
		return strings.Contains(left, right), nil
	case "starts_with":
		return strings.HasPrefix(left, right), nil
	case "ends_with":
		return strings.HasSuffix(left, right), nil
	case ">", "<", ">=", "<=":
		return evaluator.compareNumeric(left, right, operator)
	}

	return false, fmt.Errorf("unsupported operator: %s", operator)
}

// compareNumeric compares two values numerically
func (evaluator *BasicConditionEvaluator) compareNumeric(left, right, operator string) (bool, error) {
	leftNum, err1 := strconv.ParseFloat(left, 64)
	rightNum, err2 := strconv.ParseFloat(right, 64)

	if err1 != nil || err2 != nil {
		// Fall back to string comparison
		switch operator {
		case ">":
			return left > right, nil
		case "<":
			return left < right, nil
		case ">=":
			return left >= right, nil
		case "<=":
			return left <= right, nil
		}
	}

	switch operator {
	case ">":
		return leftNum > rightNum, nil
	case "<":
		return leftNum < rightNum, nil
	case ">=":
		return leftNum >= rightNum, nil
	case "<=":
		return leftNum <= rightNum, nil
	}

	return false, fmt.Errorf("invalid numeric operator: %s", operator)
}