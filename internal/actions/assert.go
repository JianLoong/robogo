package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

func assertAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("assert", 1, len(args))
	}

	// Check for unresolved variables in any argument
	if errorResult := validateArgsResolved("assert", args); errorResult != nil {
		return *errorResult
	}

	// Handle single boolean argument
	if len(args) == 1 {
		if b, ok := args[0].(bool); ok && b {
			return types.ActionResult{
				Status: constants.ActionStatusPassed,
			}
		}

		// Use simple failure function for boolean assertion failure
		return types.BooleanAssertionFailure(args[0])
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
		case constants.OperatorEqual:
			result = actualStr == expectedStr
		case constants.OperatorNotEqual:
			result = actualStr != expectedStr
		case constants.OperatorGreaterThan:
			result, _ = compareNumericWithContext(actualStr, expectedStr, constants.OperatorGreaterThan)
		case constants.OperatorLessThan:
			result, _ = compareNumericWithContext(actualStr, expectedStr, constants.OperatorLessThan)
		case constants.OperatorGreaterThanOrEqual:
			result, _ = compareNumericWithContext(actualStr, expectedStr, constants.OperatorGreaterThanOrEqual)
		case constants.OperatorLessThanOrEqual:
			result, _ = compareNumericWithContext(actualStr, expectedStr, constants.OperatorLessThanOrEqual)
		case constants.OperatorContains:
			result = strings.Contains(actualStr, expectedStr)
		default:
			return types.InvalidArgError("assert", "operator", "valid comparison operator (==, !=, >, <, >=, <=, contains)")
		}

		if result {
			return types.ActionResult{
				Status: constants.ActionStatusPassed,
			}
		}

		// Use simple failure function for comparison assertion failure
		return types.AssertionFailure(expected, actual, fmt.Sprintf("%v", operator))
	}

	// Fallback case - treat as boolean assertion
	return types.BooleanAssertionFailure(args[0])
}

// compareNumericWithContext compares two strings numerically if possible, falling back to string comparison.
// Returns the comparison result and whether numeric comparison was used.
func compareNumericWithContext(actual, expected, operator string) (bool, bool) {
	actualNum, actualErr := strconv.ParseFloat(actual, 64)
	expectedNum, expectedErr := strconv.ParseFloat(expected, 64)

	if actualErr != nil || expectedErr != nil {
		// Fall back to string comparison if not numeric
		result := false
		switch operator {
		case constants.OperatorGreaterThan:
			result = actual > expected
		case constants.OperatorLessThan:
			result = actual < expected
		case constants.OperatorGreaterThanOrEqual:
			result = actual >= expected
		case constants.OperatorLessThanOrEqual:
			result = actual <= expected
		}
		return result, false // false = string comparison was used
	}

	// Numeric comparison
	result := false
	switch operator {
	case constants.OperatorGreaterThan:
		result = actualNum > expectedNum
	case constants.OperatorLessThan:
		result = actualNum < expectedNum
	case constants.OperatorGreaterThanOrEqual:
		result = actualNum >= expectedNum
	case constants.OperatorLessThanOrEqual:
		result = actualNum <= expectedNum
	}
	return result, true // true = numeric comparison was used
}
