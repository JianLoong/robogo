package actions

import (
	"fmt"
	"time"
)

type ActionFunc func([]interface{}, bool) (string, error)

type ActionExecutor struct{}

var actionFuncs = map[string]ActionFunc{
	"log":        LogAction,
	"sleep":      SleepAction,
	"assert":     AssertAction,
	"get_time":   GetTimeAction,
	"get_random": GetRandomAction,
	"concat":     ConcatAction,
	"length":     LengthAction,
	"http":       HTTPAction,
	"http_get":   HTTPGetAction,
	"http_post":  HTTPPostAction,
	"control":    ControlFlowAction,
	"postgres":   PostgresAction,
	"variable":   VariableAction,
	"tdm":        TDMAction,
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
	fn, ok := actionFuncs[action]
	return fn, ok
}

// Execute executes an action with the provided arguments.
//
// Parameters:
//   - action: Action name to execute
//   - args: Array of arguments for the action
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Action result string and error if any
//
// Examples:
//   - Execute log: executor.Execute("log", []interface{}{"Hello World"}, false)
//   - Execute assert: executor.Execute("assert", []interface{}{"value", "==", "expected"}, true)
//
// Notes:
//   - Validates action exists before execution
//   - Passes arguments to action function
//   - Returns error for unknown actions
//   - Silent mode respects verbosity settings across all actions
func (ae *ActionExecutor) Execute(action string, args []interface{}, silent bool) (string, error) {
	if fn, ok := actionFuncs[action]; ok {
		return fn(args, silent)
	}
	return "", fmt.Errorf("unknown action: %s", action)
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
