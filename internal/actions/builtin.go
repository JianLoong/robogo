package actions

import (
	"context"
)

type ActionExecutor struct {
	registry *ActionRegistry
}

// NewActionExecutor creates a new action executor instance.
//
// Returns: ActionExecutor ready to execute Robogo actions
//
// Notes:
//   - Registers all built-in actions
//   - Provides unified interface for action execution
//   - Handles argument validation and error reporting
func NewActionExecutor(registry *ActionRegistry) *ActionExecutor {
	return &ActionExecutor{registry: registry}
}


// Execute executes an action with context support
func (ae *ActionExecutor) Execute(ctx context.Context, action string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return ae.registry.Execute(ctx, action, args, options, silent)
}
