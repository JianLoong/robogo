package execution

import (
	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
)

// Dependencies contains all the dependencies needed for execution
type Dependencies struct {
	Variables          *common.Variables
	ActionRegistry     *actions.ActionRegistry
	ConditionEvaluator ConditionEvaluator
}

// NewDependencies creates a new dependencies container
func NewDependencies(variables *common.Variables) *Dependencies {
	actionRegistry := actions.NewActionRegistry()
	
	return &Dependencies{
		Variables:          variables,
		ActionRegistry:     actionRegistry,
		ConditionEvaluator: NewBasicConditionEvaluator(variables),
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


// GetActionRegistry returns the action registry dependency
func (di *DependencyInjector) GetActionRegistry() *actions.ActionRegistry {
	return di.deps.ActionRegistry
}

// CreateUnifiedExecutor creates a unified executor with injected dependencies
func (di *DependencyInjector) CreateUnifiedExecutor() *UnifiedExecutor {
	router := NewExecutionStrategyRouter()
	
	// Register strategies with simplified dependencies - no more ActionExecutor interface
	router.RegisterStrategy(NewConditionalExecutionStrategy(di.GetConditionEvaluator(), router))
	router.RegisterStrategy(NewRetryExecutionStrategy(di.GetVariables(), di.GetActionRegistry()))
	router.RegisterStrategy(NewNestedStepsExecutionStrategy(router))
	router.RegisterStrategy(NewBasicExecutionStrategy(di.GetVariables(), di.GetActionRegistry()))
	
	return &UnifiedExecutor{
		strategyRouter: router,
	}
}

// WithVariables creates a new dependency injector with different variables
func (di *DependencyInjector) WithVariables(variables *common.Variables) *DependencyInjector {
	// Reuse the same action registry but create new condition evaluator with new variables
	newDeps := &Dependencies{
		Variables:          variables,
		ActionRegistry:     di.deps.ActionRegistry, // Reuse the same registry
		ConditionEvaluator: NewBasicConditionEvaluator(variables),
	}
	return NewDependencyInjector(newDeps)
}