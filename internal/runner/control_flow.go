package runner

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// executeIfStatement executes an if/else block, collecting StepResults
func executeIfStatement(ctx context.Context, tr *TestRunner, ifBlock *parser.ConditionalBlock, executor *actions.ActionExecutor, silent bool, stepResults *[]parser.StepResult, contextStr string, testCase *parser.TestCase) error {
	condition := tr.substituteString(ifBlock.Condition)
	output, err := executor.Execute(ctx, "control", []interface{}{"if", condition}, map[string]interface{}{}, silent)
	if err != nil {
		return fmt.Errorf("failed to evaluate if condition: %w", err)
	}
	var stepsToExecute []parser.Step
	// Convert output to boolean for comparison
	var conditionResult bool
	if boolVal, ok := output.(bool); ok {
		conditionResult = boolVal
	} else if stringVal, ok := output.(string); ok {
		conditionResult = stringVal == "true"
	} else {
		conditionResult = false
	}
	
	if conditionResult {
		stepsToExecute = ifBlock.Then
	} else {
		stepsToExecute = ifBlock.Else
	}
	return executeSteps(ctx, tr, stepsToExecute, executor, nil, silent, stepResults, contextStr, testCase)
}

// executeForLoop executes a for loop, collecting StepResults
func executeForLoop(ctx context.Context, tr *TestRunner, forBlock *parser.LoopBlock, executor *actions.ActionExecutor, silent bool, stepResults *[]parser.StepResult, contextStr string, testCase *parser.TestCase) error {
	condition := tr.substituteString(forBlock.Condition)
	output, err := executor.Execute(ctx, "control", []interface{}{"for", condition}, map[string]interface{}{}, silent)
	if err != nil {
		return fmt.Errorf("failed to evaluate for loop condition: %w", err)
	}
	var iterations int
	switch v := output.(type) {
	case string:
		iterations, err = strconv.Atoi(v)
	case int:
		iterations = v
	default:
		return fmt.Errorf("unexpected output type for for loop iteration count: %T", output)
	}
	if err != nil {
		return fmt.Errorf("failed to parse iteration count: %w", err)
	}
	maxIterations := forBlock.MaxIterations
	if maxIterations > 0 && iterations > maxIterations {
		iterations = maxIterations
	}
	for i := 0; i < iterations; i++ {
		tr.variableManager.SetVariable("iteration", i+1)
		tr.variableManager.SetVariable("index", i)
		if err := executeSteps(ctx, tr, forBlock.Steps, executor, forBlock, silent, stepResults, contextStr, testCase); err != nil {
			return fmt.Errorf("iteration %d failed: %w", i+1, err)
		}
	}
	return nil
}

// executeWhileLoop executes a while loop, collecting StepResults
func executeWhileLoop(ctx context.Context, tr *TestRunner, whileBlock *parser.LoopBlock, executor *actions.ActionExecutor, silent bool, stepResults *[]parser.StepResult, contextStr string, testCase *parser.TestCase) error {
	iteration := 0
	maxIterations := whileBlock.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 1000
	}
	for {
		iteration++
		if iteration > maxIterations {
			return fmt.Errorf("while loop exceeded maximum iterations (%d)", maxIterations)
		}
		tr.variableManager.SetVariable("iteration", iteration)
		condition := tr.substituteString(whileBlock.Condition)
		output, err := executor.Execute(ctx, "control", []interface{}{"while", condition}, map[string]interface{}{}, silent)
		if err != nil {
			return fmt.Errorf("failed to evaluate while condition: %w", err)
		}
		if output != "true" {
			break
		}
		if err := executeSteps(ctx, tr, whileBlock.Steps, executor, whileBlock, silent, stepResults, contextStr, testCase); err != nil {
			return fmt.Errorf("while iteration %d failed: %w", iteration, err)
		}
	}
	return nil
}
