package execution

import (
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// printStepExecution prints step execution details to console
func (s *BasicExecutionStrategy) printStepExecution(
	step types.Step,
	stepNum int,
	args []any,
	options map[string]any,
) {
	fmt.Printf("Step %d: %s\n", stepNum, step.Name)
	fmt.Printf("  Action: %s\n", step.Action)

	if len(args) > 0 {
		// Args are already masked at this point
		fmt.Printf("  Args: %v\n", args)
	}

	if len(options) > 0 {
		fmt.Printf("  Options: %v\n", options)
	}

	// Show conditions if present
	if step.If != "" {
		condition := s.variables.Substitute(step.If)
		fmt.Printf("  If: %s\n", condition)
	}

	if step.For != "" {
		forValue := s.variables.Substitute(step.For)
		fmt.Printf("  For: %s\n", forValue)
	}

	if step.While != "" {
		whileValue := s.variables.Substitute(step.While)
		fmt.Printf("  While: %s\n", whileValue)
	}

	if step.Result != "" {
		fmt.Printf("  Result Variable: %s\n", step.Result)
	}

	fmt.Println("  Executing... ")
}

// printStepResult prints the result of step execution
func (s *BasicExecutionStrategy) printStepResult(result types.ActionResult, duration time.Duration) {
	// Print status with color-like indicators
	switch result.Status {
	case constants.ActionStatusPassed:
		fmt.Printf("✓ PASSED (%s)\n", duration)
	case constants.ActionStatusFailed:
		fmt.Printf("✗ FAILED (%s)\n", duration)
		if errorMsg := result.GetMessage(); errorMsg != "" {
			fmt.Printf("    Error: %s\n", errorMsg)
		}
	case constants.ActionStatusSkipped:
		fmt.Printf("- SKIPPED (%s)\n", duration)
		if skipReason := result.GetSkipReason(); skipReason != "" {
			fmt.Printf("    Reason: %s\n", skipReason)
		}
	case constants.ActionStatusError:
		fmt.Printf("! ERROR (%s)\n", duration)
		if errorMsg := result.GetMessage(); errorMsg != "" {
			fmt.Printf("    Error: %s\n", errorMsg)
		}
	default:
		fmt.Printf("? %s (%s)\n", result.Status, duration)
	}

	// Show result data if present and not too large
	if result.Data != nil {
		dataStr := fmt.Sprintf("%v", result.Data)
		if len(dataStr) <= 100 { // Only show small data to avoid cluttering output
			fmt.Printf("    Data: %s\n", dataStr)
		} else {
			fmt.Printf("    Data: [%d characters]\n", len(dataStr))
		}
	}

	fmt.Println() // Add blank line for readability
}

// printSecureStepResult prints the result of step execution for no_log steps
// Only shows status and duration, no sensitive data
func (s *BasicExecutionStrategy) printSecureStepResult(result types.ActionResult, duration time.Duration) {
	// Print status with color-like indicators, but no sensitive data
	switch result.Status {
	case constants.ActionStatusPassed:
		fmt.Printf("✓ PASSED (%s) [no sensitive data logged]\n", duration)
	case constants.ActionStatusFailed:
		fmt.Printf("✗ FAILED (%s) [no sensitive data logged]\n", duration)
		// Don't show error message as it might contain sensitive information
		fmt.Printf("    Error details suppressed for security\n")
	case constants.ActionStatusSkipped:
		fmt.Printf("- SKIPPED (%s) [no sensitive data logged]\n", duration)
		fmt.Printf("    Reason details suppressed for security\n")
	case constants.ActionStatusError:
		fmt.Printf("! ERROR (%s) [no sensitive data logged]\n", duration)
		fmt.Printf("    Error details suppressed for security\n")
	default:
		fmt.Printf("? %s (%s) [no sensitive data logged]\n", result.Status, duration)
	}

	// Never show result data for no_log steps
	fmt.Println() // Add blank line for readability
}