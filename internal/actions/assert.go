package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

func assertAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 3 {
		return types.NewErrorResult("assert action requires 3 arguments: actual, operator, expected")
	}

	actual := fmt.Sprintf("%v", args[0])
	operator := fmt.Sprintf("%v", args[1])
	expected := fmt.Sprintf("%v", args[2])

	// Warn and fail gracefully if actual or expected is unresolved
	if strings.Contains(actual, "__UNRESOLVED__") || strings.Contains(expected, "__UNRESOLVED__") {
		msg := fmt.Sprintf("assertion failed due to unresolved variable: actual=%q, expected=%q", actual, expected)
		fmt.Println("[WARN] " + msg)
		return types.ActionResult{
			Status: types.ActionStatusFailed,
			Error:  msg,
			Output: msg,
		}, nil
	}

	var message string
	if len(args) > 3 {
		message = fmt.Sprintf("%v", args[3])
	}

	var result bool
	switch operator {
	case "==", "=", "equals":
		result = actual == expected
	case "!=", "not_equals":
		result = actual != expected
	case "contains":
		result = strings.Contains(actual, expected)
	case "starts_with":
		result = strings.HasPrefix(actual, expected)
	case "ends_with":
		result = strings.HasSuffix(actual, expected)
	case ">", "<", ">=", "<=":
		if actualNum, err1 := strconv.ParseFloat(actual, 64); err1 == nil {
			if expectedNum, err2 := strconv.ParseFloat(expected, 64); err2 == nil {
				switch operator {
				case ">":
					result = actualNum > expectedNum
				case "<":
					result = actualNum < expectedNum
				case ">=":
					result = actualNum >= expectedNum
				case "<=":
					result = actualNum <= expectedNum
				}
			} else {
				return types.NewErrorResult("cannot compare numeric value with non-numeric: %s", expected)
			}
		} else {
			return types.NewErrorResult("cannot perform numeric comparison with non-numeric value: %s", actual)
		}
	default:
		return types.NewErrorResult("unsupported operator: %s", operator)
	}

	if !result {
		if message != "" {
			msg := fmt.Sprintf("assertion failed: %s (actual: %s, expected: %s)", message, actual, expected)
			return types.ActionResult{
				Status: types.ActionStatusFailed,
				Error:  msg,
				Output: msg,
			}, nil
		}
		msg := fmt.Sprintf("assertion failed: %s %s %s", actual, operator, expected)
		return types.ActionResult{
			Status: types.ActionStatusFailed,
			Error:  msg,
			Output: msg,
		}, nil
	}

	if message != "" {
		msg := fmt.Sprintf("Success: %s", message)
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   msg,
			Output: msg,
		}, nil
	}
	msg := "Assertion passed"
	return types.ActionResult{
		Status: types.ActionStatusPassed,
		Data:   msg,
		Output: msg,
	}, nil
}
