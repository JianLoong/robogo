package execution

import (
	"sort"

	"github.com/JianLoong/robogo/internal/types"
)

// ExecutionStrategyRouter coordinates between different execution strategies
type ExecutionStrategyRouter struct {
	strategies []ExecutionStrategy
}

// NewExecutionStrategyRouter creates a new strategy router
func NewExecutionStrategyRouter() *ExecutionStrategyRouter {
	return &ExecutionStrategyRouter{
		strategies: make([]ExecutionStrategy, 0),
	}
}

// RegisterStrategy registers a new execution strategy
func (r *ExecutionStrategyRouter) RegisterStrategy(strategy ExecutionStrategy) {
	r.strategies = append(r.strategies, strategy)
	
	// Sort strategies by priority (descending - higher priority first)
	sort.Slice(r.strategies, func(i, j int) bool {
		return r.strategies[i].Priority() > r.strategies[j].Priority()
	})
}

// Execute selects the appropriate strategy and executes the step
func (r *ExecutionStrategyRouter) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) *types.StepResult {
	// Find the first strategy that can handle this step
	for _, strategy := range r.strategies {
		if strategy.CanHandle(step) {
			return strategy.Execute(step, stepNum, loopCtx)
		}
	}
	
	// No strategy found - return error result
	return &types.StepResult{
		Name:   step.Name,
		Action: step.Action,
		Result: types.NewErrorBuilder(types.ErrorCategoryExecution, "NO_STRATEGY_FOUND").
			WithTemplate("No execution strategy found for step: %s").
			WithContext("step_name", step.Name).
			WithContext("action", step.Action).
			Build(step.Name),
	}
}

// GetApplicableStrategies returns all strategies that can handle the given step
func (r *ExecutionStrategyRouter) GetApplicableStrategies(step types.Step) []ExecutionStrategy {
	var applicable []ExecutionStrategy
	for _, strategy := range r.strategies {
		if strategy.CanHandle(step) {
			applicable = append(applicable, strategy)
		}
	}
	return applicable
}

// GetRegisteredStrategies returns all registered strategies
func (r *ExecutionStrategyRouter) GetRegisteredStrategies() []ExecutionStrategy {
	return append([]ExecutionStrategy{}, r.strategies...)
}