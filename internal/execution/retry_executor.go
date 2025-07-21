package execution

import (
	"fmt"
	"math"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// RetryExecutor handles retry logic for step execution
type RetryExecutor struct {
	actionExecutor ActionExecutor
	variables      *common.Variables
}

// Removed - now using unified interfaces.StepExecutor

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(actionExecutor ActionExecutor, variables *common.Variables) *RetryExecutor {
	return &RetryExecutor{
		actionExecutor: actionExecutor,
		variables:      variables,
	}
}

// ExecuteStepWithRetry executes a step with retry logic
func (executor *RetryExecutor) ExecuteStepWithRetry(
	step types.Step,
	stepNum int,
	loopCtx *types.LoopContext,
) (*types.StepResult, error) {
	var lastResult *types.StepResult
	var lastErr error

	// Default configuration
	maxAttempts := step.Retry.Attempts
	if maxAttempts <= 0 {
		maxAttempts = 1 // At least try once
	}

	// Parse delay duration
	baseDelay := time.Second
	if step.Retry.Delay != "" {
		if parsedDelay, err := time.ParseDuration(step.Retry.Delay); err == nil {
			baseDelay = parsedDelay
		}
	}

	// Default backoff strategy
	backoffStrategy := step.Retry.Backoff
	if backoffStrategy == "" {
		backoffStrategy = "fixed"
	}

	// Default stop_on_success
	stopOnSuccess := step.Retry.StopOnSuccess
	if step.Retry.StopOnSuccess == false && step.Retry.Attempts > 0 {
		stopOnSuccess = true // Default to true if not explicitly set
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Calculate delay for this attempt (skip delay on first attempt)
		if attempt > 1 {
			delay := executor.calculateDelay(baseDelay, attempt-1, backoffStrategy)
			fmt.Printf("  [Retry] Waiting %v before attempt %d/%d...\n", delay, attempt, maxAttempts)
			time.Sleep(delay)
		}

		// Print retry attempt info
		if maxAttempts > 1 {
			fmt.Printf("  [Retry] Attempt %d/%d\n", attempt, maxAttempts)
		}

		// Execute the step
		result, err := executor.actionExecutor.ExecuteAction(step, stepNum, loopCtx)
		lastResult = result
		lastErr = err

		// Check if we should stop retrying
		if executor.shouldStopRetrying(result, err, step.Retry, stopOnSuccess) {
			if attempt > 1 && result.Result.Status == constants.ActionStatusPassed {
				fmt.Printf("  [Retry] ✓ Succeeded on attempt %d/%d\n", attempt, maxAttempts)
			}
			break
		}

		// Log retry reason if not the last attempt
		if attempt < maxAttempts {
			fmt.Printf("  [Retry] Failed: %s\n", executor.getRetryReason(result, err))
		}
	}

	return lastResult, lastErr
}

// shouldStopRetrying determines if we should stop retrying based on the result
func (executor *RetryExecutor) shouldStopRetrying(
	result *types.StepResult,
	err error,
	retryConfig *types.RetryConfig,
	stopOnSuccess bool,
) bool {
	// If it succeeded and we should stop on success, stop
	if stopOnSuccess && result.Result.Status == constants.ActionStatusPassed {
		return true
	}

	// Check custom retry condition (retry_if)
	if retryConfig.RetryIf != "" {
		shouldRetry, condErr := executor.evaluateRetryCondition(result, err, retryConfig.RetryIf)
		if condErr != nil {
			fmt.Printf("  [Retry] Warning: Error evaluating retry condition: %v\n", condErr)
			// Fall back to default behavior on condition evaluation error
		} else {
			return !shouldRetry // Stop if condition says don't retry
		}
	}

	// Default: retry on any failure, stop on success
	return result.Result.Status == constants.ActionStatusPassed
}

// evaluateRetryCondition evaluates the retry_if condition with error variables
func (executor *RetryExecutor) evaluateRetryCondition(result *types.StepResult, err error, condition string) (bool, error) {
	// Create a temporary variable scope with error information
	tempVars := executor.variables.Clone()
	
	// Set error variables for condition evaluation
	executor.setErrorVariables(tempVars, result, err)
	
	// Create condition evaluator using the execution package implementation
	conditionEvaluator := NewBasicConditionEvaluator(tempVars)
	
	return conditionEvaluator.Evaluate(condition)
}

// setErrorVariables sets error-related variables for retry condition evaluation
func (executor *RetryExecutor) setErrorVariables(vars *common.Variables, result *types.StepResult, err error) {
	// Set default values
	vars.Set("error_occurred", false)
	vars.Set("error_message", "")
	vars.Set("step_status", "PASSED")
	
	if result != nil {
		// Set step status
		switch result.Result.Status {
		case constants.ActionStatusPassed:
			vars.Set("step_status", "PASSED")
		case constants.ActionStatusFailed:
			vars.Set("step_status", "FAILED")
		case constants.ActionStatusError:
			vars.Set("step_status", "ERROR")
		}
		
		// Set error variables if step didn't pass
		if result.Result.Status != constants.ActionStatusPassed {
			vars.Set("error_occurred", true)
			
			// Set error message from technical errors
			if result.Result.ErrorInfo != nil {
				vars.Set("error_message", result.Result.ErrorInfo.Message)
			}
			
			// Set failure message from logical failures  
			if result.Result.FailureInfo != nil {
				vars.Set("error_message", result.Result.FailureInfo.Message)
			}
		}
	}
	
	// Handle direct errors (like network failures)
	if err != nil {
		vars.Set("error_occurred", true)
		vars.Set("error_message", err.Error())
		vars.Set("step_status", "ERROR")
	}
}

// calculateDelay calculates the delay for a retry attempt based on backoff strategy
func (executor *RetryExecutor) calculateDelay(baseDelay time.Duration, attempt int, strategy string) time.Duration {
	switch strategy {
	case "linear":
		return time.Duration(int64(baseDelay) * int64(attempt+1))
	case "exponential":
		multiplier := math.Pow(2, float64(attempt))
		return time.Duration(float64(baseDelay) * multiplier)
	case "fixed":
		fallthrough
	default:
		return baseDelay
	}
}

// getRetryReason returns a human-readable reason for the retry
func (executor *RetryExecutor) getRetryReason(result *types.StepResult, err error) string {
	if result.Result.ErrorInfo != nil {
		return result.Result.ErrorInfo.Message
	}
	if err != nil {
		return err.Error()
	}
	return "Unknown error"
}

