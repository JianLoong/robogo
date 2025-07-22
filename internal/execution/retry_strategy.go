package execution

import (
	"fmt"
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

	for attempt := 1; attempt <= config.Attempts; attempt++ {
		if attempt > 1 {
			fmt.Printf("  [Retry] Attempt %d/%d\n", attempt, config.Attempts)
		}

		result, err := s.basicStrategy.Execute(step, stepNum, loopCtx)
		lastResult = result
		lastError = err

		// Check if step succeeded
		if err == nil && result != nil && result.Result.Status == constants.ActionStatusPassed {
			return result, nil
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