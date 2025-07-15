package actions

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
)

// Core actions - assert, log, variable
func assertAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("assert action requires 3 arguments: actual, operator, expected")
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
		// Try numeric comparison first
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
				return nil, fmt.Errorf("cannot compare numeric value with non-numeric: %s", expected)
			}
		} else {
			return nil, fmt.Errorf("cannot perform numeric comparison with non-numeric value: %s", actual)
		}
	default:
		return nil, fmt.Errorf("unsupported operator: %s", operator)
	}

	if !result {
		if message != "" {
			return nil, fmt.Errorf("assertion failed: %s (actual: %s, expected: %s)", message, actual, expected)
		}
		return nil, fmt.Errorf("assertion failed: %s %s %s", actual, operator, expected)
	}

	if message != "" {
		return fmt.Sprintf("Success: %s", message), nil
	}
	return "Assertion passed", nil
}

func logAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("log action requires at least 1 argument")
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%v", arg)
	}

	message := strings.Join(parts, " ")
	fmt.Println(message)
	os.Stdout.Sync() // Flush output immediately

	return message, nil
}

func variableAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("variable action requires at least 2 arguments")
	}

	name := fmt.Sprintf("%v", args[0])
	value := args[1]

	vars.Set(name, value)

	return fmt.Sprintf("Set variable %s = %v", name, value), nil
}