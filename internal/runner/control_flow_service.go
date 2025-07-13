package runner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
)

// ControlFlowService handles control flow operations (if, for, while)
type ControlFlowService struct {
	stepExecutor StepExecutor
}

// NewControlFlowService creates a new control flow service
func NewControlFlowService(stepExecutor StepExecutor) *ControlFlowService {
	return &ControlFlowService{
		stepExecutor: stepExecutor,
	}
}

// ExecuteIf executes an if statement
func (cfs *ControlFlowService) ExecuteIf(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
	startTime := time.Now()
	
	result := &parser.StepResult{
		Step:      step,
		Status:    "PASSED",
		Duration:  0,
		Output:    "",
		Error:     "",
		Timestamp: startTime,
	}

	// Evaluate condition
	condition := step.If.Condition
	conditionResult, err := EvaluateCondition(condition)
	if err != nil {
		result.Status = "FAILED"
		result.Error = fmt.Sprintf("Failed to evaluate if condition '%s': %v", condition, err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	if !silent {
		fmt.Printf("Evaluating if condition: %s = %v\n", condition, conditionResult)
	}

	// Execute steps if condition is true
	if conditionResult {
		stepResults, err := cfs.stepExecutor.ExecuteSteps(ctx, step.If.Then, silent)
		if err != nil {
			result.Status = "FAILED"
			result.Error = fmt.Sprintf("Failed to execute if steps: %v", err)
			result.Duration = time.Since(startTime)
			return result, err
		}

		// Aggregate results
		for _, stepResult := range stepResults {
			if stepResult.Status == "FAILED" {
				result.Status = "FAILED"
				result.Error = stepResult.Error
				break
			}
		}
		result.Output = fmt.Sprintf("Executed %d steps", len(stepResults))
	} else {
		result.Output = "Condition was false, steps skipped"
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// ExecuteFor executes a for loop
func (cfs *ControlFlowService) ExecuteFor(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
	startTime := time.Now()
	
	result := &parser.StepResult{
		Step:      step,
		Status:    "PASSED",
		Duration:  0,
		Output:    "",
		Error:     "",
		Timestamp: startTime,
	}

	// Parse loop parameters
	iterationsStr := step.For.Condition
	iterations, err := strconv.Atoi(iterationsStr)
	if err != nil {
		result.Status = "FAILED"
		result.Error = fmt.Sprintf("Invalid for loop iterations '%s': %v", iterationsStr, err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	if iterations < 0 {
		result.Status = "FAILED"
		result.Error = fmt.Sprintf("For loop iterations must be non-negative, got %d", iterations)
		result.Duration = time.Since(startTime)
		return result, nil
	}

	if !silent {
		fmt.Printf("Starting for loop with %d iterations\n", iterations)
	}

	totalStepsExecuted := 0

	// Execute loop
	for i := 0; i < iterations; i++ {
		if !silent {
			fmt.Printf("For loop iteration %d/%d\n", i+1, iterations)
		}

		stepResults, err := cfs.stepExecutor.ExecuteSteps(ctx, step.For.Steps, silent)
		if err != nil {
			result.Status = "FAILED"
			result.Error = fmt.Sprintf("Failed at iteration %d: %v", i+1, err)
			result.Duration = time.Since(startTime)
			return result, err
		}

		// Check for failures
		for _, stepResult := range stepResults {
			if stepResult.Status == "FAILED" {
				result.Status = "FAILED"
				result.Error = fmt.Sprintf("Step failed at iteration %d: %s", i+1, stepResult.Error)
				result.Duration = time.Since(startTime)
				return result, fmt.Errorf("for loop failed at iteration %d", i+1)
			}
		}

		totalStepsExecuted += len(stepResults)
	}

	result.Output = fmt.Sprintf("Completed %d iterations, executed %d total steps", iterations, totalStepsExecuted)
	result.Duration = time.Since(startTime)
	return result, nil
}

// ExecuteWhile executes a while loop
func (cfs *ControlFlowService) ExecuteWhile(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
	startTime := time.Now()
	
	result := &parser.StepResult{
		Step:      step,
		Status:    "PASSED",
		Duration:  0,
		Output:    "",
		Error:     "",
		Timestamp: startTime,
	}

	condition := step.While.Condition
	iterations := 0
	totalStepsExecuted := 0
	const maxIterations = 1000 // Safety limit

	if !silent {
		fmt.Printf("Starting while loop with condition: %s\n", condition)
	}

	// Execute while loop
	for {
		// Evaluate condition
		conditionResult, err := EvaluateCondition(condition)
		if err != nil {
			result.Status = "FAILED"
			result.Error = fmt.Sprintf("Failed to evaluate while condition '%s': %v", condition, err)
			result.Duration = time.Since(startTime)
			return result, err
		}

		if !conditionResult {
			break
		}

		iterations++
		if iterations > maxIterations {
			result.Status = "FAILED"
			result.Error = fmt.Sprintf("While loop exceeded maximum iterations (%d)", maxIterations)
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("while loop runaway protection triggered")
		}

		if !silent {
			fmt.Printf("While loop iteration %d\n", iterations)
		}

		stepResults, err := cfs.stepExecutor.ExecuteSteps(ctx, step.While.Steps, silent)
		if err != nil {
			result.Status = "FAILED"
			result.Error = fmt.Sprintf("Failed at iteration %d: %v", iterations, err)
			result.Duration = time.Since(startTime)
			return result, err
		}

		// Check for failures
		for _, stepResult := range stepResults {
			if stepResult.Status == "FAILED" {
				result.Status = "FAILED"
				result.Error = fmt.Sprintf("Step failed at iteration %d: %s", iterations, stepResult.Error)
				result.Duration = time.Since(startTime)
				return result, fmt.Errorf("while loop failed at iteration %d", iterations)
			}
		}

		totalStepsExecuted += len(stepResults)
	}

	result.Output = fmt.Sprintf("Completed %d iterations, executed %d total steps", iterations, totalStepsExecuted)
	result.Duration = time.Since(startTime)
	return result, nil
}