package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// RetryManager handles retry logic for steps
// Implements RetryPolicy interface
type RetryManager struct {
	// No state needed - stateless retry operations
}

// NewRetryManager creates a new retry manager
func NewRetryManager() RetryPolicy {
	return &RetryManager{}
}

// ExecuteWithRetryLegacy executes a step with retry logic (legacy method)
func (rm *RetryManager) ExecuteWithRetryLegacy(
	ctx context.Context,
	step parser.Step,
	args []interface{},
	executor *actions.ActionExecutor,
	silent bool,
) (interface{}, error) {
	if step.Retry == nil {
		// No retry configuration, execute normally
		return executor.Execute(ctx, step.Action, args, step.Options, silent)
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
		output, err := executor.Execute(ctx, step.Action, args, step.Options, silent)

		// Store the last error and output for potential retry decision
		lastErr = err
		lastOutput = output

		// Check if we should retry based on error conditions and HTTP status codes
		shouldRetry := false
		if err != nil {
			// Traditional error-based retry
			shouldRetry = rm.shouldRetry(err, config.Conditions)
		} else {
			// Check HTTP response status codes for retry conditions
			shouldRetry = rm.shouldRetryBasedOnResponse(output, config.Conditions)
		}

		if !shouldRetry {
			// Success or non-retryable condition - return immediately
			if !silent && attempt > 1 && err == nil {
				fmt.Printf("Retry successful on attempt %d\n", attempt)
			}
			if !silent && !shouldRetry && err != nil {
				fmt.Printf("Error not retryable: %v\n", err)
			}
			return output, err
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

	errMsg := util.FormatRobogoError(err)
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

// shouldRetryBasedOnResponse determines if an HTTP response should trigger a retry based on status code
func (rm *RetryManager) shouldRetryBasedOnResponse(output interface{}, conditions []string) bool {
	if len(conditions) == 0 {
		// Default conditions: retry on 5xx errors
		conditions = []string{"5xx"}
	}

	// Try to parse output as HTTP response
	statusCode := rm.extractStatusCode(output)
	if statusCode == 0 {
		// Not an HTTP response or can't parse status code
		return false
	}

	for _, condition := range conditions {
		conditionLower := strings.ToLower(condition)

		switch {
		case conditionLower == "5xx":
			if statusCode >= 500 && statusCode < 600 {
				return true
			}
		case conditionLower == "4xx":
			if statusCode >= 400 && statusCode < 500 {
				return true
			}
		case conditionLower == "rate_limit" || conditionLower == "429":
			if statusCode == 429 {
				return true
			}
		case conditionLower == "all":
			if statusCode >= 400 {
				return true
			}
		case strings.HasPrefix(conditionLower, "5") && len(conditionLower) == 3:
			// Specific status codes like "500", "502", "503"
			if specificCode := rm.parseStatusCode(conditionLower); specificCode > 0 && statusCode == specificCode {
				return true
			}
		}
	}

	return false
}

// extractStatusCode extracts the status code from HTTP response output
func (rm *RetryManager) extractStatusCode(output interface{}) int {
	// Handle byte array (JSON response from HTTP action)
	if bytes, ok := output.([]byte); ok {
		var response map[string]interface{}
		if err := json.Unmarshal(bytes, &response); err == nil {
			if statusCode, ok := response["status_code"].(float64); ok {
				return int(statusCode)
			}
		}
	}

	// Handle string (might be JSON)
	if str, ok := output.(string); ok {
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(str), &response); err == nil {
			if statusCode, ok := response["status_code"].(float64); ok {
				return int(statusCode)
			}
		}
	}

	// Handle map directly
	if response, ok := output.(map[string]interface{}); ok {
		if statusCode, ok := response["status_code"].(float64); ok {
			return int(statusCode)
		}
		if statusCode, ok := response["status_code"].(int); ok {
			return statusCode
		}
	}

	return 0
}

// parseStatusCode parses a status code string to int
func (rm *RetryManager) parseStatusCode(codeStr string) int {
	switch codeStr {
	case "500":
		return 500
	case "501":
		return 501
	case "502":
		return 502
	case "503":
		return 503
	case "504":
		return 504
	case "429":
		return 429
	default:
		return 0
	}
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

// ShouldRetry implements RetryPolicy interface
func (rm *RetryManager) ShouldRetry(step parser.Step, attempt int, err error) bool {
	if step.Retry == nil {
		return false
	}
	
	config := step.Retry
	maxAttempts := config.Attempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	
	if attempt >= maxAttempts {
		return false
	}
	
	return rm.shouldRetry(err, config.Conditions)
}

// GetRetryDelay implements RetryPolicy interface
func (rm *RetryManager) GetRetryDelay(attempt int) time.Duration {
	// Default configuration for interface method
	config := &parser.RetryConfig{
		Delay:   time.Second,
		Backoff: "fixed",
	}
	return rm.calculateDelay(config, attempt)
}

// ExecuteWithRetry implements RetryPolicy interface with ActionExecutor interface
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, step parser.Step, executor ActionExecutor, silent bool) (interface{}, error) {
	if step.Retry == nil {
		return executor.Execute(ctx, step.Action, step.Args, step.Options, silent)
	}

	config := step.Retry
	attempts := config.Attempts
	if attempts <= 0 {
		attempts = 1
	}

	var lastErr error
	var lastOutput interface{}

	for attempt := 1; attempt <= attempts; attempt++ {
		if !silent && attempt > 1 {
			fmt.Printf("Retry attempt %d/%d for step '%s'\n", attempt, attempts, step.Name)
		}

		output, err := executor.Execute(ctx, step.Action, step.Args, step.Options, silent)
		lastErr = err
		lastOutput = output

		shouldRetry := false
		if err != nil {
			shouldRetry = rm.shouldRetry(err, config.Conditions)
		} else {
			shouldRetry = rm.shouldRetryBasedOnResponse(output, config.Conditions)
		}

		if !shouldRetry {
			if !silent && attempt > 1 && err == nil {
				fmt.Printf("Retry successful on attempt %d\n", attempt)
			}
			return output, err
		}

		if attempt == attempts {
			break
		}

		delay := rm.calculateDelay(config, attempt)
		if !silent {
			fmt.Printf("Waiting %v before retry...\n", delay)
		}
		time.Sleep(delay)
	}

	return lastOutput, fmt.Errorf("step failed after %d attempts: %w", attempts, lastErr)
}

