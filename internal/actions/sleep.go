package actions

import (
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// sleepAction pauses execution for a specified duration
// Args: [duration] - duration string (e.g., "2s", "500ms", "1m30s")
func sleepAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("sleep", 1, len(args))
	}

	durationStr := fmt.Sprintf("%v", args[0])

	// Parse the duration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_DURATION").
			WithTemplate("Invalid duration format for sleep action").
			WithContext("duration", durationStr).
			WithContext("valid_examples", "2s, 500ms, 1m30s, 2h").
			WithSuggestion("Use Go duration format: ns, us, ms, s, m, h").
			WithSuggestion("Examples: '1s', '500ms', '2m', '1h30m'").
			Build(fmt.Sprintf("invalid duration format: %s", durationStr))
	}

	// Validate reasonable duration limits
	if duration < 0 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "NEGATIVE_DURATION").
			WithTemplate("Sleep duration cannot be negative").
			WithContext("duration", durationStr).
			WithSuggestion("Use positive duration values").
			Build(fmt.Sprintf("negative duration not allowed: %s", durationStr))
	}

	// Warn about very long durations (over 5 minutes)
	if duration > 5*time.Minute {
		fmt.Printf("⚠️  Warning: Long sleep duration detected (%s). This may slow down your tests significantly.\n", duration)
	}

	// Perform the sleep
	fmt.Printf("💤 Sleeping for %s...\n", duration)
	time.Sleep(duration)
	fmt.Printf("✅ Sleep completed (%s)\n", duration)

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"duration":        durationStr,
			"duration_parsed": duration.String(),
			"duration_ms":     duration.Milliseconds(),
		},
	}
}
