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
		msg := "assert action requires 3 arguments: actual, operator, expected"
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}

	actual := fmt.Sprintf("%v", args[0])
	operator := fmt.Sprintf("%v", args[1])
	expected := fmt.Sprintf("%v", args[2])

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
				msg := fmt.Sprintf("cannot compare numeric value with non-numeric: %s", expected)
				return types.ActionResult{
					Status: types.ActionStatusError,
					Error:  msg,
					Output: msg,
				}, fmt.Errorf(msg)
			}
		} else {
			msg := fmt.Sprintf("cannot perform numeric comparison with non-numeric value: %s", actual)
			return types.ActionResult{
				Status: types.ActionStatusError,
				Error:  msg,
				Output: msg,
			}, fmt.Errorf(msg)
		}
	default:
		msg := fmt.Sprintf("unsupported operator: %s", operator)
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}

	if !result {
		if message != "" {
			msg := fmt.Sprintf("assertion failed: %s (actual: %s, expected: %s)", message, actual, expected)
			return types.ActionResult{
				Status: types.ActionStatusError,
				Error:  msg,
				Output: msg,
			}, fmt.Errorf(msg)
		}
		msg := fmt.Sprintf("assertion failed: %s %s %s", actual, operator, expected)
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  msg,
			Output: msg,
		}, fmt.Errorf(msg)
	}

	if message != "" {
		msg := fmt.Sprintf("Success: %s", message)
		return types.ActionResult{
			Status: types.ActionStatusSuccess,
			Data:   msg,
			Output: msg,
		}, nil
	}
	msg := "Assertion passed"
	return types.ActionResult{
		Status: types.ActionStatusSuccess,
		Data:   msg,
		Output: msg,
	}, nil
}
