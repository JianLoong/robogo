package runner

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/util"
)

// executeSingleStep executes a single test step and returns the result
func executeSingleStep(ctx context.Context, tr *TestRunner, step parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, contextStr string, testCase *parser.TestCase, groupIdx int) (*parser.StepResult, error) {
	stepContext := contextStr
	if parentLoop != nil {
		iteration, _ := tr.variableManager.GetVariable("iteration")
		stepContext = contextStr + fmt.Sprintf("Iteration[%v]: ", iteration)
	}

	// Check for skip at the step level using unified logic
	skipInfo := tr.ShouldSkipStep(step, stepContext)
	if skipInfo.ShouldSkip {
		PrintSkipMessage("Step", step.Name, skipInfo.Reason, silent)
		stepResult := CreateSkipResult(step, skipInfo.Reason)
		return &stepResult, nil
	}

	if step.If != nil {
		if err := executeIfStatement(ctx, tr, step.If, executor, silent, stepResults, stepContext+step.Name+"/If: ", testCase); err != nil {
			return &parser.StepResult{Step: step, Status: parser.StatusFailed, Error: err.Error()}, nil
		}
		return &parser.StepResult{Step: step, Status: parser.StatusPassed}, nil
	}
	if step.For != nil {
		if err := executeForLoop(ctx, tr, step.For, executor, silent, stepResults, stepContext+step.Name+"/For: ", testCase); err != nil {
			return &parser.StepResult{Step: step, Status: parser.StatusFailed, Error: err.Error()}, nil
		}
		return &parser.StepResult{Step: step, Status: parser.StatusPassed}, nil
	}
	if step.While != nil {
		if err := executeWhileLoop(ctx, tr, step.While, executor, silent, stepResults, stepContext+step.Name+"/While: ", testCase); err != nil {
			return &parser.StepResult{Step: step, Status: parser.StatusFailed, Error: err.Error()}, nil
		}
		return &parser.StepResult{Step: step, Status: parser.StatusPassed}, nil
	}

	stepStart := time.Now()
	stepLabel := step.Name
	if stepLabel == "" {
		stepLabel = step.Action
	}
	if !silent {
		fmt.Printf("Step %d: %s\n", len(*stepResults)+1, stepLabel)
	}

	substitutedArgs := tr.substituteVariables(step.Args)
	output, err := executeStepWithRetry(ctx, tr, step, substitutedArgs, executor, silent)
	stepDuration := time.Since(stepStart)

	// Prepare outputStr and maskedOutput after error extraction
	var outputStr string
	if s, ok := output.(string); ok {
		outputStr = s
	} else {
		outputStr = fmt.Sprintf("%v", output)
	}
	maskedOutput := tr.secretManager.MaskSecretsInString(outputStr)

	// Handle skip action error
	if actions.IsSkipError(err) {
		var outputStr string
		if s, ok := output.(string); ok {
			outputStr = s
		} else {
			outputStr = fmt.Sprintf("%v", output)
		}
		stepResult := parser.StepResult{
			Step:      step,
			Status:    parser.StatusSkipped,
			Duration:  stepDuration,
			Output:    outputStr,
			Error:     util.FormatRobogoError(err),
			Timestamp: time.Now(),
		}
		return &stepResult, err // propagate skip error up
	}

	// Get verbosity level for this step
	verbosityLevel := parser.GetVerbosityLevel(&step, testCase)

	// Format verbose output if enabled
	verboseOutput := parser.FormatVerboseOutput(verbosityLevel, step.Action, substitutedArgs, maskedOutput, stepDuration.String())

	// Determine step category based on context
	category := determineStepCategory(stepContext, step)

	// Create display name
	displayName := createDisplayName(step, stepContext, groupIdx)

	stepResult := parser.StepResult{
		Step:          step,
		Status:        parser.StatusPending,
		Duration:      stepDuration,
		Output:        outputStr,
		Timestamp:     time.Now(),
		DisplayName:   displayName,
		Category:      category,
		VerboseOutput: verboseOutput,
		Warnings:      []string{},
		Metadata:      make(map[string]interface{}),
	}

	// Handle expect_error property
	if step.ExpectError != nil {
		expectErr := validateExpectedError(tr, step.ExpectError, err, fmt.Sprintf("%v", output), silent)
		if expectErr != nil {
			stepResult.Status = parser.StatusFailed
			stepResult.Error = util.FormatRobogoError(expectErr)
			if !silent {
				fmt.Printf("Step %d failed: %s\n", len(*stepResults)+1, util.FormatRobogoError(expectErr))
			}
		} else {
			// Error expectation met - step passed
			stepResult.Status = parser.StatusPassed
			if !silent {
				fmt.Printf("Error expectation passed\n")
			}
		}
	} else if err != nil {
		// Normal error handling (no expect_error)
		stepResult.Status = parser.StatusFailed
		stepResult.Error = util.FormatRobogoError(err)
		if !silent {
			fmt.Printf("Step %d failed: %s\n", len(*stepResults)+1, util.FormatRobogoError(err))
		}
	} else {
		// Step succeeded - set status to PASSED
		stepResult.Status = parser.StatusPassed
		if !silent {
			// Display verbose output if enabled
			if verboseOutput != "" {
				fmt.Print(verboseOutput)
			} else {
				// For log actions, mask the message before displaying
				if step.Action == "log" && len(step.Args) > 0 {
					message := fmt.Sprintf("%v", substitutedArgs[0])
					maskedMessage := tr.secretManager.MaskSecretsInString(message)
					fmt.Printf("%s\n", maskedMessage)
				}
				// Removed completion message - no need to report "completed in X seconds"
			}
		}
	}

	// Robust: Always check for 'error' key in output map after all error handling
	if m, ok := output.(map[string]interface{}); ok {
		if errVal, exists := m["error"]; exists {
			if errStr, ok := errVal.(string); ok && errStr != "" {
				stepResult.Error = errStr
			}
		}
	}
	// If error was set from output map, mark step as failed if not already
	if stepResult.Error != "" && stepResult.Status == parser.StatusPassed {
		stepResult.Status = parser.StatusFailed
	}

	// Add context to step name for reporting
	if step.Name != "" {
		stepResult.Step.Name = stepContext + step.Name
	} else if stepContext != "" {
		stepResult.Step.Name = stepContext + fmt.Sprintf("Step%d", groupIdx+1)
	}

	// Store result in variable if specified (store actual value, not masked)
	if step.Result != "" {
		tr.variableManager.SetVariable(step.Result, output) // Store actual output
		if !silent {
			// Display masked version
			maskedValue := tr.secretManager.MaskSecretsInString(fmt.Sprintf("%v", output))
			fmt.Printf("\U0001F4BE Stored result in variable: %s = %s\n", step.Result, maskedValue)
		}
	}

	// Add separator line after step completion
	if !silent {
		fmt.Println(strings.Repeat("-", 80))
	}

	// --- BEGIN: Auto-populate __robogo_steps with step results ---
	// Convert StepResult to map[string]interface{}
	stepMap := map[string]interface{}{
		"name":      stepResult.Step.Name,
		"status":    stepResult.Status,
		"output":    stepResult.Output,
		"error":     stepResult.Error,
		"timestamp": stepResult.Timestamp,
	}
	stepsVar, exists := tr.variableManager.GetVariable("__robogo_steps")
	if !exists {
		// Warn if user has set __robogo_steps manually in test variables
		if _, userSet := tr.variableManager.GetVariable("__robogo_steps"); userSet {
			fmt.Println("[robogo warning] The variable '__robogo_steps' is reserved for internal use and will be overwritten.")
		}
		stepsVar = []interface{}{}
	}
	stepsSlice, ok := stepsVar.([]interface{})
	if !ok {
		stepsSlice = []interface{}{}
	}
	stepsSlice = append(stepsSlice, stepMap)
	tr.variableManager.SetVariable("__robogo_steps", stepsSlice)
	// --- END: Auto-populate __robogo_steps ---

	return &stepResult, nil
}

// executeStepWithRetry executes a step with retry logic if configured
func executeStepWithRetry(ctx context.Context, tr *TestRunner, step parser.Step, args []interface{}, executor *actions.ActionExecutor, silent bool) (interface{}, error) {
	return tr.retryManager.ExecuteWithRetry(ctx, step, args, executor, silent)
}

// validateExpectedError validates that an error occurred as expected
func validateExpectedError(tr *TestRunner, expectError interface{}, actualErr error, output string, silent bool) error {
	var errorType string
	var expectedMessage string

	// Parse expect_error configuration
	switch v := expectError.(type) {
	case string:
		// Simple string format - check if it's "any" or default to "contains"
		if v == "any" {
			errorType = "any"
			expectedMessage = ""
		} else {
			errorType = "contains"
			expectedMessage = v
		}
	case map[string]interface{}:
		// Detailed configuration
		if t, ok := v["type"].(string); ok {
			errorType = t
		} else {
			errorType = "any" // Default to any if type not specified
		}
		if msg, ok := v["message"].(string); ok {
			expectedMessage = msg
		}
	default:
		return fmt.Errorf("invalid expect_error format: must be string or object")
	}

	// If no error occurred but we expected one, that's a failure
	if actualErr == nil {
		if errorType == "any" {
			return fmt.Errorf("expected any error but action succeeded with result: '%s'", output)
		}
		return fmt.Errorf("expected error but action succeeded with result: '%s'", output)
	}

	actualErrorMsg := util.FormatRobogoError(actualErr)

	// Validate the error based on error_type
	var result bool

	switch errorType {
	case "any":
		// Any error is acceptable
		result = true
	case "contains":
		result = strings.Contains(actualErrorMsg, expectedMessage)
	case "not_contains":
		result = !strings.Contains(actualErrorMsg, expectedMessage)
	case "matches":
		// Regex pattern matching
		matched, err := regexp.MatchString(expectedMessage, actualErrorMsg)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %v", expectedMessage, err)
		}
		result = matched
	case "not_matches":
		// Regex pattern not matching
		matched, err := regexp.MatchString(expectedMessage, actualErrorMsg)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %v", expectedMessage, err)
		}
		result = !matched
	case "exact":
		actualStr := fmt.Sprintf("%v", actualErrorMsg)
		expectedStr := fmt.Sprintf("%v", expectedMessage)
		result = actualStr == expectedStr
	case "starts_with":
		actualStr := fmt.Sprintf("%v", actualErrorMsg)
		expectedStr := fmt.Sprintf("%v", expectedMessage)
		result = strings.HasPrefix(actualStr, expectedStr)
	case "ends_with":
		actualStr := fmt.Sprintf("%v", actualErrorMsg)
		expectedStr := fmt.Sprintf("%v", expectedMessage)
		result = strings.HasSuffix(actualStr, expectedStr)
	default:
		return fmt.Errorf("unsupported error type: %s (supported: any, contains, not_contains, matches, not_matches, exact, starts_with, ends_with)", errorType)
	}

	if !result {
		if errorType == "any" {
			return fmt.Errorf("error expectation failed: expected any error but got '%s'", actualErrorMsg)
		}
		return fmt.Errorf("error expectation failed: '%s' %s '%s'", actualErrorMsg, errorType, expectedMessage)
	}

	return nil
}

// determineStepCategory determines the category of a step based on context
func determineStepCategory(context string, step parser.Step) string {
	// Check context first
	if strings.Contains(context, "TDM Setup") {
		return "setup"
	}
	if strings.Contains(context, "TDM Teardown") {
		return "teardown"
	}
	if strings.Contains(context, "Validation") {
		return "validation"
	}

	// Check step action for common patterns
	switch step.Action {
	case "setup", "prepare", "initialize":
		return "setup"
	case "cleanup", "teardown", "close":
		return "teardown"
	case "assert", "validate", "verify":
		return "validation"
	default:
		return "main"
	}
}

// createDisplayName creates a human-readable display name for a step
func createDisplayName(step parser.Step, context string, groupIdx int) string {
	// Use step name if available
	if step.Name != "" {
		return step.Name
	}

	// Create descriptive name based on action and args
	if len(step.Args) > 0 {
		firstArg := fmt.Sprintf("%v", step.Args[0])
		// Truncate long arguments
		if len(firstArg) > 30 {
			firstArg = firstArg[:27] + "..."
		}
		return fmt.Sprintf("%s: %s", step.Action, firstArg)
	}

	// Fallback to action name
	return step.Action
}
