package execution

import (
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// ExecutionPipeline orchestrates the complete step execution process
type ExecutionPipeline struct {
	dependencies *Dependencies
	injector     *DependencyInjector
	executor     *UnifiedExecutor
}

// NewExecutionPipeline creates a new execution pipeline with clean dependencies
func NewExecutionPipeline(variables *common.Variables) *ExecutionPipeline {
	deps := NewDependencies(variables)
	injector := NewDependencyInjector(deps)
	executor := injector.CreateUnifiedExecutor()
	
	return &ExecutionPipeline{
		dependencies: deps,
		injector:     injector,
		executor:     executor,
	}
}

// Execute executes a single step through the pipeline
func (pipeline *ExecutionPipeline) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	return pipeline.executor.Execute(step, stepNum, loopCtx)
}

// ExecuteSteps executes multiple steps through the pipeline
func (pipeline *ExecutionPipeline) ExecuteSteps(steps []types.Step, loopCtx *types.LoopContext) ([]types.StepResult, error) {
	return pipeline.executor.ExecuteSteps(steps, loopCtx)
}

// ExecuteStepWithControlFlow executes a single step and returns results in the old format
// for backward compatibility with the runner
func (pipeline *ExecutionPipeline) ExecuteStepWithControlFlow(step types.Step, stepNum int) ([]types.StepResult, error) {
	result, err := pipeline.Execute(step, stepNum, nil)
	if result == nil {
		return []types.StepResult{}, err
	}
	
	// For now, return a single result in an array
	// In the future, this could handle nested steps that return multiple results
	return []types.StepResult{*result}, err
}

// GetVariables returns the pipeline's variable manager
func (pipeline *ExecutionPipeline) GetVariables() *common.Variables {
	return pipeline.dependencies.Variables
}

// Clone creates a new pipeline with cloned variables (useful for isolated execution)
func (pipeline *ExecutionPipeline) Clone() *ExecutionPipeline {
	clonedVars := pipeline.dependencies.Variables.Clone()
	return NewExecutionPipeline(clonedVars)
}

// WithVariables creates a new pipeline with different variables
func (pipeline *ExecutionPipeline) WithVariables(variables *common.Variables) *ExecutionPipeline {
	return NewExecutionPipeline(variables)
}

// GetExecutionStrategies returns information about available execution strategies
func (pipeline *ExecutionPipeline) GetExecutionStrategies(step types.Step) []ExecutionStrategy {
	return pipeline.executor.GetApplicableStrategies(step)
}

// PipelineBuilder provides a fluent interface for building execution pipelines
type PipelineBuilder struct {
	variables *common.Variables
}

// NewPipelineBuilder creates a new pipeline builder
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		variables: common.NewVariables(),
	}
}

// WithVariables sets the variables for the pipeline
func (builder *PipelineBuilder) WithVariables(variables *common.Variables) *PipelineBuilder {
	builder.variables = variables
	return builder
}

// WithVariableMap loads variables from a map
func (builder *PipelineBuilder) WithVariableMap(vars map[string]any) *PipelineBuilder {
	if builder.variables == nil {
		builder.variables = common.NewVariables()
	}
	builder.variables.Load(vars)
	return builder
}

// Build creates the execution pipeline
func (builder *PipelineBuilder) Build() *ExecutionPipeline {
	if builder.variables == nil {
		builder.variables = common.NewVariables()
	}
	return NewExecutionPipeline(builder.variables)
}