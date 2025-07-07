package runner

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
)

// executeSingleStep executes a single test step and returns the result
func executeSingleStep(tr *TestRunner, step parser.Step, executor *actions.ActionExecutor, parentLoop *parser.LoopBlock, silent bool, stepResults *[]parser.StepResult, context string, testCase *parser.TestCase, groupIdx int) (*parser.StepResult, error) {
	stepContext := context
	if parentLoop != nil {
		iteration, _ := tr.variableManager.GetVariable("iteration")
		stepContext = context + fmt.Sprintf("Iteration[%v]: ", iteration)
	}

	// Check for skip at the step level using unified logic
	skipInfo := tr.ShouldSkipStep(step, stepContext)
	if skipInfo.ShouldSkip {
		PrintSkipMessage("Step", step.Name, skipInfo.Reason, silent)
		stepResult := CreateSkipResult(step, skipInfo.Reason)
		return &stepResult, nil
	}

	if step.If != nil {
		if err := executeIfStatement(tr, step.If, executor, silent, stepResults, stepContext+step.Name+"/If: ", testCase); err != nil {
			return nil, err
		}
		return &parser.StepResult{Step: step, Status: "PASSED"}, nil
	}
	if step.For != nil {
		if err := executeForLoop(tr, step.For, executor, silent, stepResults, stepContext+step.Name+"/For: ", testCase); err != nil {
			return nil, err
		}
		return &parser.StepResult{Step: step, Status: "PASSED"}, nil
	}
	if step.While != nil {
		if err := executeWhileLoop(tr, step.While, executor, silent, stepResults, stepContext+step.Name+"/While: ", testCase); err != nil {
			return nil, err
		}
		return &parser.StepResult{Step: step, Status: "PASSED"}, nil
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
	output, err := executeStepWithRetry(tr, step, substitutedArgs, executor, silent)
	stepDuration := time.Since(stepStart)

	// Handle skip action error
	if actions.IsSkipError(err) {
		stepResult := parser.StepResult{
			Step:      step,
			Status:    "SKIPPED",
			Duration:  stepDuration,
			Output:    output,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
		return &stepResult, err // propagate skip error up
	}

	// Get verbosity level for this step
	verbosityLevel := parser.GetVerbosityLevel(&step, testCase)

	// Mask secrets in output for display
	maskedOutput := tr.secretManager.MaskSecretsInString(output)

	// Format verbose output if enabled
	verboseOutput := parser.FormatVerboseOutput(verbosityLevel, step.Action, substitutedArgs, maskedOutput, stepDuration.String())

	stepResult := parser.StepResult{
		Step:      step,
		Status:    "PASSED",
		Duration:  stepDuration,
		Output:    maskedOutput,
		Timestamp: time.Now(),
	}

	// Handle expect_error property
	if step.ExpectError != nil {
		expectErr := validateExpectedError(tr, step.ExpectError, err, output, silent)
		if expectErr != nil {
			stepResult.Status = "FAILED"
			stepResult.Error = expectErr.Error()
			if !silent {
				fmt.Printf("❌ Step %d failed: %s\n", len(*stepResults)+1, expectErr.Error())
			}
		} else {
			if !silent {
				fmt.Printf("✅ Error expectation passed\n")
			}
		}
	} else if err != nil {
		// Normal error handling (no expect_error)
		stepResult.Status = "FAILED"
		stepResult.Error = err.Error()
		if !silent {
			fmt.Printf("❌ Step %d failed: %s\n", len(*stepResults)+1, err.Error())
		}
	} else {
		if !silent {
			// Display verbose output if enabled
			if verboseOutput != "" {
				fmt.Print(verboseOutput)
			} else {
				// For log actions, mask the message before displaying
				if step.Action == "log" && len(step.Args) > 0 {
					message := fmt.Sprintf("%v", substitutedArgs[0])
					maskedMessage := tr.secretManager.MaskSecretsInString(message)
					fmt.Printf("📝 %s\n", maskedMessage)
				}
				// Removed completion message - no need to report "completed in X seconds"
			}
		}
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
			fmt.Printf("💾 Stored result in variable: %s = %s\n", step.Result, maskedValue)
		}
	}

	return &stepResult, nil
}

// executeStepWithRetry executes a step with retry logic if configured
func executeStepWithRetry(tr *TestRunner, step parser.Step, args []interface{}, executor *actions.ActionExecutor, silent bool) (string, error) {
	return tr.retryManager.ExecuteWithRetry(step, args, executor, silent)
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

	actualErrorMsg := actualErr.Error()

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
		result = actualErrorMsg == expectedMessage
	case "starts_with":
		result = strings.HasPrefix(actualErrorMsg, expectedMessage)
	case "ends_with":
		result = strings.HasSuffix(actualErrorMsg, expectedMessage)
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
