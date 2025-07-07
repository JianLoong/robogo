package actions

import (
	"fmt"
	"time"
)

type ActionFunc func(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)

type ActionExecutor struct {
	registry *ActionRegistry
}

// Global registry instance
var globalRegistry *ActionRegistry

// init initializes the global registry
func init() {
	globalRegistry = NewActionRegistry()
}

// NewActionExecutor creates a new action executor instance.
//
// Returns: ActionExecutor ready to execute Robogo actions
//
// Notes:
//   - Registers all built-in actions
//   - Provides unified interface for action execution
//   - Handles argument validation and error reporting
func NewActionExecutor() *ActionExecutor {
	return &ActionExecutor{}
}

func RabbitMQActionWrapper(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	strArgs := make([]string, len(args))
	for i, v := range args {
		strArgs[i] = fmt.Sprintf("%v", v)
	}
	result, err := RabbitMQAction(strArgs)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func KafkaActionWrapper(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return KafkaAction(args)
}

// GetAction retrieves an action function by name.
//
// Parameters:
//   - action: Action name to retrieve
//
// Returns: Action function and boolean indicating if found
//
// Examples:
//   - Get log action: fn, exists := GetAction("log")
//   - Get assert action: fn, exists := GetAction("assert")
//
// Notes:
//   - Returns nil function if action doesn't exist
//   - Use exists boolean to check if action was found
//   - Useful for dynamic action execution
func GetAction(action string) (ActionFunc, bool) {
	actionObj, exists := globalRegistry.Get(action)
	if !exists {
		return nil, false
	}

	// Convert Action interface back to ActionFunc
	if wrapper, ok := actionObj.(*ActionWrapper); ok {
		return wrapper.fn, true
	}
	return nil, false
}

// Execute executes an action with the provided arguments and options.
func (ae *ActionExecutor) Execute(action string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return globalRegistry.Execute(action, args, options, silent)
}

// parseDuration parses duration from various formats (int, float, string).
//
// Parameters:
//   - value: Duration value in various formats
//
// Returns: Parsed time.Duration and error if any
//
// Examples:
//   - Integer seconds: 5 -> 5s
//   - Float seconds: 0.5 -> 500ms
//   - String duration: "2m30s" -> 2m30s
//
// Supported Formats:
//   - Integer: Treated as seconds
//   - Float: Treated as seconds (supports milliseconds)
//   - String: Go duration format (e.g., "2m30s", "1h", "500ms")
//
// Notes:
//   - Integer values are converted to seconds
//   - Float values support sub-second precision
//   - String format follows Go's time.ParseDuration
func parseDuration(value interface{}) (time.Duration, error) {
	switch v := value.(type) {
	case int:
		return time.Duration(v) * time.Second, nil
	case float64:
		return time.Duration(v * float64(time.Second)), nil
	case string:
		return time.ParseDuration(v)
	default:
		return 0, fmt.Errorf("unsupported duration format: %T", value)
	}
}
