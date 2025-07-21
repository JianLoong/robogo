package execution

import (
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

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