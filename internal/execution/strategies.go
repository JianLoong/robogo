package execution

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// ExecutionStrategy defines different ways to execute steps
type ExecutionStrategy interface {
	// Execute runs the step according to this strategy
	Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error)
	
	// CanHandle returns true if this strategy can handle the given step
	CanHandle(step types.Step) bool
	
	// Priority returns the priority of this strategy (higher = more specific)
	Priority() int
}

// BasicExecutionStrategy handles simple action execution without any control flow
type BasicExecutionStrategy struct {
	actionExecutor ActionExecutor
}

// NewBasicExecutionStrategy creates a new basic execution strategy
func NewBasicExecutionStrategy(actionExecutor ActionExecutor) *BasicExecutionStrategy {
	return &BasicExecutionStrategy{
		actionExecutor: actionExecutor,
	}
}

// Execute performs basic action execution
func (s *BasicExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	return s.actionExecutor.ExecuteAction(step, stepNum, loopCtx)
}

// CanHandle returns true for steps that have an action and no control flow
func (s *BasicExecutionStrategy) CanHandle(step types.Step) bool {
	return step.Action != "" && 
		step.Retry == nil && 
		step.If == "" && 
		step.For == "" && 
		step.While == "" &&
		len(step.Steps) == 0
}

// Priority returns low priority as this is the fallback strategy
func (s *BasicExecutionStrategy) Priority() int {
	return 1
}

// RetryExecutionStrategy handles steps with retry logic
type RetryExecutionStrategy struct {
	actionExecutor ActionExecutor
	variables      *common.Variables
}

// NewRetryExecutionStrategy creates a new retry execution strategy
func NewRetryExecutionStrategy(actionExecutor ActionExecutor, variables *common.Variables) *RetryExecutionStrategy {
	return &RetryExecutionStrategy{
		actionExecutor: actionExecutor,
		variables:      variables,
	}
}

// Execute performs action execution with retry logic
func (s *RetryExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	retryExecutor := NewRetryExecutor(s.actionExecutor, s.variables)
	return retryExecutor.ExecuteStepWithRetry(step, stepNum, loopCtx)
}

// CanHandle returns true for steps that have retry configuration
func (s *RetryExecutionStrategy) CanHandle(step types.Step) bool {
	return step.Retry != nil && step.Action != ""
}

// Priority returns high priority as retry is a specific concern
func (s *RetryExecutionStrategy) Priority() int {
	return 5
}

// NestedStepsExecutionStrategy handles steps with nested steps
type NestedStepsExecutionStrategy struct {
	strategyRouter *ExecutionStrategyRouter
}

// NewNestedStepsExecutionStrategy creates a new nested steps execution strategy
func NewNestedStepsExecutionStrategy(strategyRouter *ExecutionStrategyRouter) *NestedStepsExecutionStrategy {
	return &NestedStepsExecutionStrategy{
		strategyRouter: strategyRouter,
	}
}

// Execute performs nested steps execution
func (s *NestedStepsExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	// Execute all nested steps and aggregate results
	var allResults []types.StepResult
	var firstError error
	
	for i, nestedStep := range step.Steps {
		result, err := s.strategyRouter.Execute(nestedStep, i+1, loopCtx)
		if result != nil {
			allResults = append(allResults, *result)
		}
		
		if err != nil && firstError == nil {
			firstError = err
		}
		
		// Stop on first error unless continue flag is set
		if err != nil && !nestedStep.Continue {
			break
		}
	}
	
	// Create aggregate result
	aggregateResult := &types.StepResult{
		Name:     step.Name,
		Action:   "nested_steps",
		Duration: 0, // Would need to sum durations from allResults
	}
	
	// Set overall status based on nested results
	if firstError != nil {
		aggregateResult.Result = types.ActionResult{
			Status: constants.ActionStatusError,
			// Would set error info from firstError
		}
	} else {
		aggregateResult.Result = types.ActionResult{
			Status: constants.ActionStatusPassed,
		}
	}
	
	return aggregateResult, firstError
}

// CanHandle returns true for steps that have nested steps
func (s *NestedStepsExecutionStrategy) CanHandle(step types.Step) bool {
	return len(step.Steps) > 0
}

// Priority returns high priority as nested steps are specific
func (s *NestedStepsExecutionStrategy) Priority() int {
	return 4
}