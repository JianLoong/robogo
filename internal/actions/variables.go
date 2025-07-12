package actions

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/JianLoong/robogo/internal/util"
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
func VariableAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	parser := util.NewArgParser(args, options)

	if err := parser.RequireMinArgs(1); err != nil {
		return nil, util.NewArgumentCountError("variable", 1, len(args))
	}

	operation, err := parser.GetString(0)
	if err != nil {
		return nil, util.NewArgumentTypeError("variable", 0, "string", args[0])
	}

	operation = strings.ToLower(operation)

	switch operation {
	case "set":
		if err := parser.RequireMinArgs(3); err != nil {
			return nil, util.NewArgumentCountError("variable set", 3, len(args))
		}
		variableName, err := parser.GetString(1)
		if err != nil {
			return nil, util.NewArgumentTypeError("variable set", 1, "string", args[1])
		}
		return setVariable(variableName, args[2:], silent)
	case "get":
		if err := parser.RequireMinArgs(2); err != nil {
			return nil, util.NewArgumentCountError("variable get", 2, len(args))
		}
		variableName, err := parser.GetString(1)
		if err != nil {
			return nil, util.NewArgumentTypeError("variable get", 1, "string", args[1])
		}
		return getVariable(variableName, silent)
	case "list":
		return listVariables(silent)
	case "set_variable":
		return nil, fmt.Errorf("'set_variable' is not supported. Use 'set' instead")
	default:
		return nil, fmt.Errorf("unknown variable operation: %s. Supported operations: set, get, list", operation)
	}
}

// setVariable sets a variable value
func setVariable(name string, args []interface{}, silent bool) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("set requires a value")
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

	// NOTE: Variables are managed by the execution context, not by actions
	// The variable action serves as a way to pass variable operations to the execution engine
	// The actual storage will be handled by the step execution service when it processes the result
	
	// Create result object that the execution engine can use
	result := map[string]interface{}{
		"operation": "set",
		"name":      name,
		"value":     value,
		"status":    "success",
		// Special marker to indicate this result should trigger variable setting
		"__robogo_set_variable": map[string]interface{}{
			"name":  name,
			"value": value,
		},
	}

	resultMap, err := util.ConvertToMap(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result to map: %w", err)
	}

	return resultMap, nil
}

// getVariable retrieves a variable value
func getVariable(name string, silent bool) (interface{}, error) {
	// Note: This is a placeholder implementation
	// In a real implementation, this would retrieve from the actual variable store
	// which is managed by the runner/execution context
	result := map[string]interface{}{
		"operation": "get",
		"name":      name,
		"status":    "error",
		"message":   "Variable retrieval not implemented - variables are managed by the execution context",
	}

	resultMap, err := util.ConvertToMap(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result to map: %w", err)
	}

	return resultMap, nil
}

// listVariables lists all available variables
func listVariables(silent bool) (interface{}, error) {
	// Note: This is a placeholder implementation
	// In a real implementation, this would list from the actual variable store
	// which is managed by the runner/execution context
	result := map[string]interface{}{
		"operation": "list",
		"status":    "error",
		"message":   "Variable listing not implemented - variables are managed by the execution context",
		"variables": []interface{}{},
	}

	resultMap, err := util.ConvertToMap(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result to map: %w", err)
	}

	return resultMap, nil
}

