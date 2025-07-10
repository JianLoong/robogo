package util

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryStrategy defines different retry strategies
type RetryStrategy string

const (
	RetryStrategyFixed       RetryStrategy = "fixed"
	RetryStrategyExponential RetryStrategy = "exponential"
	RetryStrategyLinear      RetryStrategy = "linear"
	RetryStrategyCustom      RetryStrategy = "custom"
)

// RetryConfig defines retry configuration
type RetryConfig struct {
	Strategy     RetryStrategy `json:"strategy"`
	MaxAttempts  int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
	Jitter       bool          `json:"jitter"`
	JitterRange  float64       `json:"jitter_range"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		Strategy:     RetryStrategyExponential,
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		JitterRange:  0.1,
	}
}

// RetryableError defines whether an error should be retried
type RetryableError interface {
	IsRetryable() bool
	GetRetryDelay(attempt int) time.Duration
	ShouldRetry(attempt int, maxAttempts int) bool
}

// RetryContext provides context for retry operations
type RetryContext struct {
	Attempt     int           `json:"attempt"`
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	TotalTime   time.Duration `json:"total_time"`
	LastError   error         `json:"last_error,omitempty"`
	StartTime   time.Time     `json:"start_time"`
}

// RetryExecutor handles retry logic for operations
type RetryExecutor struct {
	config *RetryConfig
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(config *RetryConfig) *RetryExecutor {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryExecutor{
		config: config,
	}
}

// ExecuteWithRetry executes a function with retry logic
func (re *RetryExecutor) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	retryCtx := &RetryContext{
		Attempt:     0,
		MaxAttempts: re.config.MaxAttempts,
		StartTime:   time.Now(),
	}

	for {
		retryCtx.Attempt++
		retryCtx.TotalTime = time.Since(retryCtx.StartTime)

		err := operation()
		if err == nil {
			return nil // Success
		}

		retryCtx.LastError = err

		// Check if we should retry
		if !re.shouldRetry(err, retryCtx) {
			return re.wrapFinalError(err, retryCtx)
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return re.wrapContextError(ctx.Err(), retryCtx)
		default:
		}

		// Calculate delay and wait
		delay := re.calculateDelay(retryCtx.Attempt)
		retryCtx.Delay = delay

		select {
		case <-ctx.Done():
			return re.wrapContextError(ctx.Err(), retryCtx)
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
}

// ExecuteWithRetryTyped executes a function with retry logic and returns a typed result
func (re *RetryExecutor) ExecuteWithRetryTyped(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := re.ExecuteWithRetry(ctx, func() error {
		var err error
		result, err = operation()
		return err
	})
	return result, err
}

// shouldRetry determines if an error should be retried
func (re *RetryExecutor) shouldRetry(err error, retryCtx *RetryContext) bool {
	// Check if we've exceeded max attempts
	if retryCtx.Attempt >= re.config.MaxAttempts {
		return false
	}

	// Check if error is retryable
	if retryable, ok := err.(RetryableError); ok {
		return retryable.ShouldRetry(retryCtx.Attempt, re.config.MaxAttempts)
	}

	// Check if it's a RobogoError
	if roboErr := GetRobogoError(err); roboErr != nil {
		return roboErr.IsRetryable()
	}

	// Default: retry for common transient errors
	return re.isTransientError(err)
}

// isTransientError checks if an error is likely transient
func (re *RetryExecutor) isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	
	// Network-related errors
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"timeout",
		"temporary failure",
		"service unavailable",
		"too many requests",
		"rate limit",
		"circuit breaker",
		"deadline exceeded",
		"context deadline exceeded",
		"i/o timeout",
		"network is unreachable",
		"no route to host",
	}

	for _, pattern := range transientPatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// calculateDelay calculates the delay for the next retry attempt
func (re *RetryExecutor) calculateDelay(attempt int) time.Duration {
	var delay time.Duration

	switch re.config.Strategy {
	case RetryStrategyFixed:
		delay = re.config.InitialDelay
	case RetryStrategyLinear:
		delay = re.config.InitialDelay * time.Duration(attempt)
	case RetryStrategyExponential:
		delay = re.config.InitialDelay * time.Duration(math.Pow(re.config.Multiplier, float64(attempt-1)))
	default:
		delay = re.config.InitialDelay
	}

	// Cap at max delay
	if delay > re.config.MaxDelay {
		delay = re.config.MaxDelay
	}

	// Add jitter if enabled
	if re.config.Jitter {
		delay = re.addJitter(delay)
	}

	return delay
}

// addJitter adds random jitter to the delay
func (re *RetryExecutor) addJitter(delay time.Duration) time.Duration {
	if re.config.JitterRange <= 0 {
		return delay
	}

	jitterAmount := float64(delay) * re.config.JitterRange
	jitter := (rand.Float64() - 0.5) * 2 * jitterAmount
	
	newDelay := delay + time.Duration(jitter)
	if newDelay < 0 {
		newDelay = delay / 2
	}

	return newDelay
}

// wrapFinalError wraps the final error with retry context
func (re *RetryExecutor) wrapFinalError(err error, retryCtx *RetryContext) error {
	return NewErrorBuilder(ErrorTypeExecution, fmt.Sprintf("operation failed after %d attempts", retryCtx.Attempt)).
		WithCause(err).
		WithDetails(map[string]interface{}{
			"attempts":   retryCtx.Attempt,
			"total_time": retryCtx.TotalTime,
			"last_delay": retryCtx.Delay,
		}).
		AddBreadcrumb(fmt.Sprintf("Retry attempt %d/%d failed", retryCtx.Attempt, retryCtx.MaxAttempts)).
		Build()
}

// wrapContextError wraps context cancellation errors
func (re *RetryExecutor) wrapContextError(err error, retryCtx *RetryContext) error {
	return NewErrorBuilder(ErrorTypeTimeout, "operation cancelled or timed out during retry").
		WithCause(err).
		WithDetails(map[string]interface{}{
			"attempts":   retryCtx.Attempt,
			"total_time": retryCtx.TotalTime,
		}).
		AddBreadcrumb(fmt.Sprintf("Context cancelled after %d attempts", retryCtx.Attempt)).
		Build()
}

// RecoveryStrategy defines different recovery strategies
type RecoveryStrategy string

const (
	RecoveryStrategyNone     RecoveryStrategy = "none"
	RecoveryStrategyFallback RecoveryStrategy = "fallback"
	RecoveryStrategySkip     RecoveryStrategy = "skip"
	RecoveryStrategyRetry    RecoveryStrategy = "retry"
	RecoveryStrategyCircuit  RecoveryStrategy = "circuit"
)

// RecoveryConfig defines recovery configuration
type RecoveryConfig struct {
	Strategy            RecoveryStrategy `json:"strategy"`
	FallbackAction      string           `json:"fallback_action,omitempty"`
	SkipOnError         bool             `json:"skip_on_error"`
	RetryConfig         *RetryConfig     `json:"retry_config,omitempty"`
	CircuitBreakerConfig *CircuitBreakerConfig `json:"circuit_breaker_config,omitempty"`
}

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
	MaxRequests      int           `json:"max_requests"`
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed    CircuitBreakerState = "closed"
	CircuitBreakerOpen      CircuitBreakerState = "open"
	CircuitBreakerHalfOpen  CircuitBreakerState = "half_open"
)

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	config        *CircuitBreakerConfig
	state         CircuitBreakerState
	failures      int
	successes     int
	lastFailTime  time.Time
	halfOpenTime  time.Time
	requestCount  int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          30 * time.Second,
			MaxRequests:      10,
		}
	}

	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// Execute executes an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.allowRequest() {
		return NewErrorBuilder(ErrorTypeExecution, "circuit breaker is open").
			WithDetails(map[string]interface{}{
				"state":    cb.state,
				"failures": cb.failures,
			}).
			Build()
	}

	err := operation()
	cb.recordResult(err)
	return err
}

// allowRequest checks if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	now := time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		if now.Sub(cb.lastFailTime) > cb.config.Timeout {
			cb.state = CircuitBreakerHalfOpen
			cb.halfOpenTime = now
			cb.requestCount = 0
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return cb.requestCount < cb.config.MaxRequests
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.state == CircuitBreakerHalfOpen {
			cb.state = CircuitBreakerOpen
		} else if cb.failures >= cb.config.FailureThreshold {
			cb.state = CircuitBreakerOpen
		}
	} else {
		cb.successes++
		
		if cb.state == CircuitBreakerHalfOpen {
			if cb.successes >= cb.config.SuccessThreshold {
				cb.state = CircuitBreakerClosed
				cb.failures = 0
				cb.successes = 0
			}
		}
	}

	if cb.state == CircuitBreakerHalfOpen {
		cb.requestCount++
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"state":         cb.state,
		"failures":      cb.failures,
		"successes":     cb.successes,
		"last_fail_time": cb.lastFailTime,
		"request_count": cb.requestCount,
	}
}

// RecoveryExecutor handles error recovery strategies
type RecoveryExecutor struct {
	config         *RecoveryConfig
	retryExecutor  *RetryExecutor
	circuitBreaker *CircuitBreaker
}

// NewRecoveryExecutor creates a new recovery executor
func NewRecoveryExecutor(config *RecoveryConfig) *RecoveryExecutor {
	if config == nil {
		config = &RecoveryConfig{
			Strategy: RecoveryStrategyRetry,
			RetryConfig: DefaultRetryConfig(),
		}
	}

	executor := &RecoveryExecutor{
		config: config,
	}

	if config.RetryConfig != nil {
		executor.retryExecutor = NewRetryExecutor(config.RetryConfig)
	}

	if config.CircuitBreakerConfig != nil {
		executor.circuitBreaker = NewCircuitBreaker(config.CircuitBreakerConfig)
	}

	return executor
}

// ExecuteWithRecovery executes an operation with recovery strategy
func (re *RecoveryExecutor) ExecuteWithRecovery(ctx context.Context, operation func() error, fallback func() error) error {
	switch re.config.Strategy {
	case RecoveryStrategyNone:
		return operation()
	case RecoveryStrategyFallback:
		err := operation()
		if err != nil && fallback != nil {
			return fallback()
		}
		return err
	case RecoveryStrategySkip:
		err := operation()
		if err != nil && re.config.SkipOnError {
			return nil // Skip error
		}
		return err
	case RecoveryStrategyRetry:
		if re.retryExecutor != nil {
			return re.retryExecutor.ExecuteWithRetry(ctx, operation)
		}
		return operation()
	case RecoveryStrategyCircuit:
		if re.circuitBreaker != nil {
			return re.circuitBreaker.Execute(operation)
		}
		return operation()
	default:
		return operation()
	}
}

// Helper functions

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     indexSubstring(s, substr) >= 0))
}

// indexSubstring finds the index of a substring in a string
func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Enhanced error builders for retry/recovery scenarios

// NewRetryExhaustedError creates an error for when retries are exhausted
func NewRetryExhaustedError(operation string, attempts int, lastError error) *RobogoError {
	return NewErrorBuilder(ErrorTypeExecution, fmt.Sprintf("operation '%s' failed after %d retry attempts", operation, attempts)).
		WithCause(lastError).
		WithDetails(map[string]interface{}{
			"operation": operation,
			"attempts":  attempts,
		}).
		AddBreadcrumb(fmt.Sprintf("Retry exhausted for operation: %s", operation)).
		Build()
}

// NewCircuitBreakerError creates an error for circuit breaker open state
func NewCircuitBreakerError(operation string, state CircuitBreakerState) *RobogoError {
	return NewErrorBuilder(ErrorTypeExecution, fmt.Sprintf("circuit breaker is %s for operation '%s'", state, operation)).
		WithDetails(map[string]interface{}{
			"operation": operation,
			"state":     state,
		}).
		AddBreadcrumb(fmt.Sprintf("Circuit breaker %s for: %s", state, operation)).
		Build()
}

// NewRecoveryError creates an error for recovery failures
func NewRecoveryError(operation string, strategy RecoveryStrategy, cause error) *RobogoError {
	return NewErrorBuilder(ErrorTypeExecution, fmt.Sprintf("recovery strategy '%s' failed for operation '%s'", strategy, operation)).
		WithCause(cause).
		WithDetails(map[string]interface{}{
			"operation": operation,
			"strategy":  strategy,
		}).
		AddBreadcrumb(fmt.Sprintf("Recovery failed for operation: %s", operation)).
		Build()
}

// Note: Step configuration parsing functions moved to parser package to avoid import cycles