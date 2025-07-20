package actions

import (
	"encoding/json"
	"fmt"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// jsonParseAction parses a JSON string into structured data
func jsonParseAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("json_parse", 1, len(args))
	}

	// Get the JSON string to parse
	jsonStr, ok := args[0].(string)
	if !ok {
		return types.InvalidArgError("json_parse", "JSON string", "first argument must be a string")
	}

	// Parse the JSON string
	var parsedData any
	if err := json.Unmarshal([]byte(jsonStr), &parsedData); err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "JSON_PARSE_ERROR").
			WithTemplate("Failed to parse JSON: %s").
			Build(err.Error())
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   parsedData,
	}
}

// jsonBuildAction creates a JSON string from nested YAML arguments
func jsonBuildAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	// The args slice should contain the JSON data structure
	// For json_build, we expect all args to be the JSON structure

	var jsonData any

	// If we have exactly one argument, use it as the JSON data
	if len(args) == 1 {
		jsonData = args[0]
	} else if len(args) == 0 {
		// No args, build from options if provided
		if len(options) > 0 {
			jsonData = options
		} else {
			return types.InvalidArgError("json_build", "JSON data", "at least one argument or options")
		}
	} else {
		// Multiple args - treat as an array
		jsonData = args
	}

	// Perform variable substitution on the data structure
	substitutedData := substituteVariablesInData(jsonData, vars)

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   substitutedData, // Return the parsed data directly
	}
}

// substituteVariablesInData recursively substitutes variables in nested data structures
func substituteVariablesInData(data any, vars *common.Variables) any {
	switch v := data.(type) {
	case string:
		// Substitute variables in strings
		return vars.Substitute(v)
	case map[string]any:
		// Recursively substitute in map values
		result := make(map[string]any)
		for key, value := range v {
			// Also substitute variables in keys if needed
			substitutedKey := vars.Substitute(key)
			result[substitutedKey] = substituteVariablesInData(value, vars)
		}
		return result
	case []any:
		// Recursively substitute in array elements
		result := make([]any, len(v))
		for i, value := range v {
			result[i] = substituteVariablesInData(value, vars)
		}
		return result
	case map[any]any:
		// Handle generic map (from YAML parsing)
		result := make(map[string]any)
		for key, value := range v {
			keyStr := fmt.Sprintf("%v", key)
			substitutedKey := vars.Substitute(keyStr)
			result[substitutedKey] = substituteVariablesInData(value, vars)
		}
		return result
	default:
		// For primitive types (int, float, bool, etc.), return as-is
		return v
	}
}
