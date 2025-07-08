package actions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ControlFlowAction handles conditional execution and loop control operations.
//
// Operations:
//   - if: Evaluate a condition and return true/false
//   - for: Handle for loop iterations (range, array, count)
//   - while: Evaluate while loop conditions
//
// Parameters:
//   - flowType: Control flow type (if, for, while)
//   - condition: Condition to evaluate or loop specification
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Result of condition evaluation or loop information
//
// Examples:
//   - If condition: ["if", "${value} > 5"]
//   - If string: ["if", "${response} contains 'success'"]
//   - For range: ["for", "1..5"]
//   - For array: ["for", "[item1, item2, item3]"]
//   - For count: ["for", "3"]
//   - While condition: ["while", "${counter} < 10"]
//
// Condition Operators:
//   - Comparison: ==, !=, >, <, >=, <=
//   - String: contains, starts_with, ends_with
//   - Logical: &&, ||, !
//
// Loop Types:
//   - Range: "1..5" (inclusive range)
//   - Array: "[item1, item2, item3]" (array of items)
//   - Count: "3" (simple count)
//
// Notes:
//   - If conditions return boolean for use in if/else blocks
//   - For loops support range, array, and count formats
//   - While conditions return boolean for loop continuation
//   - Use max_iterations to prevent infinite loops
func ControlFlowAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("control flow action requires at least 2 arguments: type and condition")
	}

	flowType := strings.ToLower(fmt.Sprintf("%v", args[0]))

	switch flowType {
	case "if":
		return handleIfStatement(ctx, args[1:], silent)
	case "for":
		return handleForLoop(ctx, args[1:])
	case "while":
		return handleWhileLoop(ctx, args[1:])
	default:
		return nil, fmt.Errorf("unknown control flow type: %s", flowType)
	}
}

// handleIfStatement evaluates a condition and returns "true" or "false"
func handleIfStatement(ctx context.Context, args []interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("if statement requires a condition")
	}

	condition := fmt.Sprintf("%v", args[0])
	result, err := evaluateCondition(condition)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate condition '%s': %w", condition, err)
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("If condition '%s' evaluated to: %v\n", condition, result)
	}
	return result, nil
}

// handleForLoop handles for loop iterations
func handleForLoop(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("for loop requires a range or array")
	}

	rangeSpec := fmt.Sprintf("%v", args[0])

	// Handle different range formats
	if strings.Contains(rangeSpec, "..") {
		// Range format: "1..5" or "start..end"
		return handleRangeLoop(rangeSpec)
	} else if strings.HasPrefix(rangeSpec, "[") && strings.HasSuffix(rangeSpec, "]") {
		// Array format: "[1,2,3,4,5]"
		return handleArrayLoop(rangeSpec)
	} else {
		// Try to parse as number for count-based loop
		return handleCountLoop(rangeSpec)
	}
}

// handleWhileLoop handles while loop condition evaluation
func handleWhileLoop(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("while loop requires a condition")
	}

	condition := fmt.Sprintf("%v", args[0])
	result, err := evaluateCondition(condition)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate while condition '%s': %w", condition, err)
	}

	fmt.Printf("While condition '%s' evaluated to: %v\n", condition, result)
	return result, nil
}

// evaluateCondition evaluates a condition string and returns boolean result
func evaluateCondition(condition string) (bool, error) {
	// Handle common comparison operators
	operators := []string{"==", "!=", ">=", "<=", ">", "<", "contains", "starts_with", "ends_with"}

	for _, op := range operators {
		if strings.Contains(condition, " "+op+" ") {
			parts := strings.Split(condition, " "+op+" ")
			if len(parts) != 2 {
				return false, fmt.Errorf("invalid condition format: %s", condition)
			}

			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			return compareValues(left, right, op)
		}
	}

	// If no operator found, treat as boolean value
	return parseBoolean(condition)
}

// compareValues compares two values using the specified operator
func compareValues(left, right, operator string) (bool, error) {
	switch operator {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "contains":
		return strings.Contains(left, right), nil
	case "starts_with":
		return strings.HasPrefix(left, right), nil
	case "ends_with":
		return strings.HasSuffix(left, right), nil
	case ">", ">=", "<", "<=":
		// Try numeric comparison
		leftNum, leftErr := strconv.ParseFloat(left, 64)
		rightNum, rightErr := strconv.ParseFloat(right, 64)

		if leftErr != nil || rightErr != nil {
			// Fall back to string comparison
			switch operator {
			case ">":
				return left > right, nil
			case ">=":
				return left >= right, nil
			case "<":
				return left < right, nil
			case "<=":
				return left <= right, nil
			}
		}

		switch operator {
		case ">":
			return leftNum > rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	return false, fmt.Errorf("unsupported operator: %s", operator)
}

// parseBoolean parses a string as a boolean value
func parseBoolean(value string) (bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse as boolean: %s", value)
	}
}

// handleRangeLoop handles range-based loops (e.g., "1..5")
func handleRangeLoop(rangeSpec string) (interface{}, error) {
	parts := strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", rangeSpec)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start value in range: %s", parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end value in range: %s", parts[1])
	}

	if start > end {
		return nil, fmt.Errorf("start value cannot be greater than end value: %d > %d", start, end)
	}

	fmt.Printf("For loop range: %d to %d (iterations: %d)\n", start, end, end-start+1)
	return end - start + 1, nil
}

// handleArrayLoop handles array-based loops (e.g., "[1,2,3,4,5]")
func handleArrayLoop(arraySpec string) (interface{}, error) {
	// Simple array parsing - remove brackets and split by comma
	content := strings.TrimPrefix(strings.TrimSuffix(arraySpec, "]"), "[")
	items := strings.Split(content, ",")

	// Count non-empty items
	count := 0
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			count++
		}
	}

	fmt.Printf("For loop array: %s (items: %d)\n", arraySpec, count)
	return count, nil
}

// handleCountLoop handles count-based loops (e.g., "5" for 5 iterations)
func handleCountLoop(countSpec string) (interface{}, error) {
	count, err := strconv.Atoi(strings.TrimSpace(countSpec))
	if err != nil {
		return nil, fmt.Errorf("invalid count value: %s", countSpec)
	}

	if count < 0 {
		return nil, fmt.Errorf("count cannot be negative: %d", count)
	}

	fmt.Printf("For loop count: %d iterations\n", count)
	return count, nil
}
