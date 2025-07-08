package runner

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// RetryManager handles retry logic for steps
type RetryManager struct {
	// No state needed - stateless retry operations
}

// NewRetryManager creates a new retry manager
func NewRetryManager() *RetryManager {
	return &RetryManager{}
}

// ExecuteWithRetry executes a step with retry logic
func (rm *RetryManager) ExecuteWithRetry(
	step parser.Step,
	args []interface{},
	executor *actions.ActionExecutor,
	silent bool,
) (interface{}, error) {
	if step.Retry == nil {
		// No retry configuration, execute normally
		return executor.Execute(step.Action, args, step.Options, silent)
	}

	config := step.Retry
	attempts := config.Attempts
	if attempts <= 0 {
		attempts = 1 // Default to 1 attempt if not specified
	}

	var lastErr error
	var lastOutput interface{}

	for attempt := 1; attempt <= attempts; attempt++ {
		if !silent && attempt > 1 {
			fmt.Printf("Retry attempt %d/%d for step '%s'\n", attempt, attempts, step.Name)
		}

		// Execute the step
		output, err := executor.Execute(step.Action, args, step.Options, silent)

		if err == nil {
			// Success - return immediately
			if !silent && attempt > 1 {
				fmt.Printf("Retry successful on attempt %d\n", attempt)
			}
			return output, nil
		}

		// Store the last error and output
		lastErr = err
		lastOutput = output

		// Check if we should retry based on error conditions
		if !rm.shouldRetry(err, config.Conditions) {
			if !silent {
				fmt.Printf("Error not retryable: %v\n", err)
			}
			break
		}

		// If this is the last attempt, don't wait
		if attempt == attempts {
			break
		}

		// Calculate delay for next attempt
		delay := rm.calculateDelay(config, attempt)
		if !silent {
			fmt.Printf("Waiting %v before retry...\n", delay)
		}

		time.Sleep(delay)
	}

	// All attempts failed
	return lastOutput, fmt.Errorf("step failed after %d attempts: %w", attempts, lastErr)
}

// shouldRetry determines if an error should trigger a retry based on conditions
func (rm *RetryManager) shouldRetry(err error, conditions []string) bool {
	if len(conditions) == 0 {
		// No conditions specified, retry on any error
		return true
	}

	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	for _, condition := range conditions {
		conditionLower := strings.ToLower(condition)

		switch conditionLower {
		case "5xx", "500", "502", "503", "504":
			// Check for HTTP 5xx errors
			if strings.Contains(errMsgLower, "status code 5") {
				return true
			}
		case "timeout":
			if strings.Contains(errMsgLower, "timeout") || strings.Contains(errMsgLower, "deadline exceeded") {
				return true
			}
		case "connection_error", "connection":
			if strings.Contains(errMsgLower, "connection") || strings.Contains(errMsgLower, "network") {
				return true
			}
		case "any":
			// Retry on any error
			return true
		default:
			// Custom condition - check if error message contains the condition
			if strings.Contains(errMsgLower, conditionLower) {
				return true
			}
		}
	}

	return false
}

// calculateDelay calculates the delay for the next retry attempt
func (rm *RetryManager) calculateDelay(config *parser.RetryConfig, attempt int) time.Duration {
	baseDelay := config.Delay
	if baseDelay <= 0 {
		baseDelay = time.Second // Default 1 second delay
	}

	var delay time.Duration

	switch strings.ToLower(config.Backoff) {
	case "linear":
		// Linear backoff: delay * attempt
		delay = baseDelay * time.Duration(attempt)
	case "exponential":
		// Exponential backoff: delay * 2^(attempt-1)
		multiplier := 1 << (attempt - 1)
		delay = baseDelay * time.Duration(multiplier)
	default:
		// Fixed backoff (default)
		delay = baseDelay
	}

	// Apply maximum delay cap if specified
	if config.MaxDelay > 0 && delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	// Apply jitter if enabled
	if config.Jitter {
		// Add ±10% jitter
		jitterRange := delay / 10
		delay = delay + time.Duration(rand.Int63n(int64(jitterRange*2))-int64(jitterRange))
	}

	return delay
}

// ValidateRetryConfig validates retry configuration
func (rm *RetryManager) ValidateRetryConfig(config *parser.RetryConfig) error {
	if config == nil {
		return nil
	}

	if config.Attempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if config.Delay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	if config.MaxDelay < 0 {
		return fmt.Errorf("max delay cannot be negative")
	}

	if config.MaxDelay > 0 && config.Delay > config.MaxDelay {
		return fmt.Errorf("base delay cannot be greater than max delay")
	}

	return nil
}
