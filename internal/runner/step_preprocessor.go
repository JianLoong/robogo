package runner

import (
	"strings"

	"github.com/JianLoong/robogo/internal/parser"
)

// StepPreprocessor handles step preprocessing including variable substitution
type StepPreprocessor struct {
	context TestExecutionContext
}

// NewStepPreprocessor creates a new step preprocessor
func NewStepPreprocessor(ctx TestExecutionContext) *StepPreprocessor {
	return &StepPreprocessor{
		context: ctx,
	}
}

// PreprocessStep preprocesses a step for variable substitution
func (sp *StepPreprocessor) PreprocessStep(step parser.Step, contextStr string) (parser.Step, error) {
	// Create a copy of the step to avoid modifying the original
	processedStep := step
	
	// Substitute variables in step name
	if step.Name != "" {
		processedStep.Name = sp.context.Variables().Substitute(step.Name)
	}
	
	// Substitute variables in arguments
	processedArgs := make([]interface{}, len(step.Args))
	for i, arg := range step.Args {
		processedArgs[i] = sp.substituteValueRecursive(arg)
	}
	processedStep.Args = processedArgs
	
	// Substitute variables in options
	if step.Options != nil {
		processedOptions := make(map[string]interface{})
		for key, value := range step.Options {
			processedOptions[key] = sp.substituteValueRecursive(value)
		}
		processedStep.Options = processedOptions
	}
	
	// Substitute variables in conditional statements
	if step.If != nil {
		processedIf := *step.If
		processedIf.Condition = sp.context.Variables().Substitute(step.If.Condition)
		processedStep.If = &processedIf
	}
	
	if step.For != nil {
		processedFor := *step.For
		processedFor.Condition = sp.context.Variables().Substitute(step.For.Condition)
		processedStep.For = &processedFor
	}
	
	if step.While != nil {
		processedWhile := *step.While
		processedWhile.Condition = sp.context.Variables().Substitute(step.While.Condition)
		processedStep.While = &processedWhile
	}
	
	return processedStep, nil
}

// substituteValueRecursive performs recursive variable substitution
func (sp *StepPreprocessor) substituteValueRecursive(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return sp.context.Variables().Substitute(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = sp.substituteValueRecursive(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = sp.substituteValueRecursive(val)
		}
		return result
	default:
		return value
	}
}

// extractVariableReferences extracts variable references from a string
func extractVariableReferences(text string) []string {
	var variables []string

	// Simple extraction of ${variable} patterns
	start := 0
	for {
		startIdx := strings.Index(text[start:], "${")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(text[startIdx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		varName := text[startIdx+2 : endIdx]
		if varName != "" {
			// Handle dot notation - only add the base variable name
			if dotIdx := strings.Index(varName, "."); dotIdx != -1 {
				varName = varName[:dotIdx]
			}
			variables = append(variables, varName)
		}

		start = endIdx + 1
	}

	return variables
}