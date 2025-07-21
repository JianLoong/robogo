package execution

import (
	"fmt"
	"regexp"
	"strings"
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

	// Apply extraction if specified
	var finalData any = output.Data
	if step.Extract != nil && output.Status == constants.ActionStatusPassed {
		extractedData, err := executor.applyExtraction(output.Data, step.Extract)
		if err != nil {
			// Extraction failed - convert to error result
			errorResult := types.NewErrorBuilder(types.ErrorCategoryExecution, "EXTRACTION_FAILED").
				WithTemplate("Failed to extract data: %s").
				WithContext("extraction_type", step.Extract.Type).
				WithContext("extraction_path", step.Extract.Path).
				WithContext("error", err.Error()).
				Build(err)
			result.Result = errorResult
			return result, fmt.Errorf("extraction failed: %s", err.Error())
		}
		finalData = extractedData
		// Update the result data with extracted value
		result.Result.Data = finalData
	}

	// Store result variable if specified
	if step.Result != "" {
		executor.variables.Set(step.Result, finalData)
	}

	return result, nil
}

// applyExtraction applies the specified extraction to the data
func (executor *StepExecutorImpl) applyExtraction(data any, config *types.ExtractConfig) (any, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot extract from nil data")
	}

	switch config.Type {
	case "jq":
		return executor.applyJQExtraction(data, config.Path)
	case "xpath":
		return executor.applyXPathExtraction(data, config.Path)
	case "regex":
		return executor.applyRegexExtraction(data, config.Path, config.Group)
	default:
		return nil, fmt.Errorf("unsupported extraction type: %s", config.Type)
	}
}

// applyJQExtraction applies JQ extraction to data
func (executor *StepExecutorImpl) applyJQExtraction(data any, path string) (any, error) {
	// Import and use the existing JQ action logic
	// This is simplified - we'd reuse the actual JQ action implementation
	jqAction, exists := actions.GetAction("jq")
	if !exists {
		return nil, fmt.Errorf("jq action not available")
	}
	
	// Execute JQ with the data and path
	result := jqAction([]any{data, path}, map[string]any{}, executor.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, fmt.Errorf("jq extraction failed: %s", result.GetErrorMessage())
	}
	
	return result.Data, nil
}

// applyXPathExtraction applies XPath extraction to data  
func (executor *StepExecutorImpl) applyXPathExtraction(data any, path string) (any, error) {
	xpathAction, exists := actions.GetAction("xpath")
	if !exists {
		return nil, fmt.Errorf("xpath action not available")
	}
	
	result := xpathAction([]any{data, path}, map[string]any{}, executor.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, fmt.Errorf("xpath extraction failed: %s", result.GetErrorMessage())
	}
	
	return result.Data, nil
}

// applyRegexExtraction applies regex extraction to data
func (executor *StepExecutorImpl) applyRegexExtraction(data any, pattern string, group int) (any, error) {
	// Convert data to string
	var text string
	switch v := data.(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		text = fmt.Sprintf("%v", v)
	}
	
	// Apply regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %s", err.Error())
	}
	
	matches := re.FindStringSubmatch(text)
	if matches == nil {
		return nil, fmt.Errorf("no matches found for pattern: %s", pattern)
	}
	
	// Default to group 1, or use specified group
	if group == 0 {
		group = 1
	}
	
	if group >= len(matches) {
		return nil, fmt.Errorf("capture group %d not found (only %d groups)", group, len(matches)-1)
	}
	
	return matches[group], nil
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
		// Mask sensitive information in arguments for database actions
		maskedArgs := maskSensitiveArgs(step.Action, args)
		fmt.Printf("  Args: %v\n", maskedArgs)
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

// maskSensitiveArgs masks sensitive information in step arguments based on action type
func maskSensitiveArgs(action string, args []any) []any {
	if len(args) == 0 {
		return args
	}

	// List of actions that commonly have connection strings as the second argument
	dbActions := map[string]bool{
		"postgres":  true,
		"mysql":     true,
		"sqlite":    true,
		"rabbitmq":  true,
		"spanner":   true,
	}

	maskedArgs := make([]any, len(args))
	copy(maskedArgs, args)

	// For database actions, mask the connection string (typically the second argument)
	if dbActions[strings.ToLower(action)] && len(args) >= 2 {
		if connStr, ok := args[1].(string); ok {
			maskedArgs[1] = common.MaskConnectionString(connStr)
		}
	}

	// For HTTP actions, mask potential auth headers or sensitive data
	if strings.ToLower(action) == "http" {
		// Args structure for HTTP: [method, url, body?, options?]
		for i, arg := range maskedArgs {
			if str, ok := arg.(string); ok {
				// Mask potential auth tokens, API keys, etc.
				maskedArgs[i] = common.MaskSensitiveData(str, common.DefaultSensitiveKeys)
			}
		}
	}

	return maskedArgs
}
