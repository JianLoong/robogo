package execution

import (
	"github.com/JianLoong/robogo/internal/common"
)

// Dependencies contains all the dependencies needed for execution
type Dependencies struct {
	Variables          *common.Variables
	ConditionEvaluator ConditionEvaluator
	ActionExecutor     ActionExecutor
}

// NewDependencies creates a new dependencies container
func NewDependencies(variables *common.Variables) *Dependencies {
	return &Dependencies{
		Variables:          variables,
		ConditionEvaluator: NewBasicConditionEvaluator(variables),
		ActionExecutor:     NewStepExecutor(variables),
	}
}

// DependencyInjector provides dependency injection for execution components
type DependencyInjector struct {
	deps *Dependencies
}

// NewDependencyInjector creates a new dependency injector
func NewDependencyInjector(deps *Dependencies) *DependencyInjector {
	return &DependencyInjector{
		deps: deps,
	}
}

// GetVariables returns the variables dependency
func (di *DependencyInjector) GetVariables() *common.Variables {
	return di.deps.Variables
}

// GetConditionEvaluator returns the condition evaluator dependency
func (di *DependencyInjector) GetConditionEvaluator() ConditionEvaluator {
	return di.deps.ConditionEvaluator
}

// GetActionExecutor returns the action executor dependency
func (di *DependencyInjector) GetActionExecutor() ActionExecutor {
	return di.deps.ActionExecutor
}

// CreateUnifiedExecutor creates a unified executor with injected dependencies
func (di *DependencyInjector) CreateUnifiedExecutor() *UnifiedExecutor {
	router := NewExecutionStrategyRouter()
	
	// Register strategies with injected dependencies
	router.RegisterStrategy(NewConditionalExecutionStrategy(di.GetConditionEvaluator(), router))
	router.RegisterStrategy(NewRetryExecutionStrategy(di.GetActionExecutor(), di.GetVariables()))
	router.RegisterStrategy(NewNestedStepsExecutionStrategy(router))
	router.RegisterStrategy(NewBasicExecutionStrategy(di.GetActionExecutor()))
	
	return &UnifiedExecutor{
		strategyRouter: router,
	}
}

// CreateRetryExecutor creates a retry executor with injected dependencies
func (di *DependencyInjector) CreateRetryExecutor() *RetryExecutor {
	return NewRetryExecutor(di.GetActionExecutor(), di.GetVariables())
}

// WithVariables creates a new dependency injector with different variables
func (di *DependencyInjector) WithVariables(variables *common.Variables) *DependencyInjector {
	newDeps := &Dependencies{
		Variables:          variables,
		ConditionEvaluator: NewBasicConditionEvaluator(variables),
		ActionExecutor:     NewStepExecutor(variables),
	}
	return NewDependencyInjector(newDeps)
}