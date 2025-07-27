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
func (s *NestedStepsExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) *types.StepResult {
	// Execute all nested steps and aggregate results
	var allResults []types.StepResult
	var hasError bool
	var firstErrorResult *types.StepResult
	
	for i, nestedStep := range step.Steps {
		result := s.strategyRouter.Execute(nestedStep, i+1, loopCtx)
		if result != nil {
			allResults = append(allResults, *result)
			
			// Check if this step had an error
			if result.Result.Status == constants.ActionStatusError || result.Result.Status == constants.ActionStatusFailed {
				if !hasError {
					hasError = true
					firstErrorResult = result
				}
				
				// Stop on first error unless continue flag is set
				if !nestedStep.Continue {
					break
				}
			}
		}
	}
	
	// Determine if step should be included in summary (default: true)
	includeSummary := true
	if step.Summary != nil {
		includeSummary = *step.Summary
	}

	// Create aggregate result
	aggregateResult := &types.StepResult{
		Name:           step.Name,
		Action:         "nested_steps",
		Duration:       0, // Could sum durations from allResults if needed
		IncludeSummary: includeSummary,
	}
	
	// Set overall status based on nested results
	if hasError && firstErrorResult != nil {
		// Copy error information from first failed step
		aggregateResult.Result = types.ActionResult{
			Status:      firstErrorResult.Result.Status,
			ErrorInfo:   firstErrorResult.Result.ErrorInfo,
			FailureInfo: firstErrorResult.Result.FailureInfo,
		}
	} else {
		aggregateResult.Result = types.ActionResult{
			Status: constants.ActionStatusPassed,
		}
	}
	
	return aggregateResult
}

// CanHandle returns true for steps that have nested steps
func (s *NestedStepsExecutionStrategy) CanHandle(step types.Step) bool {
	return len(step.Steps) > 0
}

// Priority returns medium priority as nested steps are specific
func (s *NestedStepsExecutionStrategy) Priority() int {
	return 2
}