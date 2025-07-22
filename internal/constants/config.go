package constants

import "time"

// Error template names for safe formatting - only used templates
const (
	// Control flow error templates (actually used)
	TemplateInvalidRangeFormat   = "control_flow.invalid_range_format"
	TemplateInvalidStartValue    = "control_flow.invalid_start_value"
	TemplateInvalidEndValue      = "control_flow.invalid_end_value"
	TemplateInvalidCountFormat   = "control_flow.invalid_count_format"
	TemplateIfConditionFailed    = "control_flow.if_condition_failed"
	TemplateWhileConditionFailed = "control_flow.while_condition_failed"
	TemplateUnknownAction        = "control_flow.unknown_action"
)

// Template content mapping - only used templates
var ErrorTemplates = map[string]string{
	// Control flow templates (actually used)
	TemplateInvalidRangeFormat:   "invalid range format: %s",
	TemplateInvalidStartValue:    "invalid start value: %s",
	TemplateInvalidEndValue:      "invalid end value: %s",
	TemplateInvalidCountFormat:   "invalid count format: %s",
	TemplateIfConditionFailed:    "if condition evaluation failed: %v",
	TemplateWhileConditionFailed: "while condition evaluation failed: %v",
	TemplateUnknownAction:        "unknown action: %s",
}

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