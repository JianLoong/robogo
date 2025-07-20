package execution

import (
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/templates"
	"github.com/JianLoong/robogo/internal/types"
)

// StepExecutorImpl handles core step execution without retry
type StepExecutorImpl struct {
	variables *common.Variables
}

// NewStepExecutor creates a new step executor
func NewStepExecutor(variables *common.Variables) *StepExecutorImpl {
	return &StepExecutorImpl{
		variables: variables,
	}
}

// ExecuteSingleStep executes a step once without retry
func (executor *StepExecutorImpl) ExecuteSingleStep(
	step types.Step,
	stepNum int,
	loopCtx *types.LoopContext,
) (*types.StepResult, error) {
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

	// Error context enrichment removed for simplicity

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
func (executor *StepExecutorImpl) printStepExecution(
	step types.Step,
	stepNum int,
	args []any,
	options map[string]any,
) {
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

	fmt.Println("  Executing... ")
}

// printStepResult prints the result of step execution
func (executor *StepExecutorImpl) printStepResult(result types.ActionResult, duration time.Duration) {
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
		// Error context removed for simplicity
		if false { // Disabled complex context printing
			// Context printing removed for simplicity
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
