package execution

import (
	"github.com/JianLoong/robogo/internal/types"
)

// ConditionalExecutionStrategy handles if conditions
type ConditionalExecutionStrategy struct {
	conditionEvaluator *BasicConditionEvaluator
	strategyRouter     *ExecutionStrategyRouter
}

// NewConditionalExecutionStrategy creates a new conditional execution strategy
func NewConditionalExecutionStrategy(conditionEvaluator *BasicConditionEvaluator, strategyRouter *ExecutionStrategyRouter) *ConditionalExecutionStrategy {
	return &ConditionalExecutionStrategy{
		conditionEvaluator: conditionEvaluator,
		strategyRouter:     strategyRouter,
	}
}

// Execute performs conditional execution
func (s *ConditionalExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	// Evaluate condition
	condition, err := s.conditionEvaluator.Evaluate(step.If)
	if err != nil {
		return nil, err
	}
	
	// If condition is false, skip execution
	if !condition {
		return &types.StepResult{
			Name:   step.Name,
			Result: types.ActionResult{Status: "SKIPPED"},
		}, nil
	}
	
	// Create a copy of the step without the if condition to avoid infinite recursion
	execStep := step
	execStep.If = ""
	
	// Execute the step normally
	return s.strategyRouter.Execute(execStep, stepNum, loopCtx)
}

// CanHandle returns true for steps with if conditions
func (s *ConditionalExecutionStrategy) CanHandle(step types.Step) bool {
	return step.If != ""
}

// Priority returns highest priority as conditional logic is most specific
func (s *ConditionalExecutionStrategy) Priority() int {
	return 4
}

