package actions

import (
	"context"
	"fmt"
)

type ActionFunc func(args []interface{}, options map[string]interface{}, silent bool) (interface{}, error)

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

func RabbitMQActionWrapper(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	strArgs := make([]string, len(args))
	for i, v := range args {
		strArgs[i] = fmt.Sprintf("%v", v)
	}
	result, err := RabbitMQActionWithContext(ctx, strArgs)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func KafkaActionWrapper(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return KafkaActionWithContext(ctx, args)
}

// Execute executes an action with the provided arguments and options.
func (ae *ActionExecutor) Execute(action string, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	return ae.registry.Execute(action, args, options, silent)
}
