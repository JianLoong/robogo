package execution

import (
	"github.com/JianLoong/robogo/internal/types"
)

// StepExecutor defines the unified interface for step execution across all components
type StepExecutor interface {
	// ExecuteStepWithContext executes a single step with context information
	ExecuteStepWithContext(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error)
}

// ConditionEvaluator defines the interface for evaluating conditions
type ConditionEvaluator interface {
	// Evaluate evaluates a condition expression and returns true/false
	Evaluate(condition string) (bool, error)
}

// ActionExecutor defines the interface for executing individual actions
type ActionExecutor interface {
	// ExecuteAction executes a single action without retry or control flow logic
	ExecuteAction(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error)
}