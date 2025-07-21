package execution

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// RetryExecutor handles retry logic for step execution
type RetryExecutor struct {
	stepExecutor StepExecutorInterface
	variables    *common.Variables
}

// StepExecutorInterface interface for step execution
type StepExecutorInterface interface {
	ExecuteSingleStep(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error)
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(stepExecutor StepExecutorInterface, variables *common.Variables) *RetryExecutor {
	return &RetryExecutor{
		stepExecutor: stepExecutor,
		variables:    variables,
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
	
	// Create condition evaluator
	conditionEvaluator := &ConditionEvaluator{
		variables: tempVars,
	}
	
	return conditionEvaluator.Evaluate(condition)
}

// setErrorVariables sets error-related variables for retry condition evaluation
func (executor *RetryExecutor) setErrorVariables(vars *common.Variables, result *types.StepResult, err error) {
	// Clear any existing error variables
	vars.Set("_error_category", "")
	vars.Set("_error_code", "")
	vars.Set("_error_message", "")
	vars.Set("_status_code", "")
	vars.Set("_has_error", "false")
	
	if result != nil && result.Result.Status != constants.ActionStatusPassed {
		vars.Set("_has_error", "true")
		
		// Set error info from technical errors
		if result.Result.ErrorInfo != nil {
			vars.Set("_error_category", string(result.Result.ErrorInfo.Category))
			vars.Set("_error_code", result.Result.ErrorInfo.Code)
			vars.Set("_error_message", result.Result.ErrorInfo.Message)
		}
		
		// Set failure info from logical failures  
		if result.Result.FailureInfo != nil {
			vars.Set("_error_category", string(result.Result.FailureInfo.Category))
			vars.Set("_error_code", result.Result.FailureInfo.Code)
			vars.Set("_error_message", result.Result.FailureInfo.Message)
		}
		
		// Extract HTTP status code if available
		if result.Result.Data != nil {
			if dataMap, ok := result.Result.Data.(map[string]any); ok {
				if statusCode, exists := dataMap["status_code"]; exists {
					vars.Set("_status_code", fmt.Sprintf("%v", statusCode))
				}
			}
		}
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

// ConditionEvaluator for retry logic
type ConditionEvaluator struct {
	variables *common.Variables
}

// Evaluate evaluates a condition string
func (evaluator *ConditionEvaluator) Evaluate(condition string) (bool, error) {
	// Substitute variables first
	condition = evaluator.variables.Substitute(condition)

	// Handle simple boolean values
	if condition == "true" {
		return true, nil
	}
	if condition == "false" {
		return false, nil
	}

	// Handle comparison operators
	operators := []string{">=", "<=", ">", "<", "==", "!=", "contains", "starts_with", "ends_with"}

	for _, op := range operators {
		if strings.Contains(condition, op) {
			return evaluator.evaluateComparison(condition, op)
		}
	}

	// If no operators found, treat non-empty strings as true
	return strings.TrimSpace(condition) != "" && strings.TrimSpace(condition) != "0", nil
}

// evaluateComparison evaluates a comparison expression
func (evaluator *ConditionEvaluator) evaluateComparison(condition, operator string) (bool, error) {
	parts := strings.SplitN(condition, operator, 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid comparison: %s", condition)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	switch operator {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "contains":
		return strings.Contains(left, right), nil
	case "starts_with":
		return strings.HasPrefix(left, right), nil
	case "ends_with":
		return strings.HasSuffix(left, right), nil
	case ">", "<", ">=", "<=":
		return evaluator.compareNumeric(left, right, operator)
	}

	return false, fmt.Errorf("unsupported operator: %s", operator)
}

// compareNumeric compares two values numerically
func (evaluator *ConditionEvaluator) compareNumeric(left, right, operator string) (bool, error) {
	leftNum, err1 := strconv.ParseFloat(left, 64)
	rightNum, err2 := strconv.ParseFloat(right, 64)

	if err1 != nil || err2 != nil {
		// Fall back to string comparison
		switch operator {
		case ">":
			return left > right, nil
		case "<":
			return left < right, nil
		case ">=":
			return left >= right, nil
		case "<=":
			return left <= right, nil
		}
	}

	switch operator {
	case ">":
		return leftNum > rightNum, nil
	case "<":
		return leftNum < rightNum, nil
	case ">=":
		return leftNum >= rightNum, nil
	case "<=":
		return leftNum <= rightNum, nil
	}

	return false, fmt.Errorf("invalid numeric operator: %s", operator)
}
