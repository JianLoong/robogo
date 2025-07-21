package execution

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// UnifiedExecutor provides a single entry point for all step execution
type UnifiedExecutor struct {
	strategyRouter *ExecutionStrategyRouter
}

// NewUnifiedExecutor creates a new unified executor with all execution strategies
func NewUnifiedExecutor(variables *common.Variables) *UnifiedExecutor {
	router := NewExecutionStrategyRouter()
	
	// Create the basic action executor
	actionExecutor := NewStepExecutor(variables)
	
	// Create condition evaluator
	conditionEvaluator := NewBasicConditionEvaluator(variables)
	
	// Register strategies in order of priority (highest priority first)
	// Note: The router automatically sorts by priority, but registering in logical order
	
	// 1. Loop execution (for/while - highest priority - most specific)
	// For now, delegate to existing control flow to avoid import cycles
	
	// 2. Conditional execution (if statements) 
	router.RegisterStrategy(NewConditionalExecutionStrategy(conditionEvaluator, router))
	
	// 3. Retry execution
	router.RegisterStrategy(NewRetryExecutionStrategy(actionExecutor, variables))
	
	// 4. Nested steps execution
	router.RegisterStrategy(NewNestedStepsExecutionStrategy(router))
	
	// 5. Basic execution (lowest priority - fallback)
	router.RegisterStrategy(NewBasicExecutionStrategy(actionExecutor))
	
	return &UnifiedExecutor{
		strategyRouter: router,
	}
}

// Execute executes a step using the appropriate strategy
func (executor *UnifiedExecutor) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	return executor.strategyRouter.Execute(step, stepNum, loopCtx)
}

// ExecuteStepWithContext implements the StepExecutor interface for compatibility
func (executor *UnifiedExecutor) ExecuteStepWithContext(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	return executor.Execute(step, stepNum, loopCtx)
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