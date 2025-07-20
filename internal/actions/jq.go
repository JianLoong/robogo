package actions

import (
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// jqAction executes jq queries on JSON data
func jqAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 2 {
		return types.MissingArgsError("jq", 2, len(args))
	}

	data := args[0]
	queryStr := fmt.Sprintf("%v", args[1])

	// Parse jq query
	query, err := gojq.Parse(queryStr)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "JQ_PARSE_ERROR").
			WithTemplate("Invalid jq query: %s\nQuery was: %s").
			WithContext("query", queryStr).
			WithSuggestion("Check quote escaping in YAML").
			WithSuggestion("Consider using YAML literal syntax: |\n  your jq query").
			Build(err.Error(), queryStr)
	}

	// Execute query - gojq doesn't support variables the same way
	// For now, we'll implement basic execution without variable support
	// Variables can be handled by embedding them in the data structure
	iter := query.Run(data)
	var results []any

	for {
		result, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := result.(error); ok {
			return types.NewErrorBuilder(types.ErrorCategoryExecution, "JQ_EXEC_ERROR").
				WithTemplate("jq execution failed: %s\nQuery: %s").
				WithContext("query", queryStr).
				WithContext("data_type", fmt.Sprintf("%T", data)).
				WithSuggestion("Verify the data structure matches your jq query").
				WithSuggestion("Use 'jq -n' syntax for testing queries").
				Build(err.Error(), queryStr)
		}
		results = append(results, result)
	}

	// Return appropriate result format
	switch len(results) {
	case 0:
		// No results - return null
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   nil,
		}
	case 1:
		// Single result - return it directly
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   results[0],
		}
	default:
		// Multiple results - return as array
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   results,
		}
	}
}