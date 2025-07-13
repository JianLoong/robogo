package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// StepExecutionService encapsulates core step execution logic
// This replaces the global execution functions with a proper service
// Implements StepExecutor interface
type StepExecutionService struct {
	context          TestExecutionContext
	retryExecutor    *util.RetryExecutor
	recoveryExecutor *util.RecoveryExecutor
	controlFlow      *ControlFlowService
	parallelExecutor *ParallelExecutionService
	preprocessor     *StepPreprocessor
}

// NewStepExecutionService creates a new step execution service
func NewStepExecutionService(ctx TestExecutionContext) StepExecutor {
	service := &StepExecutionService{
		context:          ctx,
		retryExecutor:    util.NewRetryExecutor(nil),    // Use default config
		recoveryExecutor: util.NewRecoveryExecutor(nil), // Use default config
	}
	
	// Initialize dependent services
	service.controlFlow = NewControlFlowService(service)
	service.parallelExecutor = NewParallelExecutionService(service)
	service.preprocessor = NewStepPreprocessor(ctx)
	
	return service
}

// ExecuteStep executes a single step with proper encapsulation
func (ses *StepExecutionService) ExecuteStep(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
	startTime := time.Now()

	result := &parser.StepResult{
		Step:      step,
		Status:    "FAILED",
		Duration:  0,
		Output:    "",
		Error:     "",
		Timestamp: startTime,
	}

	// Check if step should be skipped
	if skipInfo := ses.evaluateSkipCondition(step); skipInfo.ShouldSkip {
		result.Status = "SKIPPED"
		result.Duration = time.Since(startTime)
		result.Error = skipInfo.Reason

		if !silent {
			PrintSkipMessage("Step", step.Name, skipInfo.Reason, false)
		}
		return result, nil
	}

	// Handle control flow statements
	if step.If != nil {
		return ses.controlFlow.ExecuteIf(ctx, step, silent)
	}
	if step.For != nil {
		return ses.controlFlow.ExecuteFor(ctx, step, silent)
	}
	if step.While != nil {
		return ses.controlFlow.ExecuteWhile(ctx, step, silent)
	}

	// Execute step with enhanced error handling and retry logic
	output, err := ses.executeStepWithEnhancedErrorHandling(ctx, step, silent)

	result.Duration = time.Since(startTime)

	if err != nil {
		result.Status = "FAILED"
		result.Error = util.FormatRobogoError(err)
		result.Output = fmt.Sprintf("%v", output)

		// Add error context
		if roboErr := util.GetRobogoError(err); roboErr != nil {
			roboErr.WithStep(step.Name)
			if step.Action != "" {
				roboErr.WithAction(step.Action)
			}
		}
		return result, err
	}

	result.Status = "PASSED"
	result.Output = fmt.Sprintf("%v", output)

	// Handle variable setting instructions
	ses.handleVariableInstructions(output, result)

	// Store result in execution history for validation and dependencies
	if step.Result != "" {
		if err := ses.context.Variables().Set(step.Result, output); err != nil {
			result.Error = fmt.Sprintf("Failed to store result variable %s: %v", step.Result, err)
		}
	}

	return result, nil
}

// ExecuteSteps executes multiple steps sequentially
func (ses *StepExecutionService) ExecuteSteps(ctx context.Context, steps []parser.Step, silent bool) ([]parser.StepResult, error) {
	var results []parser.StepResult
	
	for i, step := range steps {
		// Preprocess step for variable substitution
		processedStep, err := ses.preprocessor.PreprocessStep(step, fmt.Sprintf("step_%d", i))
		if err != nil {
			return results, fmt.Errorf("failed to preprocess step %s: %w", step.Name, err)
		}

		result, err := ses.ExecuteStep(ctx, processedStep, silent)
		if result != nil {
			results = append(results, *result)
		}

		if err != nil {
			return results, err
		}

		// Check for early termination
		if result.Status == "FAILED" {
			return results, fmt.Errorf("step %s failed: %s", step.Name, result.Error)
		}
	}

	return results, nil
}

// ExecuteStepsParallel executes steps in parallel based on configuration
func (ses *StepExecutionService) ExecuteStepsParallel(ctx context.Context, steps []parser.Step, config *parser.ParallelConfig, silent bool) ([]parser.StepResult, error) {
	return ses.parallelExecutor.ExecuteParallel(ctx, steps, config, silent)
}

// executeStepWithEnhancedErrorHandling executes a step with retry and recovery
func (ses *StepExecutionService) executeStepWithEnhancedErrorHandling(ctx context.Context, step parser.Step, silent bool) (interface{}, error) {
	// Execute with basic error handling for now
	// TODO: Implement retry and recovery logic when the util interfaces are available
	return ses.executeStepCore(ctx, step, silent)
}

// executeStepCore executes the core step logic
func (ses *StepExecutionService) executeStepCore(ctx context.Context, step parser.Step, silent bool) (interface{}, error) {
	// Execute the action through the action context
	return ses.context.Actions().Execute(ctx, step.Action, step.Args, step.Options, silent)
}

// executeFallback executes fallback actions for recovery
func (ses *StepExecutionService) executeFallback(ctx context.Context, originalStep parser.Step, silent bool) error {
	// TODO: Implement fallback execution logic
	// This would handle recovery strategies like:
	// - Alternative actions
	// - Cleanup operations
	// - Circuit breaker patterns
	return fmt.Errorf("no fallback configured")
}

// handleVariableInstructions processes variable setting instructions from step output
func (ses *StepExecutionService) handleVariableInstructions(output interface{}, result *parser.StepResult) {
	if outputMap, ok := output.(map[string]interface{}); ok {
		if setVarInstruction, exists := outputMap["__robogo_set_variable"]; exists {
			if setVarMap, ok := setVarInstruction.(map[string]interface{}); ok {
				if varName, nameOk := setVarMap["name"].(string); nameOk {
					if varValue, valueOk := setVarMap["value"]; valueOk {
						if err := ses.context.Variables().Set(varName, varValue); err != nil {
							result.Error = fmt.Sprintf("Failed to set variable %s: %v", varName, err)
						}
					}
				}
			}
		}
	}
}

// evaluateSkipCondition evaluates whether a step should be skipped
func (ses *StepExecutionService) evaluateSkipCondition(step parser.Step) SkipInfo {
	// TODO: Implement comprehensive skip logic
	// This would handle conditions like:
	// - Conditional execution
	// - Environment-based skipping
	// - Feature flags
	return SkipInfo{ShouldSkip: false}
}

