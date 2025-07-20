package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
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

// Evaluate evaluates a condition string and returns true/false.
// It first substitutes variables, then handles simple boolean values ("true"/"false").
// For more complex conditions, it supports comparison operators:
// - Equality: ==, !=
// - Numeric: >, <, >=, <= (with automatic string fallback)
// - String: contains, starts_with, ends_with
func (evaluator *ConditionEvaluator) Evaluate(condition string) (bool, error) {
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
	operators := []string{
		// Numeric comparison operators (order matters for >= and <= before > and <)
		constants.OperatorGreaterThanOrEqual,
		constants.OperatorLessThanOrEqual,
		constants.OperatorGreaterThan,
		constants.OperatorLessThan,
		// Equality operators
		constants.OperatorEqual,
		constants.OperatorNotEqual,
		// String operators
		constants.OperatorContains,
		constants.OperatorStartsWith,
		constants.OperatorEndsWith,
	}

	for _, op := range operators {
		if strings.Contains(condition, op) {
			parts := strings.Split(condition, op)
			if len(parts) != 2 {
				continue
			}
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			switch op {
			case constants.OperatorEqual:
				return left == right, nil
			case constants.OperatorNotEqual:
				return left != right, nil
			case constants.OperatorContains:
				return strings.Contains(left, right), nil
			case constants.OperatorStartsWith:
				return strings.HasPrefix(left, right), nil
			case constants.OperatorEndsWith:
				return strings.HasSuffix(left, right), nil
			case constants.OperatorGreaterThan, constants.OperatorLessThan, constants.OperatorGreaterThanOrEqual, constants.OperatorLessThanOrEqual:
				// Try numeric comparison
				leftNum, leftErr := strconv.ParseFloat(left, 64)
				rightNum, rightErr := strconv.ParseFloat(right, 64)
				if leftErr == nil && rightErr == nil {
					switch op {
					case constants.OperatorGreaterThan:
						return leftNum > rightNum, nil
					case constants.OperatorLessThan:
						return leftNum < rightNum, nil
					case constants.OperatorGreaterThanOrEqual:
						return leftNum >= rightNum, nil
					case constants.OperatorLessThanOrEqual:
						return leftNum <= rightNum, nil
					}
				}
				// String comparison as fallback
				switch op {
				case constants.OperatorGreaterThan:
					return left > right, nil
				case constants.OperatorLessThan:
					return left < right, nil
				case constants.OperatorGreaterThanOrEqual:
					return left >= right, nil
				case constants.OperatorLessThanOrEqual:
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
