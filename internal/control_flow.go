package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// ControlFlowExecutor handles control flow execution for test steps
type ControlFlowExecutor struct {
	variables          *common.Variables
	conditionEvaluator *ConditionEvaluator
}

// NewControlFlowExecutor creates a new control flow executor
func NewControlFlowExecutor(variables *common.Variables) *ControlFlowExecutor {
	return &ControlFlowExecutor{
		variables:          variables,
		conditionEvaluator: NewConditionEvaluator(variables),
	}
}

// ExecuteStepWithControlFlow executes a step with if/for/while support
func (executor *ControlFlowExecutor) ExecuteStepWithControlFlow(step types.Step, stepNum int) ([]types.StepResult, error) {
	// Handle for loop first (if with for will be handled inside the loop)
	if step.For != "" {
		return executor.executeStepForLoop(step, stepNum)
	}

	// Handle while loop
	if step.While != "" {
		return executor.executeStepWhileLoop(step, stepNum)
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
					WithTemplate("if condition evaluation failed: %v").
					WithContext("condition", condition).
					WithContext("error", err.Error()).
					Build(err),
			}
			return []types.StepResult{*stepResult}, err
		}
		if !shouldExecute {
			// Skip step
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.NewSkippedResult(fmt.Sprintf("Skipped due to if condition: %s", condition)),
			}
			return []types.StepResult{*stepResult}, nil
		}
	}

	// Regular execution
	stepResult, err := executor.executeStep(step, stepNum)
	return []types.StepResult{*stepResult}, err
}

// executeStepForLoop executes a step in a for loop
func (executor *ControlFlowExecutor) executeStepForLoop(step types.Step, stepNum int) ([]types.StepResult, error) {
	rangeOrArray := executor.variables.Substitute(step.For)
	iterations, stepResult, err := executor.parseIterations(rangeOrArray, step)
	if err != nil {
		return []types.StepResult{*stepResult}, err
	}

	return executor.executeIterations(step, stepNum, iterations)
}

// parseIterations parses the for loop specification into a slice of iterations.
// It supports three formats:
// - Range: "1..5" creates iterations [1, 2, 3, 4, 5]
// - Array: "[item1,item2,item3]" creates iterations ["item1", "item2", "item3"]
// - Count: "3" creates iterations [1, 2, 3]
func (executor *ControlFlowExecutor) parseIterations(rangeOrArray string, step types.Step) ([]any, *types.StepResult, error) {
	if strings.Contains(rangeOrArray, "..") {
		return executor.parseRange(rangeOrArray, step)
	} else if strings.HasPrefix(rangeOrArray, "[") && strings.HasSuffix(rangeOrArray, "]") {
		return executor.parseArray(rangeOrArray, step)
	} else {
		return executor.parseCount(rangeOrArray, step)
	}
}

// parseRange parses a range specification like "1..5" and returns integers from start to end inclusive.
// Returns an error result if the range format is invalid or contains non-numeric values.
func (executor *ControlFlowExecutor) parseRange(rangeSpec string, step types.Step) ([]any, *types.StepResult, error) {
	parts := strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_RANGE_FORMAT").
				WithTemplate("invalid range format: %s").
				WithContext("range_spec", rangeSpec).
				Build(rangeSpec),
		}, fmt.Errorf("invalid range format: %s", rangeSpec)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_START_VALUE").
				WithTemplate("invalid start value: %s").
				WithContext("start_value", parts[0]).
				Build(parts[0]),
		}, fmt.Errorf("invalid start value: %s", parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_END_VALUE").
				WithTemplate("invalid end value: %s").
				WithContext("end_value", parts[1]).
				Build(parts[1]),
		}, fmt.Errorf("invalid end value: %s", parts[1])
	}

	var iterations []any
	for i := start; i <= end; i++ {
		iterations = append(iterations, i)
	}
	return iterations, nil, nil
}

// parseArray parses an array specification like "[item1,item2,item3]" and returns the items as strings.
// Items are trimmed of whitespace and returned in the order they appear.
func (executor *ControlFlowExecutor) parseArray(arraySpec string, step types.Step) ([]any, *types.StepResult, error) {
	arrayStr := arraySpec[1 : len(arraySpec)-1]
	items := strings.Split(arrayStr, ",")
	var iterations []any
	for _, item := range items {
		iterations = append(iterations, strings.TrimSpace(item))
	}
	return iterations, nil, nil
}

// parseCount parses a count specification like "3" and returns integers from 1 to count inclusive.
// Returns an error result if the count is not a valid integer.
func (executor *ControlFlowExecutor) parseCount(countSpec string, step types.Step) ([]any, *types.StepResult, error) {
	count, err := strconv.Atoi(countSpec)
	if err != nil {
		return nil, &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_COUNT_FORMAT").
				WithTemplate("invalid count format: %s").
				WithContext("count_spec", countSpec).
				Build(countSpec),
		}, fmt.Errorf("invalid count format: %s", countSpec)
	}

	var iterations []any
	for i := 1; i <= count; i++ {
		iterations = append(iterations, i)
	}
	return iterations, nil, nil
}

// executeIterations executes the step for each iteration, setting loop variables.
// Sets the following variables for each iteration:
// - iteration: 1-based iteration number
// - index: 0-based iteration index
// - item: current iteration value
func (executor *ControlFlowExecutor) executeIterations(step types.Step, stepNum int, iterations []any) ([]types.StepResult, error) {
	var results []types.StepResult
	for i, item := range iterations {
		// Set loop variables
		executor.variables.Set(constants.LoopVariableIteration, i+1)
		executor.variables.Set(constants.LoopVariableIndex, i)
		executor.variables.Set(constants.LoopVariableItem, item)

		// Check if condition within loop
		if step.If != "" {
			condition := executor.variables.Substitute(step.If)
			shouldExecute, err := executor.conditionEvaluator.Evaluate(condition)
			if err != nil {
				stepResult := &types.StepResult{
					Name:   fmt.Sprintf("%s (iteration %d)", step.Name, i+1),
					Action: step.Action,
					Result: types.NewErrorBuilder(types.ErrorCategoryExecution, "IF_CONDITION_FAILED").
						WithTemplate("if condition evaluation failed: %v").
						WithContext("condition", condition).
						WithContext("iteration", i+1).
						WithContext("error", err.Error()).
						Build(err),
				}
				results = append(results, *stepResult)
				return results, err
			}
			if !shouldExecute {
				// Skip this iteration
				stepResult := &types.StepResult{
					Name:   fmt.Sprintf("%s (iteration %d)", step.Name, i+1),
					Action: step.Action,
					Result: types.NewSkippedResult(fmt.Sprintf("Skipped due to if condition: %s", condition)),
				}
				results = append(results, *stepResult)
				continue
			}
		}

		stepResult, err := executor.executeStep(step, stepNum)
		stepResult.Name = fmt.Sprintf("%s (iteration %d)", step.Name, i+1)
		results = append(results, *stepResult)

		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// executeStepWhileLoop executes a step in a while loop until the condition becomes false.
// The loop is limited to a maximum number of iterations to prevent infinite loops.
// Sets the 'iteration' variable for each iteration starting from 1.
func (executor *ControlFlowExecutor) executeStepWhileLoop(step types.Step, stepNum int) ([]types.StepResult, error) {
	const maxIterations = constants.MaxWhileLoopIterations
	iterations := 0
	var results []types.StepResult

	for iterations < maxIterations {
		iterations++
		executor.variables.Set(constants.LoopVariableIteration, iterations)

		// Evaluate condition
		condition := executor.variables.Substitute(step.While)
		shouldContinue, err := executor.conditionEvaluator.Evaluate(condition)
		if err != nil {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.NewErrorBuilder(types.ErrorCategoryExecution, "WHILE_CONDITION_FAILED").
					WithTemplate("while condition evaluation failed: %v").
					WithContext("condition", condition).
					WithContext("iteration", iterations).
					WithContext("error", err.Error()).
					Build(err),
			}
			return append(results, *stepResult), err
		}

		if !shouldContinue {
			break
		}

		stepResult, err := executor.executeStep(step, stepNum)
		stepResult.Name = fmt.Sprintf("%s (while iteration %d)", step.Name, iterations)
		results = append(results, *stepResult)

		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// executeStep executes a single step, performing variable substitution and action execution.
// Returns the step result and an error if the action fails with an error status.
// The error is used for control flow purposes to stop execution on critical failures.
func (executor *ControlFlowExecutor) executeStep(step types.Step, stepNum int) (*types.StepResult, error) {
	start := time.Now()

	result := &types.StepResult{
		Name:   step.Name,
		Action: step.Action,
		Result: types.ActionResult{Status: types.ActionStatusError},
	}

	// Get action
	action, exists := actions.GetAction(step.Action)
	if !exists {
		result.Result = types.NewErrorBuilder(types.ErrorCategoryValidation, "UNKNOWN_ACTION").
			WithTemplate("unknown action: %s").
			WithContext("action_name", step.Action).
			Build(step.Action)
		result.Duration = time.Since(start)
		return result, fmt.Errorf("unknown action: %s", step.Action)
	}

	// Substitute variables in arguments
	args := executor.variables.SubstituteArgs(step.Args)

	// Substitute variables in options
	options := make(map[string]any)
	for k, v := range step.Options {
		if str, ok := v.(string); ok {
			options[k] = executor.variables.Substitute(str)
		} else {
			options[k] = v
		}
	}

	// Execute action
	output := action(args, options, executor.variables)
	result.Duration = time.Since(start)

	// Use the ActionResult as is
	result.Result = output

	// Return error only if the action status is error (for control flow purposes)
	if output.Status == types.ActionStatusError {
		return result, fmt.Errorf("action failed: %s", output.GetErrorMessage())
	}

	// Store result variable if specified
	if step.Result != "" {
		executor.variables.Set(step.Result, output.Data)
	}

	return result, nil
}
