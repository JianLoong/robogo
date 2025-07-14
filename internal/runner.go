package internal

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/common"
)

// TestRunner - simple, direct test execution
type TestRunner struct {
	variables *common.Variables
}

func NewTestRunner() *TestRunner {
	return &TestRunner{
		variables: common.NewVariables(),
	}
}

// RunTest - execute a single test file
func (r *TestRunner) RunTest(filename string) (*TestResult, error) {
	// Parse test file
	testCase, err := ParseTestFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test file: %w", err)
	}

	// Load test variables
	if testCase.Variables.Vars != nil {
		r.variables.Load(testCase.Variables.Vars)
	}

	start := time.Now()
	result := &TestResult{
		Name:   testCase.Name,
		Status: "PASSED",
		Steps:  make([]StepResult, 0, len(testCase.Steps)),
	}

	fmt.Printf("Running test case: %s\n", testCase.Name)
	if testCase.Description != "" {
		fmt.Printf("Description: %s\n", testCase.Description)
	}
	fmt.Printf("Steps: %d\n\n", len(testCase.Steps))

	// Execute steps
	for i, step := range testCase.Steps {
		stepResult, err := r.executeStep(step, i+1)
		result.Steps = append(result.Steps, *stepResult)

		if err != nil {
			result.Status = "FAILED"
			result.Error = err.Error()
			fmt.Printf("FAILED Step %d: %s\n", i+1, step.Name)
			fmt.Printf("Error: %s\n", err.Error())
			break
		}

		// Store result variable if specified (handled in executeStep)

		fmt.Printf("PASSED Step %d: %s\n", i+1, step.Name)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// RunSuite - execute a test suite
func (r *TestRunner) RunSuite(filename string) error {
	suite, err := ParseTestSuite(filename)
	if err != nil {
		return fmt.Errorf("failed to parse test suite: %w", err)
	}

	fmt.Printf("Running test suite: %s\n", suite.Name)
	fmt.Printf("Test cases: %d\n\n", len(suite.TestCases))

	suiteDir := filepath.Dir(filename)
	passed := 0
	failed := 0

	for i, testCaseFile := range suite.TestCases {
		// Resolve relative paths
		testPath := filepath.Join(suiteDir, testCaseFile)

		fmt.Printf("=== Test Case %d: %s ===\n", i+1, testCaseFile)

		result, err := r.RunTest(testPath)
		if err != nil {
			fmt.Printf("FAILED: %s\n\n", err.Error())
			failed++
		} else if result.Status == "FAILED" {
			fmt.Printf("FAILED: %s\n\n", result.Error)
			failed++
		} else {
			fmt.Printf("PASSED\n\n")
			passed++
		}
	}

	fmt.Printf("=== Suite Summary ===\n")
	fmt.Printf("Total: %d, Passed: %d, Failed: %d\n", len(suite.TestCases), passed, failed)

	if failed > 0 {
		return fmt.Errorf("test suite failed: %d out of %d tests failed", failed, len(suite.TestCases))
	}

	return nil
}

// executeStep - execute a single step
func (r *TestRunner) executeStep(step Step, stepNum int) (*StepResult, error) {
	start := time.Now()

	result := &StepResult{
		Name:   step.Name,
		Action: step.Action,
		Status: "FAILED",
	}

	// Get action
	action, exists := actions.GetAction(step.Action)
	if !exists {
		result.Error = fmt.Sprintf("unknown action: %s", step.Action)
		result.Duration = time.Since(start)
		return result, fmt.Errorf("unknown action: %s", step.Action)
	}

	// Substitute variables in arguments
	args := r.variables.SubstituteArgs(step.Args)

	// Substitute variables in options
	options := make(map[string]interface{})
	for k, v := range step.Options {
		if str, ok := v.(string); ok {
			options[k] = r.variables.Substitute(str)
		} else {
			options[k] = v
		}
	}

	// Execute action
	output, err := action(args, options, r.variables)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Status = "PASSED"
	result.Output = fmt.Sprintf("%v", output)

	// Store result variable if specified
	if step.Result != "" {
		r.variables.Set(step.Result, output) // Store actual output, not string representation
	}

	return result, nil
}

// GetVariables - access to variables for debugging
func (r *TestRunner) GetVariables() *common.Variables {
	return r.variables
}