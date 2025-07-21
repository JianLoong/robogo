package execution

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// UnifiedExecutor provides a single entry point for all step execution
type UnifiedExecutor struct {
	strategyRouter *ExecutionStrategyRouter
}

// NewUnifiedExecutor creates a new unified executor using dependency injection
func NewUnifiedExecutor(variables *common.Variables) *UnifiedExecutor {
	// Use dependency injection for clean architecture
	deps := NewDependencies(variables)
	injector := NewDependencyInjector(deps)
	return injector.CreateUnifiedExecutor()
}

// Execute executes a step using the appropriate strategy
func (executor *UnifiedExecutor) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	return executor.strategyRouter.Execute(step, stepNum, loopCtx)
}


// ExecuteSteps executes multiple steps sequentially
func (executor *UnifiedExecutor) ExecuteSteps(steps []types.Step, loopCtx *types.LoopContext) ([]types.StepResult, error) {
	var results []types.StepResult
	var firstError error
	
	for i, step := range steps {
		result, err := executor.Execute(step, i+1, loopCtx)
		if result != nil {
			results = append(results, *result)
		}
		
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			
			// Stop on first error unless continue flag is set
			if !step.Continue {
				break
			}
		}
	}
	
	return results, firstError
}

// GetStrategyRouter returns the internal strategy router for advanced usage
func (executor *UnifiedExecutor) GetStrategyRouter() *ExecutionStrategyRouter {
	return executor.strategyRouter
}

// GetApplicableStrategies returns all strategies that can handle the given step
func (executor *UnifiedExecutor) GetApplicableStrategies(step types.Step) []ExecutionStrategy {
	return executor.strategyRouter.GetApplicableStrategies(step)
}