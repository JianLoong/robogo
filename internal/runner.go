package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
)

// TestRunner - simple, direct test execution
type TestRunner struct {
	variables           *common.Variables
	controlFlowExecutor *ControlFlowExecutor
}

func NewTestRunner() *TestRunner {
	variables := common.NewVariables()
	return &TestRunner{
		variables:           variables,
		controlFlowExecutor: NewControlFlowExecutor(variables),
	}
}

// RunTest - execute a single test file
func (r *TestRunner) RunTest(filename string) (*types.TestResult, error) {
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
	result := &types.TestResult{
		Name:   testCase.Name,
		Status: "PASSED",
		Steps:  make([]types.StepResult, 0, len(testCase.Steps)),
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

// GetVariables - access to variables for debugging
func (r *TestRunner) GetVariables() *common.Variables {
	return r.variables
}

// printTestHeader prints the test case header information
func (r *TestRunner) printTestHeader(testCase *types.TestCase) {
	fmt.Printf("Running test case: %s\n", testCase.Name)
	if testCase.Description != "" {
		fmt.Printf("Description: %s\n", testCase.Description)
	}
	fmt.Printf("Steps: %d\n\n", len(testCase.Steps))
	os.Stdout.Sync() // Flush output
}

// handleStepResults processes the results from step execution
func (r *TestRunner) handleStepResults(stepResults []types.StepResult, stepErr error, stepName string, stepNum int) bool {
	// Check if any step failed
	failed := false
	for _, sr := range stepResults {
		if sr.Result.Status == "error" {
			fmt.Printf("FAILED Step %d: %s\n", stepNum, stepName)
			fmt.Printf("Error: %s\n", sr.Result.Error)
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
		if sr.Result.Status == "success" {
			fmt.Printf("PASSED Step %d: %s\n", stepNum, stepName)
			os.Stdout.Sync() // Flush output
		} else if sr.Result.Status == "skipped" {
			fmt.Printf("SKIPPED Step %d: %s\n", stepNum, stepName)
			os.Stdout.Sync() // Flush output
		}
	}

	return false
}

// getFirstError extracts the first error from step results
func (r *TestRunner) getFirstError(stepResults []types.StepResult) string {
	for _, sr := range stepResults {
		if sr.Result.Status == "FAILED" && sr.Result.Error != "" {
			return sr.Result.Error
		}
	}
	return ""
}
