package actions

import (
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// stringReplaceAction replaces occurrences of a substring in a string
// Args: [text, old, new] - text to search, old substring, new substring
func stringReplaceAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 3 {
		return types.MissingArgsError("string_replace", 3, len(args))
	}

	text := fmt.Sprintf("%v", args[0])
	oldStr := fmt.Sprintf("%v", args[1])
	newStr := fmt.Sprintf("%v", args[2])

	// Get replacement count (default: all occurrences)
	count := -1 // -1 means replace all
	if countOpt, ok := options["count"]; ok {
		if countInt, parseErr := fmt.Sscanf(fmt.Sprintf("%v", countOpt), "%d", &count); parseErr != nil || countInt != 1 {
			return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_COUNT").
				WithTemplate("Invalid count option for string replacement").
				WithContext("count", countOpt).
				WithSuggestion("Use a positive integer or -1 for all occurrences").
				Build(fmt.Sprintf("invalid count: %v", countOpt))
		}
	}

	var result string
	if count == -1 {
		result = strings.ReplaceAll(text, oldStr, newStr)
	} else {
		result = strings.Replace(text, oldStr, newStr, count)
	}

	// Count actual replacements made
	originalCount := strings.Count(text, oldStr)
	actualReplacements := originalCount
	if count != -1 && count < originalCount {
		actualReplacements = count
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"result":            result,
			"original_text":     text,
			"replacements_made": actualReplacements,
			"total_occurrences": originalCount,
		},
	}
}

// stringFormatAction formats a string with placeholders
// Args: [template, ...values] - template string with {} placeholders and values
func stringFormatAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("string_format", 1, len(args))
	}

	template := fmt.Sprintf("%v", args[0])
	values := args[1:]

	// Count placeholders in template
	placeholderCount := strings.Count(template, "{}")

	if len(values) != placeholderCount {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "PLACEHOLDER_MISMATCH").
			WithTemplate("Number of values doesn't match placeholders in template").
			WithContext("template", template).
			WithContext("placeholders", placeholderCount).
			WithContext("values_provided", len(values)).
			WithSuggestion("Ensure the number of values matches the number of {} placeholders").
			Build(fmt.Sprintf("expected %d values for %d placeholders, got %d", placeholderCount, placeholderCount, len(values)))
	}

	// Replace placeholders with values
	result := template
	for _, value := range values {
		valueStr := fmt.Sprintf("%v", value)
		result = strings.Replace(result, "{}", valueStr, 1)
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"result":      result,
			"template":    template,
			"values_used": len(values),
		},
	}
}

// stringAction converts a value to a string
// Args: [value] - value to convert to string
func stringAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("string", 1, len(args))
	}

	// Convert value to string
	value := fmt.Sprintf("%v", args[0])

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"value":         value,
			"original_type": fmt.Sprintf("%T", args[0]),
			"length":        len(value),
		},
	}
}