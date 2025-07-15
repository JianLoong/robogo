package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JianLoong/robogo/internal/common"
)

// TestRunner - simple, direct test execution
type TestRunner struct {
	variables         *common.Variables
	controlFlowExecutor *ControlFlowExecutor
}

func NewTestRunner() *TestRunner {
	variables := common.NewVariables()
	return &TestRunner{
		variables:         variables,
		controlFlowExecutor: NewControlFlowExecutor(variables),
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

	r.printTestHeader(testCase)

	// Execute steps
	for i, step := range testCase.Steps {
		stepResults, stepErr := r.controlFlowExecutor.ExecuteStepWithControlFlow(step, i+1)
		result.Steps = append(result.Steps, stepResults...)

		// Check if any step failed
		if r.handleStepResults(stepResults, stepErr, step.Name, i+1) {
			result.Status = "FAILED"
			result.Error = r.getFirstError(stepResults)
			break
		}
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

// GetVariables - access to variables for debugging
func (r *TestRunner) GetVariables() *common.Variables {
	return r.variables
}

// printTestHeader prints the test case header information
func (r *TestRunner) printTestHeader(testCase *TestCase) {
	fmt.Printf("Running test case: %s\n", testCase.Name)
	if testCase.Description != "" {
		fmt.Printf("Description: %s\n", testCase.Description)
	}
	fmt.Printf("Steps: %d\n\n", len(testCase.Steps))
	os.Stdout.Sync() // Flush output
}

// handleStepResults processes the results from step execution
func (r *TestRunner) handleStepResults(stepResults []StepResult, stepErr error, stepName string, stepNum int) bool {
	// Check if any step failed
	failed := false
	for _, sr := range stepResults {
		if sr.Status == "FAILED" {
			fmt.Printf("FAILED Step %d: %s\n", stepNum, stepName)
			fmt.Printf("Error: %s\n", sr.Error)
			os.Stdout.Sync() // Flush output
			failed = true
			break
		}
	}

	if failed || stepErr != nil {
		return true
	}

	// Print success for executed steps
	for _, sr := range stepResults {
		if sr.Status == "PASSED" {
			fmt.Printf("PASSED Step %d: %s\n", stepNum, stepName)
			os.Stdout.Sync() // Flush output
		} else if sr.Status == "SKIPPED" {
			fmt.Printf("SKIPPED Step %d: %s\n", stepNum, stepName)
			os.Stdout.Sync() // Flush output
		}
	}

	return false
}

// getFirstError extracts the first error from step results
func (r *TestRunner) getFirstError(stepResults []StepResult) string {
	for _, sr := range stepResults {
		if sr.Status == "FAILED" && sr.Error != "" {
			return sr.Error
		}
	}
	return ""
}