package parser

import (
	"github.com/JianLoong/robogo/internal/util"
)

// GetStepRetryConfig extracts retry configuration from a step
func GetStepRetryConfig(step Step) *util.RetryConfig {
	if step.Retry == nil {
		return nil
	}

	config := &util.RetryConfig{
		Strategy:     util.RetryStrategyExponential,
		MaxAttempts:  step.Retry.Attempts,
		InitialDelay: step.Retry.Delay,
		MaxDelay:     step.Retry.MaxDelay,
		Multiplier:   2.0, // Default multiplier
		Jitter:       step.Retry.Jitter,
	}

	// Map parser backoff strategy to retry strategy
	switch step.Retry.Backoff {
	case "fixed":
		config.Strategy = util.RetryStrategyFixed
	case "linear":
		config.Strategy = util.RetryStrategyLinear
	case "exponential":
		config.Strategy = util.RetryStrategyExponential
	default:
		config.Strategy = util.RetryStrategyExponential
	}

	return config
}

// GetStepRecoveryConfig extracts recovery configuration from a step
func GetStepRecoveryConfig(step Step) *util.RecoveryConfig {
	if step.Recovery == nil {
		return nil
	}

	config := &util.RecoveryConfig{
		Strategy:       util.RecoveryStrategy(step.Recovery.Strategy),
		FallbackAction: step.Recovery.FallbackAction,
		SkipOnError:    step.Recovery.SkipOnError,
	}

	// Include retry config if present
	if step.Retry != nil {
		config.RetryConfig = GetStepRetryConfig(step)
	}

	return config
}

// ConvertRetryConfigToParser converts util.RetryConfig to parser.RetryConfig
func ConvertRetryConfigToParser(config *util.RetryConfig) *RetryConfig {
	if config == nil {
		return nil
	}

	parserConfig := &RetryConfig{
		Attempts: config.MaxAttempts,
		Delay:    config.InitialDelay,
		MaxDelay: config.MaxDelay,
		Jitter:   config.Jitter,
	}

	// Map retry strategy to parser backoff strategy
	switch config.Strategy {
	case util.RetryStrategyFixed:
		parserConfig.Backoff = "fixed"
	case util.RetryStrategyLinear:
		parserConfig.Backoff = "linear"
	case util.RetryStrategyExponential:
		parserConfig.Backoff = "exponential"
	default:
		parserConfig.Backoff = "exponential"
	}

	return parserConfig
}

// ConvertRecoveryConfigToParser converts util.RecoveryConfig to parser.RecoveryConfig
func ConvertRecoveryConfigToParser(config *util.RecoveryConfig) *RecoveryConfig {
	if config == nil {
		return nil
	}

	return &RecoveryConfig{
		Strategy:       string(config.Strategy),
		FallbackAction: config.FallbackAction,
		SkipOnError:    config.SkipOnError,
	}
}