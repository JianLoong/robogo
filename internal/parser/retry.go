package parser

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// RetryResult represents the result of a retry operation
type RetryResult struct {
	Success    bool
	Attempts   int
	TotalTime  time.Duration
	LastError  error
	LastOutput string
	RetryLogs  []string
}

// ShouldRetry determines if an error should trigger a retry based on conditions
func ShouldRetry(err error, output string, conditions []string) bool {
	if len(conditions) == 0 {
		// Default conditions if none specified
		conditions = []string{"5xx", "timeout", "connection_error"}
	}

	errStr := strings.ToLower(err.Error())
	outputStr := strings.ToLower(output)

	for _, condition := range conditions {
		switch strings.ToLower(condition) {
		case "5xx":
			if strings.Contains(outputStr, "500") || strings.Contains(outputStr, "502") ||
				strings.Contains(outputStr, "503") || strings.Contains(outputStr, "504") {
				return true
			}
		case "4xx":
			if strings.Contains(outputStr, "429") { // Rate limiting
				return true
			}
		case "timeout":
			if strings.Contains(errStr, "timeout") || strings.Contains(outputStr, "timeout") {
				return true
			}
		case "connection_error":
			if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") ||
				strings.Contains(errStr, "refused") || strings.Contains(errStr, "unreachable") {
				return true
			}
		case "rate_limit":
			if strings.Contains(outputStr, "429") || strings.Contains(errStr, "rate limit") {
				return true
			}
		case "all":
			return true
		}
	}

	return false
}

// CalculateDelay calculates the delay for a retry attempt based on backoff strategy
func CalculateDelay(baseDelay time.Duration, attempt int, backoff string, maxDelay time.Duration, jitter bool) time.Duration {
	var delay time.Duration

	switch strings.ToLower(backoff) {
	case "linear":
		delay = baseDelay * time.Duration(attempt)
	case "exponential":
		delay = baseDelay * time.Duration(1<<(attempt-1)) // 2^(attempt-1)
	default: // fixed
		delay = baseDelay
	}

	// Apply max delay cap
	if maxDelay > 0 && delay > maxDelay {
		delay = maxDelay
	}

	// Apply jitter if enabled
	if jitter {
		jitterAmount := delay / 4 // 25% jitter
		jitterValue := time.Duration(rand.Int63n(int64(jitterAmount)))
		delay += jitterValue
	}

	return delay
}

// FormatRetryLog formats retry attempt information for display
func FormatRetryLog(attempt int, totalAttempts int, delay time.Duration, err error, verbose bool) string {
	if !verbose {
		return fmt.Sprintf("Attempt %d/%d", attempt, totalAttempts)
	}

	return fmt.Sprintf("Attempt %d/%d (delay: %v): %v", attempt, totalAttempts, delay, err)
}

// FormatRetrySummary formats a summary of retry attempts
func FormatRetrySummary(result RetryResult, verbose bool) string {
	if result.Success {
		if verbose {
			return fmt.Sprintf("Success after %d attempts (total time: %v)", result.Attempts, result.TotalTime)
		}
		return fmt.Sprintf("Success after %d attempts", result.Attempts)
	}

	if verbose {
		return fmt.Sprintf("Failed after %d attempts (total time: %v): %v", result.Attempts, result.TotalTime, result.LastError)
	}
	return fmt.Sprintf("Failed after %d attempts: %v", result.Attempts, result.LastError)
}

// ValidateRetryConfig validates retry configuration
func ValidateRetryConfig(config *RetryConfig) error {
	if config == nil {
		return nil
	}

	if config.Attempts < 1 {
		return fmt.Errorf("retry attempts must be at least 1")
	}

	if config.Delay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	if config.MaxDelay > 0 && config.Delay > config.MaxDelay {
		return fmt.Errorf("retry delay cannot be greater than max_delay")
	}

	validBackoffs := []string{"fixed", "linear", "exponential"}
	if config.Backoff != "" {
		valid := false
		for _, b := range validBackoffs {
			if strings.ToLower(config.Backoff) == b {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid backoff strategy: %s (valid: %v)", config.Backoff, validBackoffs)
		}
	}

	return nil
}

// GetDefaultRetryConfig returns default retry configuration
func GetDefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		Attempts:   1,           // No retries by default
		Delay:      time.Second, // 1 second delay
		Backoff:    "fixed",     // Fixed delay
		Conditions: []string{"5xx", "timeout", "connection_error"},
		MaxDelay:   0,     // No max delay by default
		Jitter:     false, // No jitter by default
	}
}

// MergeRetryConfig merges a custom config with defaults
func MergeRetryConfig(custom *RetryConfig) *RetryConfig {
	if custom == nil {
		return GetDefaultRetryConfig()
	}

	defaults := GetDefaultRetryConfig()

	if custom.Attempts == 0 {
		custom.Attempts = defaults.Attempts
	}
	if custom.Delay == 0 {
		custom.Delay = defaults.Delay
	}
	if custom.Backoff == "" {
		custom.Backoff = defaults.Backoff
	}
	if len(custom.Conditions) == 0 {
		custom.Conditions = defaults.Conditions
	}
	if custom.MaxDelay == 0 {
		custom.MaxDelay = defaults.MaxDelay
	}

	return custom
}
