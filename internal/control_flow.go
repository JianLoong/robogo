package internal

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/execution"
	"github.com/JianLoong/robogo/internal/loops"
	"github.com/JianLoong/robogo/internal/templates"
	"github.com/JianLoong/robogo/internal/types"
)

// ControlFlowExecutor handles control flow execution for test steps
type ControlFlowExecutor struct {
	variables          *common.Variables
	conditionEvaluator *ConditionEvaluator
	stepExecutor       *execution.StepExecutorImpl
	retryExecutor      *execution.RetryExecutor
	loopExecutor       *loops.LoopExecutor
}

// NewControlFlowExecutor creates a new control flow executor
func NewControlFlowExecutor(variables *common.Variables) *ControlFlowExecutor {
	conditionEvaluator := NewConditionEvaluator(variables)
	stepExecutor := execution.NewStepExecutor(variables)
	retryExecutor := execution.NewRetryExecutor(stepExecutor)
	loopExecutor := loops.NewLoopExecutor(variables, conditionEvaluator, &stepExecutorAdapter{stepExecutor, retryExecutor})

	return &ControlFlowExecutor{
		variables:          variables,
		conditionEvaluator: conditionEvaluator,
		stepExecutor:       stepExecutor,
		retryExecutor:      retryExecutor,
		loopExecutor:       loopExecutor,
	}
}

// stepExecutorAdapter implements the loops.StepExecutor interface
type stepExecutorAdapter struct {
	stepExecutor  *execution.StepExecutorImpl
	retryExecutor *execution.RetryExecutor
}

func (adapter *stepExecutorAdapter) ExecuteStepWithContext(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	// If retry is configured, use retry logic
	if step.Retry != nil {
		return adapter.retryExecutor.ExecuteStepWithRetry(step, stepNum, loopCtx)
	}

	// Otherwise, execute normally
	return adapter.stepExecutor.ExecuteSingleStep(step, stepNum, loopCtx)
}

// ExecuteStepWithControlFlow executes a step with if/for/while support
func (executor *ControlFlowExecutor) ExecuteStepWithControlFlow(step types.Step, stepNum int) ([]types.StepResult, error) {
	// Handle for loop first (if with for will be handled inside the loop)
	if step.For != "" {
		return executor.loopExecutor.ExecuteStepForLoop(step, stepNum)
	}

	// Handle while loop
	if step.While != "" {
		return executor.loopExecutor.ExecuteStepWhileLoop(step, stepNum)
	}

	// Handle if condition (only for non-loop steps)
	if step.If != "" {
		condition := executor.variables.Substitute(step.If)
		shouldExecute, err := executor.conditionEvaluator.Evaluate(condition)
		if err != nil {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.NewErrorBuilder(types.ErrorCategoryExecution, "IF_CONDITION_FAILED").
					WithTemplate(templates.GetTemplateConstant(constants.TemplateIfConditionFailed)).
					WithContext("condition", condition).
					WithContext("error", err.Error()).
					Build(err),
			}
			return []types.StepResult{*stepResult}, err
		}
		if !shouldExecute {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.NewSkippedResult(fmt.Sprintf("Skipped due to if condition: %s", condition)),
			}
			return []types.StepResult{*stepResult}, nil
		}
	}

	// Execute single step
	stepResult, err := executor.executeStepWithContext(step, stepNum, nil)
	return []types.StepResult{*stepResult}, err
}

// executeStepWithContext executes a step with optional loop context and retry support
func (executor *ControlFlowExecutor) executeStepWithContext(
	step types.Step,
	stepNum int,
	loopCtx *types.LoopContext,
) (*types.StepResult, error) {
	// If retry is configured, use retry logic
	if step.Retry != nil {
		return executor.retryExecutor.ExecuteStepWithRetry(step, stepNum, loopCtx)
	}

	// Otherwise, execute normally
	return executor.stepExecutor.ExecuteSingleStep(step, stepNum, loopCtx)
}
