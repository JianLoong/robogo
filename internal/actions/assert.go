package actions

import (
	"fmt"
	"strconv"
	"strings"
	
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

func assertAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.NewErrorResult("assert action requires at least 1 argument")
	}
	
	// Handle single boolean argument
	if len(args) == 1 {
		if b, ok := args[0].(bool); ok && b {
			return types.ActionResult{
				Status: types.ActionStatusPassed,
			}
		}
		return types.NewErrorResult("assertion failed: %v", args[0])
	}
	
	// Handle comparison syntax: [value, operator, expected]
	if len(args) >= 3 {
		actual := args[0]
		operator := args[1]
		expected := args[2]
		
		// Convert to strings for comparison
		actualStr := fmt.Sprintf("%v", actual)
		expectedStr := fmt.Sprintf("%v", expected)
		
		var result bool
		switch operator {
		case constants.OperatorEqual:
			result = actualStr == expectedStr
		case constants.OperatorNotEqual:
			result = actualStr != expectedStr
		case constants.OperatorGreaterThan:
			result = compareNumeric(actualStr, expectedStr, constants.OperatorGreaterThan)
		case constants.OperatorLessThan:
			result = compareNumeric(actualStr, expectedStr, constants.OperatorLessThan)
		case constants.OperatorGreaterThanOrEqual:
			result = compareNumeric(actualStr, expectedStr, constants.OperatorGreaterThanOrEqual)
		case constants.OperatorLessThanOrEqual:
			result = compareNumeric(actualStr, expectedStr, constants.OperatorLessThanOrEqual)
		case constants.OperatorContains:
			result = strings.Contains(actualStr, expectedStr)
		default:
			return types.NewErrorResult("unsupported operator: %v", operator)
		}
		
		if result {
			return types.ActionResult{
				Status: types.ActionStatusPassed,
			}
		}
		
		message := fmt.Sprintf("assertion failed: %v %v %v", actual, operator, expected)
		if len(args) > 3 {
			message = fmt.Sprintf("%v (%v)", message, args[3])
		}
		return types.NewErrorResult(message)
	}
	
	return types.NewErrorResult("assertion failed: %v", args[0])
}

// compareNumeric compares two strings numerically if possible, falling back to string comparison.
// It first attempts to parse both values as floating-point numbers. If successful, it performs
// numeric comparison. If either value cannot be parsed as a number, it falls back to string
// comparison using Go's string comparison operators.
func compareNumeric(actual, expected, operator string) bool {
	actualNum, actualErr := strconv.ParseFloat(actual, 64)
	expectedNum, expectedErr := strconv.ParseFloat(expected, 64)
	
	if actualErr != nil || expectedErr != nil {
		// Fall back to string comparison if not numeric
		switch operator {
		case constants.OperatorGreaterThan:
			return actual > expected
		case constants.OperatorLessThan:
			return actual < expected
		case constants.OperatorGreaterThanOrEqual:
			return actual >= expected
		case constants.OperatorLessThanOrEqual:
			return actual <= expected
		}
		return false
	}
	
	switch operator {
	case constants.OperatorGreaterThan:
		return actualNum > expectedNum
	case constants.OperatorLessThan:
		return actualNum < expectedNum
	case constants.OperatorGreaterThanOrEqual:
		return actualNum >= expectedNum
	case constants.OperatorLessThanOrEqual:
		return actualNum <= expectedNum
	}
	return false
}
