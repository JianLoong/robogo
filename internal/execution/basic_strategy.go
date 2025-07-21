package execution

import (
	"fmt"
	"regexp"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/templates"
	"github.com/JianLoong/robogo/internal/types"
)

// BasicExecutionStrategy handles simple action execution without any control flow
type BasicExecutionStrategy struct {
	variables      *common.Variables
	actionRegistry *actions.ActionRegistry
}

// NewBasicExecutionStrategy creates a new basic execution strategy
func NewBasicExecutionStrategy(variables *common.Variables, actionRegistry *actions.ActionRegistry) *BasicExecutionStrategy {
	return &BasicExecutionStrategy{
		variables:      variables,
		actionRegistry: actionRegistry,
	}
}

// Execute performs basic action execution directly
func (s *BasicExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) (*types.StepResult, error) {
	start := time.Now()

	result := &types.StepResult{
		Name:   step.Name,
		Action: step.Action,
		Result: types.ActionResult{Status: constants.ActionStatusError},
	}

	// Get action from registry
	action, exists := s.actionRegistry.Get(step.Action)
	if !exists {
		errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "UNKNOWN_ACTION").
			WithTemplate(templates.GetTemplateConstant(constants.TemplateUnknownAction)).
			WithContext("action", step.Action).
			WithContext("step", step.Name).
			Build(step.Action)
		
		result.Result = errorResult
		result.Duration = time.Since(start)
		return result, fmt.Errorf("unknown action: %s", step.Action)
	}

	// Substitute variables in arguments
	args := s.variables.SubstituteArgs(step.Args)

	// Substitute variables in options
	options := make(map[string]any)
	for k, v := range step.Options {
		if str, ok := v.(string); ok {
			options[k] = s.variables.Substitute(str)
		} else {
			options[k] = v
		}
	}

	// Print step execution details
	s.printStepExecution(step, stepNum, args, options)

	// Execute action directly
	output := action(args, options, s.variables)
	result.Duration = time.Since(start)
	result.Result = output

	// Print execution result
	s.printStepResult(output, result.Duration)

	// Return error only if the action status is error (for control flow purposes)
	if output.Status == constants.ActionStatusError {
		return result, fmt.Errorf("action failed: %s", output.GetErrorMessage())
	}

	// Apply extraction if specified
	var finalData any = output.Data
	if step.Extract != nil && output.Status == constants.ActionStatusPassed {
		extractedData, err := s.applyExtraction(output.Data, step.Extract)
		if err != nil {
			errorResult := types.NewErrorBuilder(types.ErrorCategoryExecution, "EXTRACTION_FAILED").
				WithTemplate("Failed to extract data: %s").
				WithContext("extraction_type", step.Extract.Type).
				WithContext("extraction_path", step.Extract.Path).
				WithContext("error", err.Error()).
				Build(err)
			result.Result = errorResult
			return result, types.NewExtractionError(err.Error())
		}
		finalData = extractedData
		result.Result.Data = finalData
	}

	// Store result variable if specified
	if step.Result != "" {
		s.variables.Set(step.Result, finalData)
	}

	return result, nil
}

// CanHandle returns true for steps that have an action and no control flow
func (s *BasicExecutionStrategy) CanHandle(step types.Step) bool {
	return step.Action != "" && 
		step.Retry == nil && 
		step.If == "" && 
		step.For == "" && 
		step.While == "" &&
		len(step.Steps) == 0
}

// Priority returns low priority as this is the fallback strategy
func (s *BasicExecutionStrategy) Priority() int {
	return 1
}

// Helper methods for extraction

// applyExtraction applies the specified extraction to the data
func (s *BasicExecutionStrategy) applyExtraction(data any, config *types.ExtractConfig) (any, error) {
	if data == nil {
		return nil, types.NewNilDataError()
	}

	switch config.Type {
	case "jq":
		return s.applyJQExtraction(data, config.Path)
	case "xpath":
		return s.applyXPathExtraction(data, config.Path)
	case "regex":
		return s.applyRegexExtraction(data, config.Path, config.Group)
	default:
		return nil, types.NewUnsupportedExtractionTypeError(config.Type)
	}
}

// applyJQExtraction applies JQ extraction to data
func (s *BasicExecutionStrategy) applyJQExtraction(data any, path string) (any, error) {
	jqAction, exists := s.actionRegistry.Get("jq")
	if !exists {
		return nil, types.NewExtractionError("jq action not available")
	}
	
	result := jqAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetErrorMessage())
	}
	
	return result.Data, nil
}

// applyXPathExtraction applies XPath extraction to data  
func (s *BasicExecutionStrategy) applyXPathExtraction(data any, path string) (any, error) {
	xpathAction, exists := s.actionRegistry.Get("xpath")
	if !exists {
		return nil, types.NewExtractionError("xpath action not available")
	}
	
	result := xpathAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetErrorMessage())
	}
	
	return result.Data, nil
}

// applyRegexExtraction applies regex extraction to data
func (s *BasicExecutionStrategy) applyRegexExtraction(data any, pattern string, group int) (any, error) {
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
		return nil, types.NewInvalidRegexPatternError(pattern, err.Error())
	}
	
	matches := re.FindStringSubmatch(text)
	if matches == nil {
		return nil, types.NewNoRegexMatchError(pattern)
	}
	
	// Default to group 1, or use specified group
	if group == 0 {
		group = 1
	}
	
	if group >= len(matches) {
		return nil, types.NewInvalidCaptureGroupError(group, len(matches)-1)
	}
	
	return matches[group], nil
}

// printStepExecution prints step execution details to console
func (s *BasicExecutionStrategy) printStepExecution(
	step types.Step,
	stepNum int,
	args []any,
	options map[string]any,
) {
	fmt.Printf("Step %d: %s\n", stepNum, step.Name)
	fmt.Printf("  Action: %s\n", step.Action)

	if len(args) > 0 {
		// Mask sensitive information in arguments for database actions
		maskedArgs := s.maskSensitiveArgs(step.Action, args)
		fmt.Printf("  Args: %v\n", maskedArgs)
	}

	if len(options) > 0 {
		fmt.Printf("  Options: %v\n", options)
	}

	// Show conditions if present
	if step.If != "" {
		condition := s.variables.Substitute(step.If)
		fmt.Printf("  If: %s\n", condition)
	}

	if step.For != "" {
		forValue := s.variables.Substitute(step.For)
		fmt.Printf("  For: %s\n", forValue)
	}

	if step.While != "" {
		whileValue := s.variables.Substitute(step.While)
		fmt.Printf("  While: %s\n", whileValue)
	}

	if step.Result != "" {
		fmt.Printf("  Result Variable: %s\n", step.Result)
	}

	fmt.Println("  Executing... ")
}

// printStepResult prints the result of step execution
func (s *BasicExecutionStrategy) printStepResult(result types.ActionResult, duration time.Duration) {
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

// maskSensitiveArgs masks sensitive information in step arguments based on action type
func (s *BasicExecutionStrategy) maskSensitiveArgs(action string, args []any) []any {
	// For database actions, mask connection strings that might contain passwords
	if action == "postgres" || action == "spanner" {
		if len(args) > 0 {
			if connStr, ok := args[0].(string); ok {
				// Mask password in connection string
				maskedConnStr := regexp.MustCompile(`password=([^;]+)`).ReplaceAllString(connStr, "password=***")
				maskedArgs := make([]any, len(args))
				maskedArgs[0] = maskedConnStr
				copy(maskedArgs[1:], args[1:])
				return maskedArgs
			}
		}
	}
	
	return args
}