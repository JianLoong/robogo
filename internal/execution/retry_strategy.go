package execution

import (
	"fmt"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// RetryExecutionStrategy handles steps with retry logic
type RetryExecutionStrategy struct {
	basicStrategy *BasicExecutionStrategy
	variables     *common.Variables
}

// NewRetryExecutionStrategy creates a new retry execution strategy
func NewRetryExecutionStrategy(variables *common.Variables, actionRegistry *actions.ActionRegistry) *RetryExecutionStrategy {
	return &RetryExecutionStrategy{
		basicStrategy: NewBasicExecutionStrategy(variables, actionRegistry),
		variables:     variables,
	}
}

// Execute performs action execution with retry logic
func (s *RetryExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	return s.executeStepWithRetry(step, stepNum, loopCtx)
}

// CanHandle returns true for steps that have retry configuration
func (s *RetryExecutionStrategy) CanHandle(step types.Step) bool {
	return step.Retry != nil && step.Action != ""
}

// Priority returns high priority as retry is a specific concern
func (s *RetryExecutionStrategy) Priority() int {
	return 3
}

// executeStepWithRetry executes a step with retry logic (embedded from RetryExecutor)
func (s *RetryExecutionStrategy) executeStepWithRetry(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	config := step.Retry
	var lastResult *types.StepResult
	var lastError error

	// Create a condition evaluator for retry_if conditions
	conditionEvaluator := NewBasicConditionEvaluator(s.variables)

	for attempt := 1; attempt <= config.Attempts; attempt++ {
		if attempt > 1 {
			fmt.Printf("  [Retry] Attempt %d/%d\n", attempt, config.Attempts)
		}

		result, err := s.basicStrategy.Execute(step, stepNum, loopCtx)
		lastResult = result
		lastError = err

		// Check if we should stop retrying based on success
		if err == nil && result != nil && result.Result.Status == constants.ActionStatusPassed {
			// If stop_on_success is true or not specified, stop retrying on success
			if config.StopOnSuccess {
				return result, nil
			}
		}

		// Set error variables for condition evaluation
		errorOccurred := err != nil || (result != nil && result.Result.Status != constants.ActionStatusPassed)
		errorMessage := ""
		if err != nil {
			errorMessage = err.Error()
		}

		// Store error info in variables for potential use in retry_if conditions
		s.variables.Set("error_occurred", errorOccurred)
		s.variables.Set("error_message", errorMessage)
		if result != nil {
			s.variables.Set("step_status", string(result.Result.Status))
		}

		// Check if we should retry based on retry_on error types
		if len(config.RetryOn) > 0 {
			shouldRetry := false

			// Check if the error type matches any in the retry_on list
			for _, errorType := range config.RetryOn {
				switch strings.ToLower(errorType) {
				case "all":
					shouldRetry = errorOccurred
				case "http_error":
					shouldRetry = errorOccurred && strings.Contains(errorMessage, "HTTP")
				case "timeout":
					shouldRetry = errorOccurred && strings.Contains(errorMessage, "timeout")
				case "connection_error":
					shouldRetry = errorOccurred && (strings.Contains(errorMessage, "connection") ||
						strings.Contains(errorMessage, "dial") ||
						strings.Contains(errorMessage, "network"))
				case "assertion_failed":
					shouldRetry = errorOccurred && strings.Contains(errorMessage, "assertion")
				}

				if shouldRetry {
					fmt.Printf("  [Retry] Error type '%s' matched, continuing retry\n", errorType)
					break
				}
			}

			if !shouldRetry {
				fmt.Printf("  [Retry] Error type doesn't match retry_on criteria, stopping retry\n")
				return lastResult, lastError
			}
		}

		// Check if we should retry based on retry_if condition
		if config.RetryIf != "" {
			// Evaluate the retry_if condition
			shouldRetry, evalErr := conditionEvaluator.Evaluate(config.RetryIf)

			if evalErr != nil {
				fmt.Printf("  [Retry] Warning: Failed to evaluate retry_if condition: %v\n", evalErr)
				// Continue with default behavior on evaluation error
			} else if !shouldRetry {
				// If the condition evaluates to false, stop retrying
				fmt.Printf("  [Retry] Condition evaluated to false, stopping retry\n")
				return lastResult, lastError
			} else {
				fmt.Printf("  [Retry] Condition evaluated to true, continuing retry\n")
			}
		}

		// If this was the last attempt, don't wait
		if attempt == config.Attempts {
			break
		}

		// Calculate delay and wait
		delay := s.calculateDelay(config, attempt-1)
		if delay > 0 {
			fmt.Printf("  [Retry] Waiting %v before next attempt...\n", delay)
			time.Sleep(delay)
		}
	}

	return lastResult, lastError
}

// calculateDelay calculates the delay for retry attempts
func (s *RetryExecutionStrategy) calculateDelay(config *types.RetryConfig, attemptNum int) time.Duration {
	if config.Delay == "" {
		return 0
	}

	baseDuration, err := time.ParseDuration(config.Delay)
	if err != nil {
		return time.Second // Default to 1 second if parsing fails
	}

	switch config.Backoff {
	case "exponential":
		multiplier := 1
		for i := 0; i < attemptNum; i++ {
			multiplier *= 2
		}
		return time.Duration(multiplier) * baseDuration
	case "linear":
		return time.Duration(attemptNum+1) * baseDuration
	default: // "fixed" or unrecognized
		return baseDuration
	}
}
