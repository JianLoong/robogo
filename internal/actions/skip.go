package actions

import (
	"context"
	"fmt"
)

// SkipAction allows test cases to be explicitly skipped with a reason.
//
// Parameters:
//   - reason: Reason for skipping the test case (optional)
//   - silent: Whether to suppress output (respects verbosity settings)
//
// Returns: Skip reason as string
//
// Examples:
//   - Simple skip: ["Skipping this test"]
//   - Conditional skip: ["${environment} == 'prod'", "Skipping in production"]
//   - No reason: []
//
// Use Cases:
//   - Skip tests that are not applicable to current environment
//   - Skip tests that depend on external services that are down
//   - Skip tests that are known to be flaky
//   - Skip tests that are not yet implemented
//
// Notes:
//   - When this action is executed, the test case is marked as "skipped"
//   - The test case execution stops after this action
//   - The reason is included in test reports
//   - Use with conditional logic for dynamic skipping
func SkipAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	reason := "Test case skipped"
	if len(args) > 0 {
		reason = fmt.Sprintf("%v", args[0])
	}

	// Only print if not silent
	if !silent {
		fmt.Printf("%s\n", reason)
	}

	// Return a special error that indicates this is a skip
	return reason, &SkipError{Reason: reason}
}

// SkipError is a special error type that indicates a test case should be skipped
type SkipError struct {
	Reason string
}

func (e *SkipError) Error() string {
	return fmt.Sprintf("SKIP: %s", e.Reason)
}

// IsSkipError checks if an error is a skip error
func IsSkipError(err error) bool {
	_, ok := err.(*SkipError)
	return ok
}
