package internal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/templates"
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
					WithTemplate(templates.GetTemplateConstant(constants.TemplateIfConditionFailed)).
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
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidRangeFormat)).
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
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidStartValue)).
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
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidEndValue)).
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
				WithTemplate(templates.GetTemplateConstant(constants.TemplateInvalidCountFormat)).
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

		stepResult, err := executor.executeStepWithContext(step, stepNum, loopCtx)
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
	whileCondition := executor.variables.Substitute(step.While)

	for iterations < maxIterations {
		iterations++
		executor.variables.Set(constants.LoopVariableIteration, iterations)

		// Create loop context for this iteration
		loopCtx := types.NewWhileLoopContext(iterations, whileCondition, maxIterations)

		// Evaluate condition
		condition := executor.variables.Substitute(step.While)
		shouldContinue, err := executor.conditionEvaluator.Evaluate(condition)
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
				Name:   step.Name,
				Action: step.Action,
				Result: builder.WithContext("condition", condition).
					WithContext("error", err.Error()).
					Build(err),
			}
			return append(results, *stepResult), err
		}

		if !shouldContinue {
			break
		}

		stepResult, err := executor.executeStepWithContext(step, stepNum, loopCtx)
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
	return executor.executeStepWithContext(step, stepNum, nil)
}

// executeStepWithContext executes a step with optional loop context and retry support
func (executor *ControlFlowExecutor) executeStepWithContext(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	// If retry is configured, use retry logic
	if step.Retry != nil {
		return executor.executeStepWithRetry(step, stepNum, loopCtx)
	}
	
	// Otherwise, execute normally
	return executor.executeSingleStep(step, stepNum, loopCtx)
}

// executeSingleStep executes a step once without retry
func (executor *ControlFlowExecutor) executeSingleStep(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	start := time.Now()

	result := &types.StepResult{
		Name:   step.Name,
		Action: step.Action,
		Result: types.ActionResult{Status: constants.ActionStatusError},
	}

	// Create step context
	stepContext := types.NewStepContext(stepNum, step.Name, step.Action)
	if loopCtx != nil {
		stepContext.WithLoopContext(loopCtx)
	}

	// Get action
	action, exists := actions.GetAction(step.Action)
	if !exists {
		// Enrich error with step context
		builder := types.NewErrorBuilder(types.ErrorCategoryValidation, "UNKNOWN_ACTION").
			WithTemplate(templates.GetTemplateConstant(constants.TemplateUnknownAction))

		// Add step context to error
		for key, value := range stepContext.ToMap() {
			builder.WithContext(key, value)
		}

		result.Result = builder.Build(step.Action)
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

	// Enrich step context with execution details
	stepContext.WithArguments(convertToInterfaceSlice(args)).
		WithOptions(convertToInterfaceMap(options)).
		WithVariables(executor.variables.GetSnapshot())

	// Add condition context if present
	if step.If != "" || step.For != "" || step.While != "" || step.Result != "" {
		conditionCtx := types.NewConditionContext(step.If, step.For, step.While, step.Result)
		stepContext.WithConditions(conditionCtx)
	}

	// Print step execution details
	executor.printStepExecution(step, stepNum, args, options)

	// Execute action
	output := action(args, options, executor.variables)
	result.Duration = time.Since(start)

	// Enrich error with step context if it's an error
	if output.Status == constants.ActionStatusError && output.ErrorInfo != nil {
		// Add step context to the existing error
		for key, value := range stepContext.ToMap() {
			if _, exists := output.ErrorInfo.Context[key]; !exists {
				output.ErrorInfo.Context[key] = value
			}
		}
	}

	// Use the ActionResult as is
	result.Result = output

	// Print execution result
	executor.printStepResult(output, result.Duration)

	// Return error only if the action status is error (for control flow purposes)
	if output.Status == constants.ActionStatusError {
		return result, fmt.Errorf("action failed: %s", output.GetErrorMessage())
	}

	// Store result variable if specified
	if step.Result != "" {
		executor.variables.Set(step.Result, output.Data)
	}

	return result, nil
}

// printStepExecution prints step execution details to console
func (executor *ControlFlowExecutor) printStepExecution(step types.Step, stepNum int, args []any, options map[string]any) {
	fmt.Printf("Step %d: %s\n", stepNum, step.Name)
	fmt.Printf("  Action: %s\n", step.Action)

	if len(args) > 0 {
		fmt.Printf("  Args: %v\n", args)
	}

	if len(options) > 0 {
		fmt.Printf("  Options: %v\n", options)
	}

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

	fmt.Print("  Executing... ")
}

// printStepResult prints the result of step execution
func (executor *ControlFlowExecutor) printStepResult(result types.ActionResult, duration time.Duration) {
	// Print status with color-like indicators
	switch result.Status {
	case constants.ActionStatusPassed:
		fmt.Printf("✓ PASSED (%s)\n", duration)
	case constants.ActionStatusFailed:
		fmt.Printf("✗ FAILED (%s)\n", duration)
		if errorMsg := result.GetErrorMessage(); errorMsg != "" {
			fmt.Printf("    Error: %s\n", errorMsg)
		}
	case constants.ActionStatusSkipped:
		fmt.Printf("- SKIPPED (%s)\n", duration)
		if skipReason := result.GetSkipReason(); skipReason != "" {
			fmt.Printf("    Reason: %s\n", skipReason)
		}
	case constants.ActionStatusError:
		fmt.Printf("! ERROR (%s)\n", duration)
		if errorMsg := result.GetErrorMessage(); errorMsg != "" {
			fmt.Printf("    Error: %s\n", errorMsg)
		}
		// Show enhanced context for errors
		if result.ErrorInfo != nil && len(result.ErrorInfo.Context) > 0 {
			fmt.Printf("    Context: ")
			var contextParts []string

			// Step information
			if stepNum, ok := result.ErrorInfo.Context["step_number"]; ok {
				contextParts = append(contextParts, fmt.Sprintf("Step %v", stepNum))
			}
			if stepName, ok := result.ErrorInfo.Context["step_name"]; ok {
				contextParts = append(contextParts, fmt.Sprintf("'%v'", stepName))
			}

			// Loop context
			if loopCtx, ok := result.ErrorInfo.Context["loop_context"]; ok {
				if loopData, ok := loopCtx.(map[string]interface{}); ok {
					if loopType, ok := loopData["type"]; ok && loopType == "for" {
						if iteration, ok := loopData["iteration"]; ok {
							contextParts = append(contextParts, fmt.Sprintf("iteration %v", iteration))
						}
						if item, ok := loopData["item"]; ok {
							contextParts = append(contextParts, fmt.Sprintf("item '%v'", item))
						}
					} else if loopType == "while" {
						if iteration, ok := loopData["iteration"]; ok {
							contextParts = append(contextParts, fmt.Sprintf("while iteration %v", iteration))
						}
					}
				}
			}

			// Action-specific context
			if method, ok := result.ErrorInfo.Context["method"]; ok {
				contextParts = append(contextParts, fmt.Sprintf("Method: %v", method))
			}
			if url, ok := result.ErrorInfo.Context["url"]; ok {
				contextParts = append(contextParts, fmt.Sprintf("URL: %v", url))
			}

			if len(contextParts) > 0 {
				fmt.Printf("%s\n", strings.Join(contextParts, ", "))
			}
		}
	default:
		fmt.Printf("? %s (%s)\n", result.Status, duration)
	}

	// Show result data if present and not too large
	if result.Data != nil {
		dataStr := fmt.Sprintf("%v", result.Data)
		if len(dataStr) <= 100 { // Only show small data to avoid cluttering output
			fmt.Printf("    Data: %s\n", dataStr)
		} else {
			fmt.Printf("    Data: [%d characters]\n", len(dataStr))
		}
	}

	fmt.Println() // Add blank line for readability
}

// Helper functions for type conversion

// convertToInterfaceSlice converts []any to []interface{}
func convertToInterfaceSlice(slice []any) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

// convertToInterfaceMap converts map[string]any to map[string]interface{}
func convertToInterfaceMap(m map[string]any) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// executeStepWithRetry executes a step with retry logic
func (executor *ControlFlowExecutor) executeStepWithRetry(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	var lastResult *types.StepResult
	var lastErr error
	
	// Default configuration
	maxAttempts := step.Retry.Attempts
	if maxAttempts <= 0 {
		maxAttempts = 1 // At least try once
	}
	
	// Parse delay duration
	baseDelay := time.Second
	if step.Retry.Delay != "" {
		if parsedDelay, err := time.ParseDuration(step.Retry.Delay); err == nil {
			baseDelay = parsedDelay
		}
	}
	
	// Default backoff strategy
	backoffStrategy := step.Retry.Backoff
	if backoffStrategy == "" {
		backoffStrategy = "fixed"
	}
	
	// Default stop_on_success
	stopOnSuccess := step.Retry.StopOnSuccess
	if step.Retry.StopOnSuccess == false && step.Retry.Attempts > 0 {
		stopOnSuccess = true // Default to true if not explicitly set
	}
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Calculate delay for this attempt (skip delay on first attempt)
		if attempt > 1 {
			delay := executor.calculateDelay(baseDelay, attempt-1, backoffStrategy)
			fmt.Printf("  [Retry] Waiting %v before attempt %d/%d...\n", delay, attempt, maxAttempts)
			time.Sleep(delay)
		}
		
		// Print retry attempt info
		if maxAttempts > 1 {
			fmt.Printf("  [Retry] Attempt %d/%d\n", attempt, maxAttempts)
		}
		
		// Execute the step
		result, err := executor.executeSingleStep(step, stepNum, loopCtx)
		lastResult = result
		lastErr = err
		
		// Check if we should stop retrying
		if executor.shouldStopRetrying(result, err, step.Retry, stopOnSuccess) {
			if attempt > 1 && result.Result.Status == constants.ActionStatusPassed {
				fmt.Printf("  [Retry] ✓ Succeeded on attempt %d/%d\n", attempt, maxAttempts)
			}
			break
		}
		
		// Log retry reason if not the last attempt
		if attempt < maxAttempts {
			fmt.Printf("  [Retry] Failed: %s\n", executor.getRetryReason(result, err))
		}
	}
	
	return lastResult, lastErr
}

// shouldStopRetrying determines if we should stop retrying based on the result
func (executor *ControlFlowExecutor) shouldStopRetrying(result *types.StepResult, err error, retryConfig *types.RetryConfig, stopOnSuccess bool) bool {
	// If it succeeded and we should stop on success, stop
	if stopOnSuccess && result.Result.Status == constants.ActionStatusPassed {
		return true
	}
	
	// If there are specific retry conditions, check them
	if len(retryConfig.RetryOn) > 0 {
		return !executor.shouldRetryForError(result, err, retryConfig.RetryOn)
	}
	
	// Default: retry on any failure, stop on success
	return result.Result.Status == constants.ActionStatusPassed
}

// shouldRetryForError checks if we should retry based on specific error types
func (executor *ControlFlowExecutor) shouldRetryForError(result *types.StepResult, err error, retryOn []string) bool {
	if result.Result.Status == constants.ActionStatusPassed {
		return false
	}
	
	// Check error categories and codes
	if result.Result.ErrorInfo != nil {
		errorCategory := string(result.Result.ErrorInfo.Category)
		errorCode := result.Result.ErrorInfo.Code
		
		for _, condition := range retryOn {
			switch condition {
			case "assertion_failed":
				if errorCategory == "assertion" {
					return true
				}
			case "http_error":
				if errorCategory == "http" || errorCategory == "request" {
					return true
				}
			case "timeout":
				if errorCode == "TIMEOUT" || errorCode == "REQUEST_TIMEOUT" {
					return true
				}
			case "connection_error":
				if errorCode == "CONNECTION_FAILED" || errorCode == "CONNECTION_REFUSED" {
					return true
				}
			case "all":
				return true
			}
		}
		return false
	}
	
	// If no specific error info, retry for any configured condition that includes "all"
	for _, condition := range retryOn {
		if condition == "all" {
			return true
		}
	}
	
	return false
}

// calculateDelay calculates the delay for a retry attempt based on backoff strategy
func (executor *ControlFlowExecutor) calculateDelay(baseDelay time.Duration, attempt int, strategy string) time.Duration {
	switch strategy {
	case "linear":
		return time.Duration(int64(baseDelay) * int64(attempt+1))
	case "exponential":
		multiplier := math.Pow(2, float64(attempt))
		return time.Duration(float64(baseDelay) * multiplier)
	case "fixed":
		fallthrough
	default:
		return baseDelay
	}
}

// getRetryReason returns a human-readable reason for the retry
func (executor *ControlFlowExecutor) getRetryReason(result *types.StepResult, err error) string {
	if result.Result.ErrorInfo != nil {
		return result.Result.ErrorInfo.Message
	}
	if err != nil {
		return err.Error()
	}
	return "Unknown error"
}
