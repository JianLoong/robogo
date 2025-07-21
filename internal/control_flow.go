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
	retryExecutor := execution.NewRetryExecutor(stepExecutor, variables)
	loopExecutor := loops.NewLoopExecutor(variables, conditionEvaluator, &stepExecutorAdapter{stepExecutor, retryExecutor, variables})

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
	variables     *common.Variables
}

func (adapter *stepExecutorAdapter) ExecuteStepWithContext(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	// Handle nested steps (when step.Steps is populated instead of step.Action)
	if len(step.Steps) > 0 {
		// Log the parent step before executing nested steps
		adapter.printParentStepExecution(step, stepNum, loopCtx)
		return adapter.executeNestedSteps(step, stepNum, loopCtx)
	}

	// If retry is configured, use retry logic
	if step.Retry != nil {
		return adapter.retryExecutor.ExecuteStepWithRetry(step, stepNum, loopCtx)
	}

	// Otherwise, execute normally
	return adapter.stepExecutor.ExecuteSingleStep(step, stepNum, loopCtx)
}

// executeNestedSteps executes a group of nested steps and aggregates their results
func (adapter *stepExecutorAdapter) executeNestedSteps(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	var allResults []types.StepResult
	var lastError error
	
	// Execute each nested step
	for _, nestedStep := range step.Steps {
		// Execute the nested step (this will recursively handle if/for/while/retry)
		result, err := adapter.ExecuteStepWithContext(nestedStep, stepNum, loopCtx)
		
		if result != nil {
			// Update step name to show it's nested
			result.Name = fmt.Sprintf("%s -> %s", step.Name, result.Name)
			allResults = append(allResults, *result)
		}
		
		// Store the last error but continue executing remaining steps
		if err != nil {
			lastError = err
			// If continue flag is not set on the nested step, stop here
			if !nestedStep.Continue {
				break
			}
		}
	}
	
	// Aggregate results into a single step result
	if len(allResults) == 0 {
		errorResult := types.NewErrorBuilder(types.ErrorCategoryExecution, "NO_NESTED_RESULTS").
			WithTemplate("No results from nested steps").
			Build(lastError)
		
		return &types.StepResult{
			Name:   step.Name,
			Action: "nested_steps",
			Result: errorResult,
		}, lastError
	}
	
	// Use the last step's result as the primary result, but store all results
	lastResult := allResults[len(allResults)-1]
	
	// Create aggregated result
	aggregatedResult := &types.StepResult{
		Name:   step.Name,
		Action: "nested_steps",
		Result: lastResult.Result, // Use the last step's result status
	}
	
	// If any step failed and continue wasn't set, return error
	if lastError != nil {
		aggregatedResult.Result.Status = constants.ActionStatusError
	}
	
	return aggregatedResult, lastError
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

	// Handle nested steps
	if len(step.Steps) > 0 {
		// Log the parent step before executing nested steps
		executor.printParentStepExecution(step, stepNum, nil)
		return executor.executeNestedSteps(step, stepNum)
	}

	// Execute single step
	stepResult, err := executor.executeStepWithContext(step, stepNum, nil)
	return []types.StepResult{*stepResult}, err
}

// executeNestedSteps executes nested steps recursively using control flow
func (executor *ControlFlowExecutor) executeNestedSteps(step types.Step, stepNum int) ([]types.StepResult, error) {
	var allResults []types.StepResult
	var lastError error
	
	// Execute each nested step with full control flow support
	for i, nestedStep := range step.Steps {
		// Execute the nested step recursively (this handles if/for/while/retry/nested steps)
		stepResults, err := executor.ExecuteStepWithControlFlow(nestedStep, i+1)
		
		// Append all results from the nested step execution
		for _, result := range stepResults {
			// Update step name to show it's nested
			result.Name = fmt.Sprintf("%s -> %s", step.Name, result.Name)
			allResults = append(allResults, result)
		}
		
		// Store the last error but continue executing remaining steps
		if err != nil {
			lastError = err
			// If continue flag is not set on the nested step, stop here
			if !nestedStep.Continue {
				break
			}
		}
	}
	
	// Return all nested step results
	if len(allResults) == 0 {
		// Create an error result if no nested steps were executed
		errorResult := types.NewErrorBuilder(types.ErrorCategoryExecution, "NO_NESTED_RESULTS").
			WithTemplate("No results from nested steps").
			Build(lastError)
		
		return []types.StepResult{{
			Name:   step.Name,
			Action: "nested_steps",
			Result: errorResult,
		}}, lastError
	}
	
	return allResults, lastError
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

// printParentStepExecution prints step execution details for parent steps with nested steps
func (adapter *stepExecutorAdapter) printParentStepExecution(step types.Step, stepNum int, loopCtx *types.LoopContext) {
	fmt.Printf("Step %d: %s\n", stepNum, step.Name)
	fmt.Printf("  Action: nested_steps\n")

	// Show conditions if present
	if step.If != "" {
		condition := adapter.variables.Substitute(step.If)
		fmt.Printf("  If: %s\n", condition)
	}

	if step.For != "" {
		forValue := adapter.variables.Substitute(step.For)
		fmt.Printf("  For: %s\n", forValue)
	}

	if step.While != "" {
		whileValue := adapter.variables.Substitute(step.While)
		fmt.Printf("  While: %s\n", whileValue)
	}

	if step.Result != "" {
		fmt.Printf("  Result Variable: %s\n", step.Result)
	}

	// Show loop context if present
	if loopCtx != nil {
		if loopCtx.Type == "for" {
			fmt.Printf("  Loop Context: for iteration %d (index %d, item: %v)\n", 
				loopCtx.Iteration, loopCtx.Index, loopCtx.Item)
		} else if loopCtx.Type == "while" {
			fmt.Printf("  Loop Context: while iteration %d\n", loopCtx.Iteration)
		}
	}

	fmt.Printf("  Nested Steps: %d\n", len(step.Steps))
	fmt.Println("  Executing nested steps... ")
}

// printParentStepExecution prints step execution details for parent steps with nested steps (for ControlFlowExecutor)
func (executor *ControlFlowExecutor) printParentStepExecution(step types.Step, stepNum int, loopCtx *types.LoopContext) {
	fmt.Printf("Step %d: %s\n", stepNum, step.Name)
	fmt.Printf("  Action: nested_steps\n")

	// Show conditions if present
	if step.If != "" {
		condition := executor.variables.Substitute(step.If)
		fmt.Printf("  If: %s\n", condition)
	}

	if step.For != "" {
		forValue := executor.variables.Substitute(step.For)
		fmt.Printf("  For: %s\n", forValue)
	}

	if step.While != "" {
		whileValue := executor.variables.Substitute(step.While)
		fmt.Printf("  While: %s\n", whileValue)
	}

	if step.Result != "" {
		fmt.Printf("  Result Variable: %s\n", step.Result)
	}

	// Show loop context if present
	if loopCtx != nil {
		if loopCtx.Type == "for" {
			fmt.Printf("  Loop Context: for iteration %d (index %d, item: %v)\n", 
				loopCtx.Iteration, loopCtx.Index, loopCtx.Item)
		} else if loopCtx.Type == "while" {
			fmt.Printf("  Loop Context: while iteration %d\n", loopCtx.Iteration)
		}
	}

	fmt.Printf("  Nested Steps: %d\n", len(step.Steps))
	fmt.Println("  Executing nested steps... ")
}
