package actions

import (
	"encoding/json"
	"fmt"
	"strings"
)

// VariableAction handles variable-related operations
func VariableAction(args []interface{}) (string, error) {
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
		return setVariable(variableName, args[2:])
	case "get":
		if len(args) < 2 {
			return "", fmt.Errorf("get requires variable_name")
		}
		variableName := fmt.Sprintf("%v", args[1])
		return getVariable(variableName)
	case "list":
		return listVariables()
	default:
		return "", fmt.Errorf("unknown variable operation: %s", operation)
	}
}

// setVariable sets a variable value
func setVariable(name string, args []interface{}) (string, error) {
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
func getVariable(name string) (string, error) {
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
func listVariables() (string, error) {
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
