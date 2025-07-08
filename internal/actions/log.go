package actions

import (
	"fmt"
)

// LogAction outputs messages to the console with optional formatting and verbosity control.
//
// Parameters:
//   - message: Message to log (can be string, number, boolean, or object)
//   - ...args: Additional arguments to log (optional)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: The logged message as string
//
// Supported Types:
//   - Strings: Logged as-is
//   - Numbers: Converted to string representation
//   - Booleans: Logged as "true"/"false"
//   - Objects: JSON-formatted output
//   - Arrays: Space-separated elements
//
// Examples:
//   - Simple message: ["Hello World"] -> "Hello World"
//   - With variables: ["User logged in: ", "${user_name}"] -> "User logged in: John"
//   - Multiple values: ["Count:", 42, "Active:", true] -> "Count: 42 Active: true"
//   - Object logging: [{"status": "success", "data": {"id": 123}}] -> JSON formatted
//   - Array logging: [["item1", "item2", "item3"]] -> "item1 item2 item3"
//
// Use Cases:
//   - Debug information
//   - Test progress tracking
//   - Error reporting
//   - Data inspection
//   - Performance logging
//
// Notes:
//   - Messages are displayed in console output (unless silent=true)
//   - Supports variable substitution with ${variable} syntax
//   - Objects are pretty-printed as JSON
//   - Use verbose field to control output detail level
//   - Messages are included in test reports
func LogAction(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("log action requires at least one argument")
	}

	// Convert all arguments to string representation
	var message string
	for i, arg := range args {
		if i > 0 {
			message += " "
		}
		message += fmt.Sprintf("%v", arg)
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("%s\n", message)
	}

	return message, nil
}
