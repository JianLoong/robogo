package actions

import (
	"context"
	"fmt"
	"time"
)

// SleepAction pauses execution for a specified duration.
//
// Parameters:
//   - duration: Time to sleep (can be int, float, or string)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Sleep duration as string
//
// Supported Formats:
//   - Integer: Treated as seconds (e.g., 5 -> 5 seconds)
//   - Float: Treated as seconds with sub-second precision (e.g., 0.5 -> 500ms)
//   - String: Go duration format (e.g., "2m30s", "1h", "500ms")
//
// Examples:
//   - Sleep 5 seconds: [5] -> "5s"
//   - Sleep 500ms: [0.5] -> "500ms"
//   - Sleep 2 minutes: ["2m"] -> "2m0s"
//   - Sleep 1 hour 30 minutes: ["1h30m"] -> "1h30m0s"
//
// Use Cases:
//   - Rate limiting API calls
//   - Waiting for external systems
//   - Simulating user delays
//   - Test timing control
//
// Notes:
//   - Duration is validated before execution
//   - Supports precise timing with float values
//   - String format follows Go's time.ParseDuration
//   - Use for realistic test scenarios
func SleepAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("sleep action requires duration argument")
	}

	var duration time.Duration
	var err error

	switch v := args[0].(type) {
	case int:
		duration = time.Duration(v) * time.Second
	case float64:
		duration = time.Duration(v * float64(time.Second))
	case string:
		duration, err = time.ParseDuration(v)
		if err != nil {
			return "", fmt.Errorf("invalid duration string: %w", err)
		}
	default:
		return "", fmt.Errorf("invalid duration type: %T", v)
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("Sleeping for %v\n", duration)
	}

	time.Sleep(duration)
	return duration.String(), nil
}
