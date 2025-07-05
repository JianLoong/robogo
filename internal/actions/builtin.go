package actions

import (
	"fmt"
	"time"
)

type ActionFunc func([]interface{}) (string, error)

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
}

// NewActionExecutor creates a new action executor
func NewActionExecutor() *ActionExecutor {
	return &ActionExecutor{}
}

// Execute executes an action with arguments
func (ae *ActionExecutor) Execute(action string, args []interface{}) (string, error) {
	if fn, ok := actionFuncs[action]; ok {
		return fn(args)
	}
	return "", fmt.Errorf("unknown action: %s", action)
}

// parseDuration parses duration from various formats
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
