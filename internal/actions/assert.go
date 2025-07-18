package actions

import (
	"fmt"
	"strconv"
	"strings"
	
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

func assertAction(args []any, options map[string]any, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 1 {
		return types.NewErrorResult("assert action requires at least 1 argument")
	}
	
	// Handle single boolean argument
	if len(args) == 1 {
		if b, ok := args[0].(bool); ok && b {
			return types.ActionResult{
				Status: types.ActionStatusPassed,
			}, nil
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
		case "==":
			result = actualStr == expectedStr
		case "!=":
			result = actualStr != expectedStr
		case ">":
			result = compareNumeric(actualStr, expectedStr, ">")
		case "<":
			result = compareNumeric(actualStr, expectedStr, "<")
		case ">=":
			result = compareNumeric(actualStr, expectedStr, ">=")
		case "<=":
			result = compareNumeric(actualStr, expectedStr, "<=")
		case "contains":
			result = strings.Contains(actualStr, expectedStr)
		default:
			return types.NewErrorResult("unsupported operator: %v", operator)
		}
		
		if result {
			return types.ActionResult{
				Status: types.ActionStatusPassed,
			}, nil
		}
		
		message := fmt.Sprintf("assertion failed: %v %v %v", actual, operator, expected)
		if len(args) > 3 {
			message = fmt.Sprintf("%v (%v)", message, args[3])
		}
		return types.NewErrorResult(message)
	}
	
	return types.NewErrorResult("assertion failed: %v", args[0])
}

func compareNumeric(actual, expected, operator string) bool {
	actualNum, err1 := strconv.ParseFloat(actual, 64)
	expectedNum, err2 := strconv.ParseFloat(expected, 64)
	
	if err1 != nil || err2 != nil {
		// Fall back to string comparison if not numeric
		switch operator {
		case ">":
			return actual > expected
		case "<":
			return actual < expected
		case ">=":
			return actual >= expected
		case "<=":
			return actual <= expected
		}
		return false
	}
	
	switch operator {
	case ">":
		return actualNum > expectedNum
	case "<":
		return actualNum < expectedNum
	case ">=":
		return actualNum >= expectedNum
	case "<=":
		return actualNum <= expectedNum
	}
	return false
}
