package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
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
func (cfe *ControlFlowExecutor) ExecuteStepWithControlFlow(step types.Step, stepNum int) ([]types.StepResult, error) {
	// Handle for loop first (if with for will be handled inside the loop)
	if step.For != "" {
		return cfe.executeStepForLoop(step, stepNum)
	}

	// Handle while loop
	if step.While != "" {
		return cfe.executeStepWhileLoop(step, stepNum)
	}

	// Handle if condition (only for non-loop steps)
	if step.If != "" {
		condition := cfe.variables.Substitute(step.If)
		shouldExecute, err := cfe.conditionEvaluator.Evaluate(condition)
		if err != nil {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("if condition evaluation failed: %v", err)},
			}
			return []types.StepResult{*stepResult}, err
		}
		if !shouldExecute {
			// Skip step
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.ActionResult{Status: types.ActionStatusSkipped, Reason: fmt.Sprintf("Skipped due to if condition: %s", condition)},
			}
			return []types.StepResult{*stepResult}, nil
		}
	}

	// Regular execution
	stepResult, err := cfe.executeStep(step, stepNum)
	return []types.StepResult{*stepResult}, err
}

// executeStepForLoop executes a step in a for loop
func (cfe *ControlFlowExecutor) executeStepForLoop(step types.Step, stepNum int) ([]types.StepResult, error) {
	rangeOrArray := cfe.variables.Substitute(step.For)
	var iterations []interface{}

	// Parse range, array, or count
	if strings.Contains(rangeOrArray, "..") {
		// Range: "1..5"
		parts := strings.Split(rangeOrArray, "..")
		if len(parts) != 2 {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("invalid range format: %s", rangeOrArray)},
			}
			return []types.StepResult{*stepResult}, fmt.Errorf("invalid range format: %s", rangeOrArray)
		}
		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("invalid start value in range: %s", parts[0])},
			}
			return []types.StepResult{*stepResult}, err
		}
		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("invalid end value in range: %s", parts[1])},
			}
			return []types.StepResult{*stepResult}, err
		}
		for i := start; i <= end; i++ {
			iterations = append(iterations, i)
		}
	} else if strings.HasPrefix(rangeOrArray, "[") && strings.HasSuffix(rangeOrArray, "]") {
		// Array: "[item1,item2,item3]"
		arrayStr := rangeOrArray[1 : len(rangeOrArray)-1]
		items := strings.Split(arrayStr, ",")
		for _, item := range items {
			iterations = append(iterations, strings.TrimSpace(item))
		}
	} else {
		// Count: "3"
		count, err := strconv.Atoi(rangeOrArray)
		if err != nil {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("invalid count format: %s", rangeOrArray)},
			}
			return []types.StepResult{*stepResult}, err
		}
		for i := 1; i <= count; i++ {
			iterations = append(iterations, i)
		}
	}

	// Execute step for each iteration
	var results []types.StepResult
	for i, item := range iterations {
		// Set loop variables
		cfe.variables.Set("iteration", i+1)
		cfe.variables.Set("index", i)
		cfe.variables.Set("item", item)

		// Check if condition within loop
		if step.If != "" {
			condition := cfe.variables.Substitute(step.If)
			shouldExecute, err := cfe.conditionEvaluator.Evaluate(condition)
			if err != nil {
				stepResult := &types.StepResult{
					Name:   fmt.Sprintf("%s (iteration %d)", step.Name, i+1),
					Action: step.Action,
					Result: types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("if condition evaluation failed: %v", err)},
				}
				results = append(results, *stepResult)
				return results, err
			}
			if !shouldExecute {
				// Skip this iteration
				stepResult := &types.StepResult{
					Name:   fmt.Sprintf("%s (iteration %d)", step.Name, i+1),
					Action: step.Action,
					Result: types.ActionResult{Status: types.ActionStatusSkipped, Reason: fmt.Sprintf("Skipped due to if condition: %s", condition)},
				}
				results = append(results, *stepResult)
				continue
			}
		}

		stepResult, err := cfe.executeStep(step, stepNum)
		stepResult.Name = fmt.Sprintf("%s (iteration %d)", step.Name, i+1)
		results = append(results, *stepResult)

		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// executeStepWhileLoop executes a step in a while loop
func (cfe *ControlFlowExecutor) executeStepWhileLoop(step types.Step, stepNum int) ([]types.StepResult, error) {
	maxIterations := 10 // Default max iterations
	iterations := 0
	var results []types.StepResult

	for iterations < maxIterations {
		iterations++
		cfe.variables.Set("iteration", iterations)

		// Evaluate condition
		condition := cfe.variables.Substitute(step.While)
		shouldContinue, err := cfe.conditionEvaluator.Evaluate(condition)
		if err != nil {
			stepResult := &types.StepResult{
				Name:   step.Name,
				Action: step.Action,
				Result: types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("while condition evaluation failed: %v", err)},
			}
			return append(results, *stepResult), err
		}

		if !shouldContinue {
			break
		}

		stepResult, err := cfe.executeStep(step, stepNum)
		stepResult.Name = fmt.Sprintf("%s (while iteration %d)", step.Name, iterations)
		results = append(results, *stepResult)

		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// executeStep executes a single step (extracted from TestRunner)
func (cfe *ControlFlowExecutor) executeStep(step types.Step, stepNum int) (*types.StepResult, error) {
	start := time.Now()

	result := &types.StepResult{
		Name:   step.Name,
		Action: step.Action,
		Result: types.ActionResult{Status: types.ActionStatusError},
	}

	// Get action
	action, exists := actions.GetAction(step.Action)
	if !exists {
		result.Result = types.ActionResult{Status: types.ActionStatusError, Error: fmt.Sprintf("unknown action: %s", step.Action)}
		result.Duration = time.Since(start)
		return result, fmt.Errorf("unknown action: %s", step.Action)
	}

	// Substitute variables in arguments
	args := cfe.variables.SubstituteArgs(step.Args)

	// Substitute variables in options
	options := make(map[string]interface{})
	for k, v := range step.Options {
		if str, ok := v.(string); ok {
			options[k] = cfe.variables.Substitute(str)
		} else {
			options[k] = v
		}
	}

	// Execute action
	output, err := action(args, options, cfe.variables)
	result.Duration = time.Since(start)

	if err != nil {
		// If the action returned an ActionResult with Status 'error', propagate as error
		if output.Status == types.ActionStatusError {
			result.Result = output
			return result, err
		}
		// If the action returned an ActionResult with Status 'failed', propagate as failed
		if output.Status == types.ActionStatusFailed {
			result.Result = output
			return result, nil
		}
		// Otherwise, treat as error
		result.Result = types.ActionResult{Status: types.ActionStatusError, Error: err.Error()}
		return result, err
	}

	// Use the ActionResult as is
	result.Result = output

	// Store result variable if specified
	if step.Result != "" {
		cfe.variables.Set(step.Result, output.Data)
	}

	return result, nil
}
