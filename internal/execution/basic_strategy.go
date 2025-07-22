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
func (s *BasicExecutionStrategy) Execute(step types.Step, stepNum int, loopCtx *types.LoopContext) *types.StepResult {
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
		return result
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
	
	// Pass security information to actions for security-aware behavior
	if step.NoLog {
		options["__no_log"] = true
	}
	if len(step.SensitiveFields) > 0 {
		// Convert []string to []any for options interface
		sensitiveFieldsAny := make([]any, len(step.SensitiveFields))
		for i, field := range step.SensitiveFields {
			sensitiveFieldsAny[i] = field
		}
		options["sensitive_fields"] = sensitiveFieldsAny
	}

	// Print step execution details (unless no_log is enabled)
	if !step.NoLog {
		// Apply masking using step-level sensitive fields
		maskedArgs := s.getMaskedArgsForPrinting(step.Action, args, step.SensitiveFields)
		s.printStepExecution(step, stepNum, maskedArgs, options)
	} else {
		// For no_log steps, print minimal info without sensitive details
		fmt.Printf("Step %d: %s [no_log enabled]\n", stepNum, step.Name)
		fmt.Printf("  Action: %s\n", step.Action)
		fmt.Println("  Executing... ")
	}

	// Execute action directly
	output := action(args, options, s.variables)
	result.Duration = time.Since(start)
	result.Result = output

	// Print execution result (unless no_log is enabled)
	if !step.NoLog {
		s.printStepResult(output, result.Duration)
	} else {
		// For no_log steps, print only status and duration, no sensitive data
		s.printSecureStepResult(output, result.Duration)
	}

	// Apply extraction if specified and action was successful
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
			return result
		}
		finalData = extractedData
		result.Result.Data = finalData
	}

	// Store result variable if specified and action was successful
	if step.Result != "" && (output.Status == constants.ActionStatusPassed || finalData != nil) {
		s.variables.Set(step.Result, finalData)
	}

	return result
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