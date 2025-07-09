package util

import (
	"context"
	"time"
)

// WithDefaultTimeout applies a default timeout to the context if no timeout is already set
// or if the existing timeout is longer than the default
func WithDefaultTimeout(ctx context.Context, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > defaultTimeout {
		return context.WithTimeout(ctx, defaultTimeout)
	}
	// Return the context as-is with a no-op cancel function
	return ctx, func() {}
}

// WithActionTimeout applies a timeout specific to an action, respecting existing timeouts
func WithActionTimeout(ctx context.Context, actionTimeout time.Duration, action string) (context.Context, context.CancelFunc) {
	if deadline, ok := ctx.Deadline(); ok {
		// If existing deadline is sooner, use it
		if time.Until(deadline) <= actionTimeout {
			return ctx, func() {}
		}
	}
	
	// Apply the action timeout
	return context.WithTimeout(ctx, actionTimeout)
}

// ParseTimeout parses a timeout string into a time.Duration
func ParseTimeout(timeoutStr string, defaultTimeout time.Duration) time.Duration {
	if timeoutStr == "" {
		return defaultTimeout
	}
	
	if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
		return parsedTimeout
	}
	
	return defaultTimeout
}

// IsContextTimeout checks if an error is a context timeout
func IsContextTimeout(err error) bool {
	return err == context.DeadlineExceeded || err == context.Canceled
}