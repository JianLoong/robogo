package runner

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/your-org/robogo/internal/actions"
	"github.com/your-org/robogo/internal/parser"
)

// TestRunner runs test cases
type TestRunner struct {
	variables map[string]interface{} // Store variables for the test case
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		variables: make(map[string]interface{}),
	}
}

// RunTestFile runs a test case from a file
func RunTestFile(filename string, silent bool) (*parser.TestResult, error) {
	// Parse the test case
	testCase, err := parser.ParseTestFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test file: %w", err)
	}

	// Run the test case
	return RunTestCase(testCase, silent)
}

// RunTestCase runs a single test case
func RunTestCase(testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	runner := NewTestRunner()
	return runner.runTestCase(testCase, silent)
}

// runTestCase runs a single test case
func (tr *TestRunner) runTestCase(testCase *parser.TestCase, silent bool) (*parser.TestResult, error) {
	startTime := time.Now()

	// Create action executor
	executor := actions.NewActionExecutor()

	// Initialize result
	result := &parser.TestResult{
		TestCase:    *testCase,
		Status:      "PASSED",
		TotalSteps:  len(testCase.Steps),
		StepResults: make([]parser.StepResult, 0, len(testCase.Steps)),
	}

	if !silent {
		fmt.Printf("🚀 Running test case: %s\n", testCase.Name)
		if testCase.Description != "" {
			fmt.Printf("📋 Description: %s\n", testCase.Description)
		}
		fmt.Printf("📝 Steps: %d\n\n", len(testCase.Steps))
	}

	// Execute each step
	for i, step := range testCase.Steps {
		stepStart := time.Now()
		stepLabel := step.Name
		if stepLabel == "" {
			stepLabel = step.Action
		}
		if !silent {
			fmt.Printf("Step %d: %s\n", i+1, stepLabel)
		}

		// Substitute variables in arguments
		substitutedArgs := tr.substituteVariables(step.Args)

		// Execute the action
		output, err := executor.Execute(step.Action, substitutedArgs)
		stepDuration := time.Since(stepStart)
		stepResult := parser.StepResult{
			Step:      step,
			Status:    "PASSED",
			Duration:  stepDuration,
			Output:    output,
			Timestamp: time.Now(),
		}

		if err != nil {
			stepResult.Status = "FAILED"
			stepResult.Error = err.Error()
			result.Status = "FAILED"
			result.FailedSteps++
			result.ErrorMessage = err.Error()

			if !silent {
				fmt.Printf("❌ Step %d failed: %s\n", i+1, err.Error())
			}
		} else {
			result.PassedSteps++
			if !silent {
				fmt.Printf("✅ Step %d completed in %v\n", stepDuration)
			}
		}

		// Store result in variable if specified
		if step.Result != "" {
			tr.variables[step.Result] = output
			if !silent {
				fmt.Printf("💾 Stored result in variable: %s = %s\n", step.Result, output)
			}
		}

		result.StepResults = append(result.StepResults, stepResult)
	}

	result.Duration = time.Since(startTime)

	if !silent {
		fmt.Printf("\n🏁 Test completed in %v\n", result.Duration)
		fmt.Printf("\n📊 Test Results:\n")
		fmt.Printf("✅ Status: %s\n", result.Status)
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)
		fmt.Printf("📝 Steps: %d total, %d passed, %d failed\n", result.TotalSteps, result.PassedSteps, result.FailedSteps)
	}

	return result, nil
}

// substituteVariables replaces ${variable} references with actual values
func (tr *TestRunner) substituteVariables(args []interface{}) []interface{} {
	substituted := make([]interface{}, len(args))

	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			substituted[i] = tr.substituteString(v)
		case []interface{}:
			substituted[i] = tr.substituteVariables(v)
		case map[string]interface{}:
			substituted[i] = tr.substituteMap(v)
		default:
			substituted[i] = arg
		}
	}

	return substituted
}

// substituteString replaces ${variable} references in a string
func (tr *TestRunner) substituteString(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if value, exists := tr.variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}
		return match // Return original if variable not found
	})
}

// substituteMap substitutes variables in map values
func (tr *TestRunner) substituteMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case string:
			result[k] = tr.substituteString(val)
		case []interface{}:
			result[k] = tr.substituteVariables(val)
		case map[string]interface{}:
			result[k] = tr.substituteMap(val)
		default:
			result[k] = v
		}
	}
	return result
}
