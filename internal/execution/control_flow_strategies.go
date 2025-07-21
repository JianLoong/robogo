package execution

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// ConditionalExecutionStrategy handles if conditions
type ConditionalExecutionStrategy struct {
	conditionEvaluator ConditionEvaluator
	strategyRouter     *ExecutionStrategyRouter
}

// NewConditionalExecutionStrategy creates a new conditional execution strategy
func NewConditionalExecutionStrategy(conditionEvaluator ConditionEvaluator, strategyRouter *ExecutionStrategyRouter) *ConditionalExecutionStrategy {
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

// Priority returns high priority as conditional logic is specific
func (s *ConditionalExecutionStrategy) Priority() int {
	return 6
}

// LoopExecutionStrategy handles for and while loops using existing loop executor
type LoopExecutionStrategy struct {
	variables          *common.Variables
	conditionEvaluator ConditionEvaluator
	stepExecutor       StepExecutor
}

// NewLoopExecutionStrategy creates a new loop execution strategy
func NewLoopExecutionStrategy(variables *common.Variables, conditionEvaluator ConditionEvaluator, stepExecutor StepExecutor) *LoopExecutionStrategy {
	return &LoopExecutionStrategy{
		variables:          variables,
		conditionEvaluator: conditionEvaluator,
		stepExecutor:       stepExecutor,
	}
}

// Execute performs loop execution by delegating to existing loop logic
func (s *LoopExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	// For now, we'll fall back to using the existing control flow logic
	// This avoids the import cycle while still providing the strategy interface
	// TODO: Refactor loop logic to be strategy-based later
	return s.stepExecutor.ExecuteStepWithContext(step, stepNum, loopCtx)
}

// CanHandle returns true for steps with for or while loops
func (s *LoopExecutionStrategy) CanHandle(step types.Step) bool {
	return step.For != "" || step.While != ""
}

// Priority returns high priority as loops are specific
func (s *LoopExecutionStrategy) Priority() int {
	return 7
}