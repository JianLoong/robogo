package execution

import (
	"fmt"
	"math"
	"time"

	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// RetryExecutor handles retry logic for step execution
type RetryExecutor struct {
	stepExecutor StepExecutorInterface
}

// StepExecutorInterface interface for step execution
type StepExecutorInterface interface {
	ExecuteSingleStep(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error)
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(stepExecutor StepExecutorInterface) *RetryExecutor {
	return &RetryExecutor{
		stepExecutor: stepExecutor,
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
		result, err := executor.stepExecutor.ExecuteSingleStep(step, stepNum, loopCtx)
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

	// If there are specific retry conditions, check them
	if len(retryConfig.RetryOn) > 0 {
		return !executor.shouldRetryForError(result, err, retryConfig.RetryOn)
	}

	// Default: retry on any failure, stop on success
	return result.Result.Status == constants.ActionStatusPassed
}

// shouldRetryForError checks if we should retry based on specific error types
func (executor *RetryExecutor) shouldRetryForError(result *types.StepResult, err error, retryOn []string) bool {
	if result.Result.Status == constants.ActionStatusPassed {
		return false
	}

	// Check error categories and codes
	if result.Result.ErrorInfo != nil {
		errorCategory := string(result.Result.ErrorInfo.Category)
		errorCode := result.Result.ErrorInfo.Code

		for _, condition := range retryOn {
			switch condition {
			case "assertion_failed":
				if errorCategory == "assertion" {
					return true
				}
			case "http_error":
				if errorCategory == "http" || errorCategory == "request" {
					return true
				}
			case "timeout":
				if errorCode == "TIMEOUT" || errorCode == "REQUEST_TIMEOUT" {
					return true
				}
			case "connection_error":
				if errorCode == "CONNECTION_FAILED" || errorCode == "CONNECTION_REFUSED" {
					return true
				}
			case "all":
				return true
			}
		}
		return false
	}

	// If no specific error info, retry for any configured condition that includes "all"
	for _, condition := range retryOn {
		if condition == "all" {
			return true
		}
	}

	return false
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
