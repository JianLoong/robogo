package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
)

// ConditionEvaluator handles condition evaluation for control flow
type ConditionEvaluator struct {
	variables *common.Variables
}

// NewConditionEvaluator creates a new condition evaluator
func NewConditionEvaluator(variables *common.Variables) *ConditionEvaluator {
	return &ConditionEvaluator{
		variables: variables,
	}
}

// Evaluate evaluates a condition string and returns true/false
func (ce *ConditionEvaluator) Evaluate(condition string) (bool, error) {
	// Substitute variables first
	condition = ce.variables.Substitute(condition)
	
	// Handle simple boolean values
	if condition == "true" {
		return true, nil
	}
	if condition == "false" {
		return false, nil
	}

	// Handle comparison operators
	operators := []string{">=", "<=", "==", "!=", ">", "<", "contains", "starts_with", "ends_with"}
	
	for _, op := range operators {
		if strings.Contains(condition, op) {
			parts := strings.Split(condition, op)
			if len(parts) != 2 {
				continue
			}
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			
			switch op {
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
				// Try numeric comparison
				leftNum, err1 := strconv.ParseFloat(left, 64)
				rightNum, err2 := strconv.ParseFloat(right, 64)
				if err1 == nil && err2 == nil {
					switch op {
					case ">":
						return leftNum > rightNum, nil
					case "<":
						return leftNum < rightNum, nil
					case ">=":
						return leftNum >= rightNum, nil
					case "<=":
						return leftNum <= rightNum, nil
					}
				}
				// String comparison as fallback
				switch op {
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
		}
	}
	
	// If no operator found, try to parse as boolean
	if condition == "1" {
		return true, nil
	}
	if condition == "0" {
		return false, nil
	}
	
	return false, fmt.Errorf("unable to evaluate condition: %s", condition)
}