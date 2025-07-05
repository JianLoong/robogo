package actions

import (
	"encoding/json"
	"fmt"
	"strings"
)

// VariableAction manages test variables for dynamic data handling and state management.
//
// Operations:
//   - set: Set a variable value (set, var_name, value)
//   - get: Get a variable value (get, var_name)
//   - list: List all available variables (list)
//
// Parameters:
//   - operation: Variable operation to perform (set, get, list)
//   - var_name: Variable name (for set/get operations)
//   - value: Variable value (for set operation)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: JSON result with operation status and data
//
// Examples:
//   - Set variable: ["set", "api_key", "abc123"]
//   - Set complex value: ["set", "user_data", {"name": "John", "age": 30}]
//   - Get variable: ["get", "user_id"]
//   - List variables: ["list"]
//
// Variable Naming:
//   - Use descriptive names: user_id, api_response, test_data
//   - Supports dot notation: users.admin.name
//   - Case sensitive: UserID != userid
//
// Notes:
//   - Variables persist throughout test execution
//   - Supports complex data types (strings, numbers, objects, arrays)
//   - Use ${variable_name} syntax to reference in other actions
//   - Variables are shared across all steps in a test case
func VariableAction(args []interface{}, silent bool) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("variable action requires at least 1 argument: operation")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))

	switch operation {
	case "set":
		if len(args) < 3 {
			return "", fmt.Errorf("set requires variable_name and value")
		}
		variableName := fmt.Sprintf("%v", args[1])
		return setVariable(variableName, args[2:], silent)
	case "get":
		if len(args) < 2 {
			return "", fmt.Errorf("get requires variable_name")
		}
		variableName := fmt.Sprintf("%v", args[1])
		return getVariable(variableName, silent)
	case "list":
		return listVariables(silent)
	default:
		return "", fmt.Errorf("unknown variable operation: %s", operation)
	}
}

// setVariable sets a variable value
func setVariable(name string, args []interface{}, silent bool) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("set requires a value")
	}

	// Convert the value to a string representation
	var value interface{}
	if len(args) == 1 {
		value = args[0]
	} else {
		// If multiple args, join them as a string
		var parts []string
		for _, arg := range args {
			parts = append(parts, fmt.Sprintf("%v", arg))
		}
		value = strings.Join(parts, " ")
	}

	// Create result object
	result := map[string]interface{}{
		"operation": "set",
		"name":      name,
		"value":     value,
		"status":    "success",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return string(jsonResult), nil
}

// getVariable gets a variable value (placeholder - actual implementation would need access to runner variables)
func getVariable(name string, silent bool) (string, error) {
	result := map[string]interface{}{
		"operation": "get",
		"name":      name,
		"status":    "not_implemented",
		"message":   "get requires access to runner variables - use ${variable_name} syntax instead",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return string(jsonResult), nil
}

// listVariables lists all variables (placeholder - actual implementation would need access to runner variables)
func listVariables(silent bool) (string, error) {
	result := map[string]interface{}{
		"operation": "list",
		"status":    "not_implemented",
		"message":   "list requires access to runner variables - not yet implemented",
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return string(jsonResult), nil
}
