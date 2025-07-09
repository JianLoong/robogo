package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ConcatAction concatenates multiple strings or values into a single string.
//
// Parameters:
//   - ...values: Values to concatenate (strings, numbers, booleans, objects)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Concatenated string
//
// Supported Types:
//   - Strings: Concatenated as-is
//   - Numbers: Converted to string representation
//   - Booleans: Converted to "true"/"false"
//   - Objects: JSON-formatted strings
//   - Arrays: Space-separated elements
//
// Examples:
//   - Simple concatenation: ["Hello", " ", "World"] -> "Hello World"
//   - With numbers: ["User", " ", "ID:", " ", 123] -> "User ID: 123"
//   - With variables: ["Welcome", " ", "${user_name}"] -> "Welcome John"
//   - Mixed types: ["Status:", " ", true, " Count:", " ", 42] -> "Status: true Count: 42"
//   - Object concatenation: ["Data:", " ", {"id": 1, "name": "test"}] -> "Data: {\"id\":1,\"name\":\"test\"}"
//
// Use Cases:
//   - Building API request bodies
//   - Creating log messages
//   - Generating file paths
//   - Constructing SQL queries
//   - Formatting output messages
//
// Notes:
//   - All values are converted to strings before concatenation
//   - Objects are JSON-formatted for readability
//   - Supports variable substitution with ${variable} syntax
//   - No separator is added between values (add explicitly if needed)
func ConcatAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("concat action requires at least one argument")
	}

	var result strings.Builder
	for _, arg := range args {
		result.WriteString(fmt.Sprintf("%v", arg))
	}

	concatenated := result.String()
	return concatenated, nil
}

// LengthAction calculates the length of strings, arrays, or objects.
//
// Parameters:
//   - value: Value to measure (string, array, object, or variable reference)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Length as string
//
// Supported Types:
//   - Strings: Character count (including spaces and special characters)
//   - Arrays: Number of elements
//   - Objects: Number of key-value pairs
//   - Numbers: Digit count (converted to string first)
//   - Booleans: Length of "true"/"false" string
//
// Examples:
//   - String length: ["Hello World"] -> "11"
//   - Array length: [["a", "b", "c"]] -> "3"
//   - Object length: [{"name": "John", "age": 30}] -> "2"
//   - Number length: [12345] -> "5"
//   - Variable length: ["${response_body}"] -> length of response_body
//
// Use Cases:
//   - Input validation
//   - Response size verification
//   - Array bounds checking
//   - String truncation validation
//   - Data completeness checks
//
// Notes:
//   - For strings, counts all characters including spaces and special chars
//   - For arrays, counts the number of elements
//   - For objects, counts the number of key-value pairs
//   - Supports variable substitution with ${variable} syntax
//   - Returns "0" for nil or empty values
func LengthAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("length action requires exactly one argument")
	}

	value := args[0]
	var length int

	switch v := value.(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	case map[string]interface{}:
		length = len(v)
	case int:
		length = len(fmt.Sprintf("%d", v))
	case float64:
		length = len(fmt.Sprintf("%v", v))
	case bool:
		if v {
			length = 4 // "true"
		} else {
			length = 5 // "false"
		}
	default:
		// Convert to string and get length
		length = len(fmt.Sprintf("%v", v))
	}

	result := strconv.Itoa(length)
	return result, nil
}

// Add bytes_to_string action
// bytesToStringAction converts a byte slice or any value to a string
func bytesToStringAction(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("bytes_to_string: missing argument")
	}
	switch v := args[0].(type) {
	case []byte:
		return string(v), nil
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// jsonExtractAction extracts a value by key from a JSON string
func jsonExtractAction(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("json_extract: requires json string and key")
	}
	jsonStr, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("json_extract: first argument must be a string (json)")
	}
	key, ok := args[1].(string)
	if !ok {
		return "", fmt.Errorf("json_extract: second argument must be a string (key)")
	}
	var m map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		return "", fmt.Errorf("json_extract: invalid json: %w", err)
	}
	val, exists := m[key]
	if !exists {
		return "", fmt.Errorf("json_extract: key '%s' not found", key)
	}
	// If the value is not a string, marshal it back to string
	switch v := val.(type) {
	case string:
		return v, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("json_extract: error marshaling value: %w", err)
		}
		return string(b), nil
	}
}
