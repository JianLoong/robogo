package constants

import "time"

// Database and messaging action timeouts
const (
	// DefaultDatabaseTimeout is the default timeout for database operations
	DefaultDatabaseTimeout = 30 * time.Second
	
	// DefaultConnectionLifetime is the default lifetime for database connections
	DefaultConnectionLifetime = 1 * time.Second
	
	// DefaultMessagingTimeout is the default timeout for messaging operations
	DefaultMessagingTimeout = 30 * time.Second
)

// Control flow constants
const (
	// MaxWhileLoopIterations is the maximum number of iterations for while loops
	MaxWhileLoopIterations = 10
)