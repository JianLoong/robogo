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

// Action operation constants
const (
	// Database operations
	OperationQuery   = "query"
	OperationSelect  = "select"
	OperationExecute = "execute"
	OperationInsert  = "insert"
	OperationUpdate  = "update"
	OperationDelete  = "delete"
	
	// Messaging operations
	OperationPublish = "publish"
	OperationConsume = "consume"
	
	// Comparison operators
	OperatorEqual              = "=="
	OperatorNotEqual           = "!="
	OperatorGreaterThan        = ">"
	OperatorLessThan           = "<"
	OperatorGreaterThanOrEqual = ">="
	OperatorLessThanOrEqual    = "<="
	OperatorContains           = "contains"
	OperatorStartsWith         = "starts_with"
	OperatorEndsWith           = "ends_with"
)

// Loop variable names
const (
	LoopVariableIteration = "iteration"
	LoopVariableIndex     = "index"
	LoopVariableItem      = "item"
)

// Common error messages
const (
	ErrorInvalidRangeFormat = "invalid range format: %s"
	ErrorInvalidStartValue  = "invalid start value in range: %s"
	ErrorInvalidEndValue    = "invalid end value in range: %s"
	ErrorInvalidCountFormat = "invalid count format: %s"
)