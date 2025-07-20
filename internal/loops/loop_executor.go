package loops

import (
	"fmt"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/templates"
	"github.com/JianLoong/robogo/internal/types"
)

// LoopExecutor handles execution of for and while loops
type LoopExecutor struct {
	variables          *common.Variables
	conditionEvaluator ConditionEvaluator
	stepExecutor       StepExecutor
	parser             *LoopParser
}

// ConditionEvaluator interface for condition evaluation
type ConditionEvaluator interface {
	Evaluate(condition string) (bool, error)
}

// StepExecutor interface for step execution
type StepExecutor interface {
	ExecuteStepWithContext(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error)
}

// NewLoopExecutor creates a new loop executor
func NewLoopExecutor(variables *common.Variables, conditionEvaluator ConditionEvaluator, stepExecutor StepExecutor) *LoopExecutor {
	return &LoopExecutor{
		variables:          variables,
		conditionEvaluator: conditionEvaluator,
		stepExecutor:       stepExecutor,
		parser:             NewLoopParser(),
	}
}

// ExecuteStepForLoop executes a step in a for loop
func (executor *LoopExecutor) ExecuteStepForLoop(step types.Step, stepNum int) ([]types.StepResult, error) {
	rangeOrArray := executor.variables.Substitute(step.For)
	iterations, stepResult, err := executor.parser.ParseIterations(rangeOrArray, step)
	if err != nil {
		return []types.StepResult{*stepResult}, err
	}

	return executor.executeIterations(step, stepNum, iterations)
}

// ExecuteStepWhileLoop executes a step in a while loop until the condition becomes false.
// The loop is limited to a maximum number of iterations to prevent infinite loops.
// Sets the 'iteration' variable for each iteration starting from 1.
func (executor *LoopExecutor) ExecuteStepWhileLoop(step types.Step, stepNum int) ([]types.StepResult, error) {
	const maxIterations = constants.MaxWhileLoopIterations
	iterations := 0
	var results []types.StepResult
	whileCondition := executor.variables.Substitute(step.While)

	for iterations < maxIterations {
		iterations++

		// Set iteration variable
		executor.variables.Set(constants.LoopVariableIteration, iterations)

		// Create loop context for this iteration
		loopCtx := types.NewWhileLoopContext(iterations, whileCondition, constants.MaxWhileLoopIterations)

		// Evaluate condition
		condition := executor.variables.Substitute(step.While)
		shouldExecute, err := executor.conditionEvaluator.Evaluate(condition)
		if err != nil {
			// Create step context for error enrichment
			stepContext := types.NewStepContext(stepNum, step.Name, step.Action).
				WithLoopContext(loopCtx)

			builder := types.NewErrorBuilder(types.ErrorCategoryExecution, "WHILE_CONDITION_FAILED").
				WithTemplate(templates.GetTemplateConstant(constants.TemplateWhileConditionFailed))

			// Add step context to error
			for key, value := range stepContext.ToMap() {
				builder.WithContext(key, value)
			}

			stepResult := &types.StepResult{
				Name:   fmt.Sprintf("%s (iteration %d)", step.Name, iterations),
				Action: step.Action,
				Result: builder.WithContext("condition", condition).
					WithContext("error", err.Error()).
					Build(err),
			}
			results = append(results, *stepResult)
			return results, err
		}

		if !shouldExecute {
			break
		}

		// Check if condition within loop
		if step.If != "" {
			ifCondition := executor.variables.Substitute(step.If)
			shouldExecuteIf, err := executor.conditionEvaluator.Evaluate(ifCondition)
			if err != nil {
				// Create step context for error enrichment
				stepContext := types.NewStepContext(stepNum, step.Name, step.Action).
					WithLoopContext(loopCtx)

				builder := types.NewErrorBuilder(types.ErrorCategoryExecution, "IF_CONDITION_FAILED").
					WithTemplate(templates.GetTemplateConstant(constants.TemplateIfConditionFailed))

				// Add step context to error
				for key, value := range stepContext.ToMap() {
					builder.WithContext(key, value)
				}

				stepResult := &types.StepResult{
					Name:   fmt.Sprintf("%s (iteration %d)", step.Name, iterations),
					Action: step.Action,
					Result: builder.WithContext("condition", ifCondition).
						WithContext("error", err.Error()).
						Build(err),
				}
				results = append(results, *stepResult)
				return results, err
			}
			if !shouldExecuteIf {
				// Skip this iteration
				stepResult := &types.StepResult{
					Name:   fmt.Sprintf("%s (iteration %d)", step.Name, iterations),
					Action: step.Action,
					Result: types.NewSkippedResult(fmt.Sprintf("Skipped due to if condition: %s", ifCondition)),
				}
				results = append(results, *stepResult)
				continue
			}
		}

		stepResult, err := executor.stepExecutor.ExecuteStepWithContext(step, stepNum, loopCtx)
		stepResult.Name = fmt.Sprintf("%s (iteration %d)", step.Name, iterations)
		results = append(results, *stepResult)

		if err != nil {
			return results, err
		}
	}

	if iterations >= maxIterations {
		// Create step context for error enrichment
		stepContext := types.NewStepContext(stepNum, step.Name, step.Action)

		builder := types.NewErrorBuilder(types.ErrorCategoryExecution, "MAX_WHILE_ITERATIONS").
			WithTemplate("Maximum while loop iterations exceeded")

		// Add step context to error
		for key, value := range stepContext.ToMap() {
			builder.WithContext(key, value)
		}

		stepResult := &types.StepResult{
			Name:   step.Name,
			Action: step.Action,
			Result: builder.WithContext("max_iterations", maxIterations).
				WithContext("condition", executor.variables.Substitute(step.While)).
				Build(fmt.Sprintf("exceeded maximum iterations (%d)", maxIterations)),
		}
		results = append(results, *stepResult)
		return results, fmt.Errorf("while loop exceeded maximum iterations (%d)", maxIterations)
	}

	return results, nil
}

// executeIterations executes the step for each iteration, setting loop variables.
// Sets the following variables for each iteration:
// - iteration: 1-based iteration number
// - index: 0-based iteration index
// - item: current iteration value
func (executor *LoopExecutor) executeIterations(step types.Step, stepNum int, iterations []any) ([]types.StepResult, error) {
	var results []types.StepResult
	forCondition := executor.variables.Substitute(step.For)

	for i, item := range iterations {
		// Set loop variables
		executor.variables.Set(constants.LoopVariableIteration, i+1)
		executor.variables.Set(constants.LoopVariableIndex, i)
		executor.variables.Set(constants.LoopVariableItem, item)

		// Create loop context for this iteration
		loopCtx := types.NewForLoopContext(i+1, i, item, forCondition)

		// Check if condition within loop
		if step.If != "" {
			condition := executor.variables.Substitute(step.If)
			shouldExecute, err := executor.conditionEvaluator.Evaluate(condition)
			if err != nil {
				// Create step context for error enrichment
				stepContext := types.NewStepContext(stepNum, step.Name, step.Action).
					WithLoopContext(loopCtx)

				builder := types.NewErrorBuilder(types.ErrorCategoryExecution, "IF_CONDITION_FAILED").
					WithTemplate(templates.GetTemplateConstant(constants.TemplateIfConditionFailed))

				// Add step context to error
				for key, value := range stepContext.ToMap() {
					builder.WithContext(key, value)
				}

				stepResult := &types.StepResult{
					Name:   fmt.Sprintf("%s (iteration %d)", step.Name, i+1),
					Action: step.Action,
					Result: builder.WithContext("condition", condition).
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

		stepResult, err := executor.stepExecutor.ExecuteStepWithContext(step, stepNum, loopCtx)
		stepResult.Name = fmt.Sprintf("%s (iteration %d)", step.Name, i+1)
		results = append(results, *stepResult)

		if err != nil {
			return results, err
		}
	}

	return results, nil
}
