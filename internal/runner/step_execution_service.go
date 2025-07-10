package runner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// StepExecutionService encapsulates all step execution logic
// This replaces the global execution functions with a proper service
// Implements StepExecutor interface
type StepExecutionService struct {
	context        ExecutionContext
	retryExecutor  *util.RetryExecutor
	recoveryExecutor *util.RecoveryExecutor
}

// NewStepExecutionService creates a new step execution service
func NewStepExecutionService(ctx ExecutionContext) StepExecutor {
	return &StepExecutionService{
		context:          ctx,
		retryExecutor:    util.NewRetryExecutor(nil), // Use default config
		recoveryExecutor: util.NewRecoveryExecutor(nil), // Use default config
	}
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
		return ses.executeIfStatement(ctx, step, silent)
	}
	if step.For != nil {
		return ses.executeForLoop(ctx, step, silent)
	}
	if step.While != nil {
		return ses.executeWhileLoop(ctx, step, silent)
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
	
	// Store result in variables if specified
	if step.Result != "" {
		if err := ses.context.Variables().Set(step.Result, output); err != nil {
			result.Error = fmt.Sprintf("Failed to store result: %v", err)
		}
	}
	
	return result, nil
}

// ExecuteSteps executes multiple steps with proper dependency management
func (ses *StepExecutionService) ExecuteSteps(ctx context.Context, steps []parser.Step, silent bool) ([]parser.StepResult, error) {
	results := make([]parser.StepResult, 0, len(steps))
	
	for i, step := range steps {
		// Set step context for variable substitution
		stepContext := fmt.Sprintf("step_%d", i+1)
		
		// Substitute variables in step
		processedStep, err := ses.preprocessStep(step, stepContext)
		if err != nil {
			return results, fmt.Errorf("failed to preprocess step %d: %w", i+1, err)
		}
		
		// Execute the step
		result, err := ses.ExecuteStep(ctx, processedStep, silent)
		if result != nil {
			results = append(results, *result)
		}
		
		// Handle step failure
		if err != nil {
			if !processedStep.ContinueOnFailure {
				return results, fmt.Errorf("step %d failed: %w", i+1, err)
			}
			// Continue with next step if continue_on_failure is true
		}
		
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}
	}
	
	return results, nil
}

// ExecuteStepsParallel executes steps in parallel with dependency management
func (ses *StepExecutionService) ExecuteStepsParallel(ctx context.Context, steps []parser.Step, config *parser.ParallelConfig, silent bool) ([]parser.StepResult, error) {
	if config == nil {
		return ses.ExecuteSteps(ctx, steps, silent)
	}
	
	// Analyze step dependencies
	stepGroups := ses.analyzeStepDependencies(steps, config)
	
	allResults := make([]parser.StepResult, 0, len(steps))
	
	if !silent {
		fmt.Printf("Running %d step groups in parallel\n", len(stepGroups))
	}
	
	// Execute step groups in sequence, steps within groups in parallel
	for groupIdx, group := range stepGroups {
		if !silent {
			fmt.Printf("Running %d steps in parallel for group %d\n", len(group), groupIdx)
		}
		
		groupResults, err := ses.executeStepGroupParallel(ctx, group, config.MaxConcurrency, silent)
		if err != nil {
			return allResults, fmt.Errorf("failed to execute step group %d: %w", groupIdx+1, err)
		}
		
		allResults = append(allResults, groupResults...)
	}
	
	return allResults, nil
}

// Private helper methods

func (ses *StepExecutionService) evaluateSkipCondition(step parser.Step) SkipInfo {
	if step.Skip == nil {
		return SkipInfo{ShouldSkip: false}
	}
	
	// Use the same logic as the existing skip evaluation
	switch v := step.Skip.(type) {
	case bool:
		if v {
			return SkipInfo{ShouldSkip: true, Reason: "skip condition is true"}
		}
		return SkipInfo{ShouldSkip: false}
	case string:
		// Substitute variables in the skip condition
		substituted := ses.context.Variables().Substitute(v)
		if substituted != "" && substituted != "false" && substituted != "0" {
			return SkipInfo{ShouldSkip: true, Reason: substituted}
		}
		return SkipInfo{ShouldSkip: false}
	default:
		// Convert to string and evaluate
		strValue := fmt.Sprintf("%v", v)
		if strValue != "" && strValue != "false" && strValue != "0" {
			return SkipInfo{ShouldSkip: true, Reason: strValue}
		}
		return SkipInfo{ShouldSkip: false}
	}
}

// executeStepWithEnhancedErrorHandling executes a step with comprehensive error handling, retry, and recovery
func (ses *StepExecutionService) executeStepWithEnhancedErrorHandling(ctx context.Context, step parser.Step, silent bool) (interface{}, error) {
	// Create error context with step information
	errorContext := util.NewErrorContext().
		AddBreadcrumb(fmt.Sprintf("Executing step: %s", step.Name)).
		SetVariable("step_name", step.Name).
		SetVariable("step_action", step.Action)
	
	// Check if step has custom retry configuration
	retryConfig := parser.GetStepRetryConfig(step)
	if retryConfig != nil {
		// Create custom retry executor for this step
		customRetryExecutor := util.NewRetryExecutor(retryConfig)
		
		// Execute with custom retry logic
		return customRetryExecutor.ExecuteWithRetryTyped(ctx, func() (interface{}, error) {
			return ses.executeStepCore(ctx, step, silent, errorContext)
		})
	}
	
	// Check if step has recovery configuration
	recoveryConfig := parser.GetStepRecoveryConfig(step)
	if recoveryConfig != nil {
		// Create custom recovery executor for this step
		customRecoveryExecutor := util.NewRecoveryExecutor(recoveryConfig)
		
		// Execute with recovery logic
		var result interface{}
		err := customRecoveryExecutor.ExecuteWithRecovery(ctx, func() error {
			var execErr error
			result, execErr = ses.executeStepCore(ctx, step, silent, errorContext)
			return execErr
		}, func() error {
			// Fallback function
			return ses.executeFallback(ctx, step, recoveryConfig, silent, errorContext)
		})
		
		return result, err
	}
	
	// Default execution with standard retry
	return ses.retryExecutor.ExecuteWithRetryTyped(ctx, func() (interface{}, error) {
		return ses.executeStepCore(ctx, step, silent, errorContext)
	})
}

// executeStepCore executes the core step logic with enhanced error context
func (ses *StepExecutionService) executeStepCore(ctx context.Context, step parser.Step, silent bool, errorContext *util.ErrorContext) (interface{}, error) {
	// Add execution breadcrumb
	errorContext.AddBreadcrumb(fmt.Sprintf("Executing action: %s", step.Action))
	
	// Execute through the action executor
	result, err := ses.context.Actions().Execute(ctx, step.Action, step.Args, step.Options, silent)
	
	if err != nil {
		// For debugging: preserve original error unless we're adding significant value
		if roboErr := util.GetRobogoError(err); roboErr != nil {
			// Only enhance if we don't already have step context
			if roboErr.Step == "" {
				roboErr.WithStep(step.Name)
			}
			if roboErr.Action == "" {
				roboErr.WithAction(step.Action)
			}
			return result, roboErr
		}
		// For regular errors, don't wrap unless it's truly helpful
		// Just return the original error to keep debugging simple
		return result, err
	}
	
	return result, nil
}

// executeFallback executes a fallback operation
func (ses *StepExecutionService) executeFallback(ctx context.Context, originalStep parser.Step, config *util.RecoveryConfig, silent bool, errorContext *util.ErrorContext) error {
	if config.FallbackAction == "" {
		return nil // No fallback defined
	}
	
	errorContext.AddBreadcrumb(fmt.Sprintf("Executing fallback: %s", config.FallbackAction))
	
	// Create a fallback step
	fallbackStep := parser.Step{
		Name:   fmt.Sprintf("Fallback for %s", originalStep.Name),
		Action: config.FallbackAction,
		Args:   originalStep.Args, // Use same args as original
	}
	
	_, err := ses.executeStepCore(ctx, fallbackStep, silent, errorContext)
	return err
}

func (ses *StepExecutionService) preprocessStep(step parser.Step, contextStr string) (parser.Step, error) {
	// Create a copy of the step to avoid modifying the original
	processedStep := step
	
	// Substitute variables in step name with debugging
	if step.Name != "" {
		if execCtx, ok := ses.context.(*DefaultExecutionContext); ok {
			processedStep.Name = execCtx.SubstituteWithDebug(step.Name)
		} else {
			processedStep.Name = ses.context.Variables().Substitute(step.Name)
		}
	}
	
	// Substitute variables in arguments
	if len(step.Args) > 0 {
		processedArgs := make([]interface{}, len(step.Args))
		for i, arg := range step.Args {
			if argStr, ok := arg.(string); ok {
				if execCtx, ok := ses.context.(*DefaultExecutionContext); ok {
					processedArgs[i] = execCtx.SubstituteWithDebug(argStr)
				} else {
					processedArgs[i] = ses.context.Variables().Substitute(argStr)
				}
			} else {
				processedArgs[i] = arg
			}
		}
		processedStep.Args = processedArgs
	}
	
	// Substitute variables in options
	if len(step.Options) > 0 {
		processedOptions := make(map[string]interface{})
		for key, value := range step.Options {
			if valueStr, ok := value.(string); ok {
				if execCtx, ok := ses.context.(*DefaultExecutionContext); ok {
					processedOptions[key] = execCtx.SubstituteWithDebug(valueStr)
				} else {
					processedOptions[key] = ses.context.Variables().Substitute(valueStr)
				}
			} else {
				processedOptions[key] = value
			}
		}
		processedStep.Options = processedOptions
	}
	
	return processedStep, nil
}

func (ses *StepExecutionService) analyzeStepDependencies(steps []parser.Step, config *parser.ParallelConfig) [][]parser.Step {
	// Simple dependency analysis - can be enhanced with proper dependency graph
	if !config.Steps {
		// Return each step as its own group (sequential execution)
		groups := make([][]parser.Step, len(steps))
		for i, step := range steps {
			groups[i] = []parser.Step{step}
		}
		return groups
	}
	
	// For now, group steps that don't depend on each other
	// This is a simplified implementation - real dependency analysis would be more complex
	var safeSteps []parser.Step
	var unsafeSteps []parser.Step
	
	for _, step := range steps {
		if ses.isStepSafeForParallelExecution(step) {
			safeSteps = append(safeSteps, step)
		} else {
			unsafeSteps = append(unsafeSteps, step)
		}
	}
	
	var groups [][]parser.Step
	
	// Add safe steps as one group
	if len(safeSteps) > 0 {
		groups = append(groups, safeSteps)
	}
	
	// Add unsafe steps individually
	for _, step := range unsafeSteps {
		groups = append(groups, []parser.Step{step})
	}
	
	return groups
}

func (ses *StepExecutionService) isStepSafeForParallelExecution(step parser.Step) bool {
	// Define actions that are safe for parallel execution
	safeActions := []string{
		"log", "sleep", "get_time", "get_random", "length", 
		"http", "postgres", "assert",
	}
	
	for _, safeAction := range safeActions {
		if step.Action == safeAction {
			return true
		}
	}
	
	return false
}

func (ses *StepExecutionService) executeStepGroupParallel(ctx context.Context, steps []parser.Step, maxConcurrency int, silent bool) ([]parser.StepResult, error) {
	if len(steps) == 1 {
		// Single step - execute directly
		result, err := ses.ExecuteStep(ctx, steps[0], silent)
		if result != nil {
			return []parser.StepResult{*result}, err
		}
		return nil, err
	}
	
	// Parallel execution with semaphore for concurrency control
	results := make([]parser.StepResult, len(steps))
	errors := make([]error, len(steps))
	
	semaphore := make(chan struct{}, maxConcurrency)
	done := make(chan struct{})
	
	for i, step := range steps {
		go func(index int, s parser.Step) {
			defer func() { done <- struct{}{} }()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result, err := ses.ExecuteStep(ctx, s, silent)
			if result != nil {
				results[index] = *result
			}
			errors[index] = err
		}(i, step)
	}
	
	// Wait for all steps to complete
	for i := 0; i < len(steps); i++ {
		<-done
	}
	
	// Check for errors
	for i, err := range errors {
		if err != nil && !steps[i].ContinueOnFailure {
			return results, fmt.Errorf("parallel step %d failed: %w", i+1, err)
		}
	}
	
	return results, nil
}

// Control flow execution methods

// executeIfStatement executes an if/else block
func (ses *StepExecutionService) executeIfStatement(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
	startTime := time.Now()
	
	condition := ses.context.Variables().Substitute(step.If.Condition)
	output, err := ses.context.Actions().Execute(ctx, "control", []interface{}{"if", condition}, map[string]interface{}{}, silent)
	if err != nil {
		return &parser.StepResult{
			Step:      step,
			Status:    "FAILED",
			Duration:  time.Since(startTime),
			Error:     fmt.Sprintf("failed to evaluate if condition: %v", err),
			Timestamp: startTime,
		}, err
	}

	// Convert output to boolean for comparison
	var conditionResult bool
	if boolVal, ok := output.(bool); ok {
		conditionResult = boolVal
	} else if stringVal, ok := output.(string); ok {
		conditionResult = stringVal == "true"
	} else {
		conditionResult = false
	}

	var stepsToExecute []parser.Step
	if conditionResult {
		stepsToExecute = step.If.Then
	} else {
		stepsToExecute = step.If.Else
	}

	// Execute the chosen steps
	stepResults, err := ses.ExecuteSteps(ctx, stepsToExecute, silent)
	
	// Determine overall status
	status := "PASSED"
	if err != nil {
		status = "FAILED"
	} else {
		for _, result := range stepResults {
			if result.Status == "FAILED" {
				status = "FAILED"
				break
			}
		}
	}

	return &parser.StepResult{
		Step:      step,
		Status:    status,
		Duration:  time.Since(startTime),
		Output:    fmt.Sprintf("Executed %d steps in if/else block", len(stepResults)),
		Timestamp: startTime,
	}, err
}

// executeForLoop executes a for loop
func (ses *StepExecutionService) executeForLoop(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
	startTime := time.Now()
	
	condition := ses.context.Variables().Substitute(step.For.Condition)
	output, err := ses.context.Actions().Execute(ctx, "control", []interface{}{"for", condition}, map[string]interface{}{}, silent)
	if err != nil {
		return &parser.StepResult{
			Step:      step,
			Status:    "FAILED",
			Duration:  time.Since(startTime),
			Error:     fmt.Sprintf("failed to evaluate for loop condition: %v", err),
			Timestamp: startTime,
		}, err
	}

	var iterations int
	switch v := output.(type) {
	case string:
		iterations, err = strconv.Atoi(v)
		if err != nil {
			return &parser.StepResult{
				Step:      step,
				Status:    "FAILED",
				Duration:  time.Since(startTime),
				Error:     fmt.Sprintf("invalid iteration count: %v", err),
				Timestamp: startTime,
			}, err
		}
	case int:
		iterations = v
	default:
		return &parser.StepResult{
			Step:      step,
			Status:    "FAILED",
			Duration:  time.Since(startTime),
			Error:     "for loop condition must return a number",
			Timestamp: startTime,
		}, fmt.Errorf("invalid for loop condition type")
	}

	// Check max iterations
	maxIterations := step.For.MaxIterations
	if maxIterations > 0 && iterations > maxIterations {
		iterations = maxIterations
	}

	totalSteps := 0
	status := "PASSED"
	
	// Execute loop iterations
	for i := 0; i < iterations; i++ {
		// Set iteration variable
		if err := ses.context.Variables().Set("iteration", i); err != nil {
			status = "FAILED"
			break
		}

		stepResults, err := ses.ExecuteSteps(ctx, step.For.Steps, silent)
		totalSteps += len(stepResults)
		
		if err != nil {
			status = "FAILED"
			break
		}
		
		for _, result := range stepResults {
			if result.Status == "FAILED" {
				status = "FAILED"
				break
			}
		}
		
		if status == "FAILED" {
			break
		}
	}

	return &parser.StepResult{
		Step:      step,
		Status:    status,
		Duration:  time.Since(startTime),
		Output:    fmt.Sprintf("Executed %d iterations with %d total steps", iterations, totalSteps),
		Timestamp: startTime,
	}, nil
}

// executeWhileLoop executes a while loop
func (ses *StepExecutionService) executeWhileLoop(ctx context.Context, step parser.Step, silent bool) (*parser.StepResult, error) {
	startTime := time.Now()
	
	maxIterations := step.While.MaxIterations
	if maxIterations == 0 {
		maxIterations = 100 // Default safety limit
	}

	iterations := 0
	totalSteps := 0
	status := "PASSED"

	for iterations < maxIterations {
		// Evaluate condition
		condition := ses.context.Variables().Substitute(step.While.Condition)
		output, err := ses.context.Actions().Execute(ctx, "control", []interface{}{"while", condition}, map[string]interface{}{}, silent)
		if err != nil {
			status = "FAILED"
			break
		}

		// Convert output to boolean
		var conditionResult bool
		if boolVal, ok := output.(bool); ok {
			conditionResult = boolVal
		} else if stringVal, ok := output.(string); ok {
			conditionResult = stringVal == "true"
		} else {
			conditionResult = false
		}

		if !conditionResult {
			break
		}

		// Set iteration variable
		if err := ses.context.Variables().Set("iteration", iterations); err != nil {
			status = "FAILED"
			break
		}

		// Execute loop steps
		stepResults, err := ses.ExecuteSteps(ctx, step.While.Steps, silent)
		totalSteps += len(stepResults)
		iterations++
		
		if err != nil {
			status = "FAILED"
			break
		}
		
		for _, result := range stepResults {
			if result.Status == "FAILED" {
				status = "FAILED"
				break
			}
		}
		
		if status == "FAILED" {
			break
		}
	}

	if iterations >= maxIterations {
		status = "FAILED"
	}

	return &parser.StepResult{
		Step:      step,
		Status:    status,
		Duration:  time.Since(startTime),
		Output:    fmt.Sprintf("Executed %d iterations with %d total steps", iterations, totalSteps),
		Timestamp: startTime,
	}, nil
}

